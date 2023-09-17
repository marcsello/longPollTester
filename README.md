# Long Poll Tester

This is a very simple utility to test the long polling capabilities of your setup.

It works by starting an HTTP server. Any request arriving to this server will be halted indefinitely (or a specified timeout).
When the connection is closed it prints the connection details and how long were the connection alive.

By using these metrics, you can determine if it's possible to use long polling in a given environment.

## Compile/install

The simplest way to use this program is (if you have go installed) is just to use the following command:
```
$ go install github.com/marcsello/longPollTester@latest
```

After that's complete you can use the `longPollTester` command from the commandline:

```
$ longPollTester
2023/09/17 21:59:43 Starting HTTP server on :8080
```

If you don't have go installed, check out the Releases page, I have uploaded a pre-compiled binary version there.

## Usage

Currently two command line options supported. Use -help to get more info:
```
$ longPollTester -help
Usage of longPollTester:
  -bind string
    	Address string to bind the server to (default ":8080")
  -timeout duration
    	Time for the server to return with 204 after, don't set it to never return (for intermediate timeout tests)
```

When you start the server, it binds to `:8080` port by default. It accepts any URL and any method. Any new connections will be printed to the console.
When a client connects a log line will be printed to stdout, and another one when the client disconnects. All requests given a unique (incremental) id, it is printed at the beginning of each log line between square brackets.

```
$ longPollTester -timeout=60s
2023/09/17 21:54:17 Starting HTTP server on :8080
2023/09/17 21:54:30 [1] New GET request from [::1]:40242 to /test
2023/09/17 21:55:00 [1] Connection closed! Reason: context canceled. Was open for 30.000760339s
2023/09/17 21:55:16 [2] New GET request from [::1]:35130 to /test
2023/09/17 21:56:16 [2] Connection closed! Reason: context deadline exceeded. Was open for 1m0.045066943s
```

Possible reasons for a connection to close: 
- `context canceled` = The connection closed abruptly (possible the client gave up waiting)   
- `context deadline exceeded` = The configured timeout (with `-timeout=`) is reached. The server returned HTTP 204 and closed the connection.

**Pro-tip:** Use the url to uniquely identify connections.