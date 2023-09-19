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

Currently, the following command line options supported. (Use -help to list them):
```
$ longPollTester -help
Usage of longPollTester:
  -bind string
        Specifies the server's address to bind to. (default ":8080")
  -keepalive-interval duration
        Sets the frequency to write data for keeping the connection alive. Leave unset to disable keep-alive. Requires enabling write-header-early.
  -keepalive-payload string
        Defines the payload for keep-alive data. (default " ")
  -payload string
        Specifies the payload to be sent in the response (generates errors when status is set to 204).
  -status int
        Sets the HTTP status code to include in the response. (default 200)
  -timeout duration
        Defines the duration for the server to respond and subsequently close the connection. Omit to keep the connection open indefinitely.
  -write-header-early
        Enables writing headers as soon as the client connects (while keeping the connection open).
```

When you start the server, it binds to `:8080` port by default. It accepts any path and any method. Any new connections will be printed to the console.
When a client connects a log line will be printed to stdout, and another one when the client disconnects. All requests given a unique (incremental) identifier, it is printed at the beginning of each log line between square brackets.

```
$ longPollTester -timeout=60s
2023/09/19 23:42:26 Starting HTTP server on :8080
2023/09/19 23:42:42 [1] New GET request from [::1]:34492 to /test1
2023/09/19 23:43:12 [1] Connection closed! Reason: client closed connection (context canceled). Was open for 30.000415375s
2023/09/19 23:43:41 [2] New GET request from [::1]:45362 to /test1
2023/09/19 23:44:41 [2] Connection closed! Reason: server timeout. Was open for 1m0.058346014s
```

Possible reasons for a connection to close: 
- `client closed connection` = The connection closed abruptly (possible the client gave up waiting)   
- `server timeout` = The configured timeout (`-timeout=` parameter) is reached. The server returned a response and closed the connection.
- `write error` = Something went wrong while writing either the keep-alive data or the final response.

**Pro-tip:** Use the url to uniquely identify connections.

### Timeout

This utility offers two distinct operating modes: one with a timeout and the other without.

**Without Timeout:** In this mode, the connection remains open indefinitely. 
In other words, the server will never respond, making it particularly useful for testing scenarios involving intermediate network timeout errors.

**With Timeout:** Enabling the timeout mode instructs the server to return a response after a specified duration and subsequently close the connection. 
This mode is useful for assessing the viability and reliability of server timeouts in specific long-polling configurations.

### Write header early

In the context of long-polling, headers are typically written when an event is received or when the server-side connection times out.
However, there are certain use cases, such as keep-alive (as discussed below), that can benefit from sending headers as soon as the client establishes a connection.

To enable this behavior, you can set the `write-header-early flag`.

Enabling this feature causes the server to immediately transmit the HTTP status and headers to the client while keeping the connection open (with the `Transfer-Encoding: chunked` header). 
Consequently, the client cannot discern the presence of an event solely from the HTTP status and headers; it has to rely on the payload.

**Note:** This won't work if you set the `-status` to 204 (No content). 
The server/client will close the connection immediately as there isn't any content expected.  

### Keepalive

Some clients automatically close a connection if there is no activity within a specific time frame. 
To address this issue, you can enable the "keep-alive" feature. With this feature enabled, the server periodically sends data over the connection to prevent it from being prematurely closed. 
By default, a single whitespace character is sent because it does not mess up JSON parsing.

To enable this behavior, simply set the `-keepalive-interval=` option to specify the desired delay between sending keep-alive data.

**Note**: This setting requires the `write-header-early` option to be enabled.

