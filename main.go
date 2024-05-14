package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <WebSocket URL>")
	}

	rawUrl := os.Args[1]

	u, err := url.Parse(rawUrl)
	if err != nil {
		log.Fatalf("Invalid URL: %v", err)
	}

	if u.Scheme != "ws" && u.Scheme != "wss" {
		log.Fatal("URL must be a WebSocket URL starting with ws:// or wss://")
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	fmt.Println("Connected to CDP WebSocket. Enter JSON array of CDP commands:")

	commandsChan := make(chan []byte)

	// Create a scanner with a larger buffer
	scanner := bufio.NewScanner(os.Stdin)
	const maxCapacity = 512 * 1024 // Increase the buffer size to 512KB
	buffer := make([]byte, maxCapacity)
	scanner.Buffer(buffer, maxCapacity)

	var browserContextId string
	var windowId float64
	var targetId string
	var sessionId string
	var async bool

	inFlightResponses := make(map[int64]chan string, 0)
	var inFlightResponsesMu sync.RWMutex

	var inputString string
	go func() {
		for scanner.Scan() {
			var commands []map[string]interface{}
			input := scanner.Text()

			if strings.Contains(input, `async`) {
				async = true
				log.Println("async set")
				continue
			}

			if strings.Contains(input, `sync`) {
				async = false
				log.Println("sync set")
				continue
			}

			// Assume we're in the middle of a list.
			if !strings.HasPrefix(input, "[") && !strings.HasSuffix(input, "]") && len(inputString) != 0 {
				log.Println("Detected the middle of an array of CDP commands")
				inputString += input
				continue
			}

			// Assume we're starting a list but not have the full list yet.
			if strings.HasPrefix(input, "[") && !strings.HasSuffix(input, "]") {
				log.Println("Detected the start of an array of CDP commands")
				inputString = input
				continue
			}

			// Assume we're at the end of a list.
			if !strings.HasPrefix(input, "[") && strings.HasSuffix(input, "]") {
				log.Println("Detected the end of an array of CDP commands")
				inputString += input
				input = inputString
			}

			// Wrap a single CDP command into an array
			if !strings.HasPrefix(input, "[") && !strings.HasSuffix(input, "]") {
				input = "[" + input + "]"
			}

			err := json.Unmarshal([]byte(input), &commands)
			if err != nil {
				log.Println(input)
				log.Fatalf("Failed to parse JSON input: %v", err)
			}

			performRequest := func(cmd map[string]interface{}) {
				cmdJSON, err := json.Marshal(cmd)
				if err != nil {
					log.Println("error marshaling command:", err)
					return
				}

				id := int64(cmd["id"].(float64))
				c := make(chan string)
				defer close(c)

				inFlightResponsesMu.Lock()
				inFlightResponses[id] = c
				inFlightResponsesMu.Unlock()

				commandsChan <- cmdJSON

				resp := <-c

				fmt.Println("<-", resp)

				inFlightResponsesMu.Lock()
				delete(inFlightResponses, id)
				inFlightResponsesMu.Unlock()
			}

			if async {
				log.Println("running group of CDP commands asynchronously")
				var wg sync.WaitGroup
				for _, command := range commands {
					wg.Add(1)
					go func(cmd map[string]interface{}) {
						defer wg.Done()

						performRequest(cmd)
					}(command)
				}
				wg.Wait()
			} else {
				log.Println("running group of CDP commands synchronously")
				for _, command := range commands {
					performRequest(command)
				}
			}

			inputString = ""
		}
	}()

	// Send commands over WebSocket
	go func() {
		for cmd := range commandsChan {
			var fields map[string]interface{}
			err := json.Unmarshal(cmd, &fields)
			if err != nil {
				log.Fatalf("Failed to parse JSON input before write: %v", err)
			}

			if value, ok := fields["sessionId"]; ok {
				if value == "" {
					fields["sessionId"] = sessionId
					log.Println("replaced sessionId")
				}
			}
			if result, ok := fields["params"]; ok {
				res := result.(map[string]interface{})
				if value, ok := res["windowId"]; ok {
					log.Println("found windowId", value)
					if value == float64(0) {
						res["windowId"] = windowId
						fields["params"] = res
					}
				}
				if value, ok := res["targetId"]; ok {
					if value == "" {
						res["targetId"] = targetId
						fields["params"] = res
					}
				}
				if value, ok := res["frameId"]; ok { // frameId is also targetId
					if value == "" {
						res["frameId"] = targetId
						fields["params"] = res
					}
				}
				if value, ok := res["browserContextId"]; ok {
					if value == "" {
						res["browserContextId"] = browserContextId
						fields["params"] = res
					}
				}
			}

			cmd, err := json.Marshal(fields)
			if err != nil {
				log.Println("error marshaling command before write:", err)
				return
			}

			log.Println(string(cmd))

			err = c.WriteMessage(websocket.TextMessage, cmd)
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}()

	responses := make(chan string)
	// Receive and handle responses
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				close(responses)
				return
			}
			responses <- string(message)
		}
	}()

	// Print responses as they arrive
	go func() {
		for response := range responses {
			var r map[string]interface{}
			if err := json.Unmarshal([]byte(response), &r); err != nil {
				log.Fatalf("Failed to parse JSON response: %v", err)
				continue
			}

			func() {
				if _, ok := r["id"]; !ok {
					fmt.Println("<-", response)
					return
				}

				id := (int64)(r["id"].(float64))

				inFlightResponsesMu.Lock()
				defer inFlightResponsesMu.Unlock()

				v, ok := inFlightResponses[id]
				if !ok {
					fmt.Println("----> ERROR! Failed to find waiting requester with", id)
					fmt.Println("----> ERROR! <-", response)
				}

				v <- response
			}()

			if result, ok := r["result"]; ok {
				res := result.(map[string]interface{})
				if value, ok := res["browserContextId"]; ok {
					browserContextId = value.(string)
					log.Println("browser context id is", browserContextId)
				}
				if value, ok := res["windowId"]; ok {
					windowId = value.(float64)
					log.Println("window id is", windowId)
				}
				if value, ok := res["targetId"]; ok {
					targetId = value.(string)
					log.Println("target id is", targetId)
				}
			} else if result, ok := r["params"]; ok && r["method"] == "Target.attachedToTarget" {
				res := result.(map[string]interface{})
				if value, ok := res["sessionId"]; ok {
					sessionId = value.(string)
					log.Println("session id is", sessionId)
				}
			} else if result, ok := r["error"]; ok {
				log.Println("----> ERROR!", result)
			}
		}
	}()

	// Setup clean shutdown on SIGTERM or SIGINT
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	sig := <-sigs
	fmt.Println("Received", sig, ", shutting down.")

	// Explicitly close the WebSocket connection
	c.Close()
	fmt.Println("WebSocket connection closed.")
}
