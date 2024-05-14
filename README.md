## Run-CDP

This command line tool can attach to a running chrome debugging session with the CDP websocket URL, and then allows the user to perform CDP requests against any running sessions.

### Warning

This isn't a great tool if you need to work with multiple sessions from one terminal session, instead try working with multiple terminal windows (untested).

## How to use

### Connect

```bash
go run main.go ws://127.0.0.1:1234/devtools/browser/ab05c987-1234-324d-b56a-12c40bbab02c
```

You will be presented with the following if the command was successful:

```bash
Connected to CDP WebSocket. Enter JSON array of CDP commands:
```

### CDP

Now run the CDP commands. You can run them one at a time, in a synchronous group or an asynchronous group:

#### One at a time

Paste the CDP command and press enter e.g.:

```bash
{"id":1,"method":"Browser.getVersion"}
```

You will get a response back from chromium which will be printed to stdout. Note that you may see responses that seem to be unrelated to your request, but they're all from the session that you are connected to.

#### In a synchronous group

You can run multiple CDP commands one after the other. Each request will wait for a response before preceding to the next CDP request e.g.:

```bash
[
    {"id":2,"method":"Target.setAutoAttach","params":{"autoAttach":true,"waitForDebuggerOnStart":true,"flatten":true}},
    {"id":3,"method":"Target.createBrowserContext","params":{"disposeOnDetach":true}},
    {"id":5,"method":"Target.createTarget","params":{"url":"about:blank","browserContextId":""}},
    {"id":8,"sessionId":"","method":"Target.setAutoAttach","params":{"autoAttach":true,"waitForDebuggerOnStart":true,"flatten":true}},
    {"id":9,"sessionId":"","method":"Page.enable"},
    {"id":11,"sessionId":"","method":"Log.enable"},
    {"id":12,"sessionId":"","method":"Runtime.enable"},
    {"id":14,"sessionId":"","method":"Page.setLifecycleEventsEnabled","params":{"enabled":true}},
    {"id":15,"sessionId":"","method":"Page.createIsolatedWorld","params":{"frameId":"","worldName":"__k6_browser_utility_world__","grantUniveralAccess":true}},
    {"id":16,"sessionId":"","method":"Page.addScriptToEvaluateOnNewDocument","params":{"source":"//# sourceURL=__xk6_browser_evaluation_script__","worldName":"__k6_browser_utility_world__"}},
    {"id":17,"sessionId":"","method":"Network.enable","params":{}},
    {"id":18,"sessionId":"","method":"Emulation.setDeviceMetricsOverride","params":{"width":1280,"height":720,"deviceScaleFactor":1,"mobile":false,"screenWidth":1280,"screenHeight":720,"screenOrientation":{"type":"landscapePrimary","angle":90}}},
    {"id":19,"sessionId":"","method":"Emulation.setLocaleOverride","params":{"locale":"en-US"}},
    {"id":20,"sessionId":"","method":"Emulation.setEmulatedMedia","params":{"media":"screen","features":[{"name":"prefers-color-scheme","value":"light"},{"name":"prefers-reduced-motion","value":""}]}},
    {"id":21,"sessionId":"","method":"Emulation.setFocusEmulationEnabled","params":{"enabled":true}},
    {"id":22,"sessionId":"","method":"Emulation.setUserAgentOverride","params":{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.40 Safari/537.36","acceptLanguage":"en-US"}},
    {"id":23,"sessionId":"","method":"Runtime.runIfWaitingForDebugger"}
]
```

The response for each will be printed.


#### In an asynchronous group

To run a group asychronously, type `async` and enter, then add the group e.g.:

```bash
async
```

Now the group:

```bash
[{"id":9,"method":"Log.enable","params":{},"sessionId":""},
{"id":10,"method":"Page.setLifecycleEventsEnabled","params":{"enabled":true},"sessionId":""},
{"id":11,"method":"Runtime.enable","params":{},"sessionId":""},
{"id":12,"method":"Page.addScriptToEvaluateOnNewDocument","params":{"source":"","worldName":"__k6_utility_world__"},"sessionId":""},
{"id":13,"method":"Network.enable","sessionId":""},
{"id":14,"method":"Target.setAutoAttach","params":{"autoAttach":true,"waitForDebuggerOnStart":true,"flatten":true},"sessionId":""},
{"id":15,"method":"Emulation.setFocusEmulationEnabled","params":{"enabled":true},"sessionId":""},
{"id":16,"method":"Emulation.setDeviceMetricsOverride","params":{"mobile":false,"width":1280,"height":720,"screenWidth":1280,"screenHeight":720,"deviceScaleFactor":1,"screenOrientation":{"angle":90,"type":"landscapePrimary"}},"sessionId":""},
{"id":17,"method":"Browser.setWindowBounds","params":{"windowId":,"bounds":{"width":1282,"height":800}},"sessionId":""},
{"id":18,"method":"Emulation.setUserAgentOverride","params":{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.40 Safari/537.36","acceptLanguage":"en-US"},"sessionId":""},
{"id":19,"method":"Emulation.setLocaleOverride","params":{"locale":"en-US"},"sessionId":""},
{"id":20,"method":"Emulation.setEmulatedMedia","params":{"media":"","features":[{"name":"prefers-color-scheme","value":"light"},{"name":"prefers-reduced-motion","value":"no-preference"},{"name":"forced-colors","value":"none"}]},"sessionId":""},
{"id":21,"method":"Runtime.runIfWaitingForDebugger","sessionId":""}]
```

To return back to running groups as synchronous, enter `sync` and press enter.

## Running Chrome in debug mode

```bash
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --disable-field-trial-config --disable-background-networking --enable-features=NetworkService,NetworkServiceInProcess \
--disable-background-timer-throttling --disable-backgrounding-occluded-windows --disable-back-forward-cache --disable-breakpad --disable-client-side-phishing-detection \
--disable-component-extensions-with-background-pages --disable-component-update --no-default-browser-check --disable-default-apps --disable-dev-shm-usage --disable-extensions \
--disable-features=ImprovedCookieControls,LazyFrameLoading,GlobalMediaControls,DestroyProfileOnBrowserClose,MediaRouter,DialMediaRouteProvider,AcceptCHFrame,AutoExpandDetailsElement,CertificateTransparencyComponentUpdater,AvoidUnnecessaryBeforeUnloadCheckSync,Translate,HttpsUpgrades,PaintHolding \
--allow-pre-commit-input --disable-hang-monitor --disable-ipc-flooding-protection --disable-popup-blocking --disable-prompt-on-repost --disable-renderer-backgrounding \
--force-color-profile=srgb --metrics-recording-only --no-first-run --enable-automation --password-store=basic --use-mock-keychain --no-service-autorun --export-tagged-pdf \
--disable-search-engine-choice-screen --unsafely-disable-devtools-self-xss-warnings --no-startup-window --user-data-dir=/var/folders/7s/686mfghs7vndw75bv87qwrmh0000gn/T/ --remote-debugging-port=1234 --remote-debugging-address=0.0.0.0
```
