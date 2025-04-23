---
showTableOfContents: true
title: "Adding Websocket to Server (Lab 06)"
type: "page"
---

## Overview

We want to now implement a protocol over which our server and client could communicate with one another. There
are of course many options, for example HTTP/1.1 would be perfectly suitable. The thing is about using HTTP/1.1,
and thus relying on a classic Request + Response model, is that you need to use something like polling so the
client can get updates from the server. This is not only noisy + inelegant, but also creates a lag between the 
time that you server receives updates from the agent, to when your client receives that info from the server in turn.

This might not be a big deal, but it's completely unnecessary since the technology to solve this issue exists - websockets.

## Websockets Quick Intro

WebSockets provide a communication protocol that establishes a persistent, full-duplex connection between a server and a 
client over a single TCP connection. Unlike the traditional HTTP request-response model which necessitates client polling for 
updates, WebSockets allow the server to proactively push data to the client the moment it becomes available. 
This eliminates the latency and overhead of polling, making WebSockets ideal for real-time applications where 
immediate data transfer from server to client is crucial.

Also note that (as you will soon see), a Websocket connections starts out as a HTTP/1.1 connection. During the HTTP handshake
the client will send a standard HTTP GET request containing an "Upgrade" header. If the server then agrees to this
the connection transitions from HTTP to the persistent WebSocket protocol.

## Implementation
In order for our server and client to connect with one another we need to furnish both halves with the ability to
create and handle websocket connections. In this lab we'll add it to our server in Go, and in the next lesson we'll
add it to our client in JS. 


## Library

For our Golang-implementation we'll use a popular library called Gorilla. 

Let's run the following command in our project's root directory to add the module.

```
go get github.com/gorilla/websocket
```

## internal/websocket/wss.go

Now let's create a new directory and file here - `internal/websocket/wss.go`.

The first thing we'll do is define our port as a package-level variable, create a `WebSocketServer` struct with a 
single field (for now), and create the accompanying constructor to instantiate this struct.


```go
package websocket

var WebSocketPort = 8080

// WebSocketServer represents a simple WebSocket server
type WebSocketServer struct {
	port int
}

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer(port int) *WebSocketServer {
	return &WebSocketServer{
		port: port,
	}
}
```

Next, we want to create our function that will start our Websocket server. But before we do that, we're also going
to declare something known as a global singleton.

Let's look at this code first, then I'll dig into both a bit more.

```go
// Global WebSocket server instance
var GlobalWSServer *WebSocketServer

func StartWebSocketServer() {

	GlobalWSServer = NewWebSocketServer(WebSocketPort)
	
	
	go func() {
		err := GlobalWSServer.Start()
		if err != nil {
			log.Fatalf("WebSocket server error: %v", err)
		}
	}()


	time.Sleep(100 * time.Millisecond)
	
	fmt.Println("WebSocket server is running. You can now connect from the web UI.")

}
```

Our global singleton `var GlobalWSServer *WebSocketServer`, which as you can see we then immediately assign the return
value of our constructor inside our function, is a pattern used to create a single instance of a struct, and make it available globally.
Here this is achieved by the fact that it's declared at package-level (ie not inside a function scope), and since
it's capitalized, it means its exported (public). 

We're doing this since we only want a single instance of a websocket server, and we want to make it accessible 
throughout our application. 

Then inside our function `StartWebSocketServer()`, after we've called our constructor, we'll call another function
called `Start()` (which we'll create soon) in its own goroutine. Notice what seems like an arbitrary sleep, this just
introduces a pause to ensure the new function (`Start()`) has time to execute before our calling function here exits.

Note that this is not great practice for a number of reasons, the correct way to have handled this would have been
with channels, but that would be way more complex, and so we use this "quick hack" since it'll do the job most of the time.

Let's now create that `Start()` method.


```go
// Start begins the WebSocket server
func (s *WebSocketServer) Start() error {
	// Set up HTTP handler for the WebSocket endpoint
	http.HandleFunc("/ws", s.handleWebSocket)

	// Start the server
	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("WebSocket server starting on %s\n", addr)

	// Start the HTTP server (this is a blocking call)
	return http.ListenAndServe(addr, nil)
}
```

You can see we're creating an endpoint at `/ws`, which will call a handler `s.handleWebSocket`, which we will create
next. The other major event is that we're creating a HTTP server at the end by calling `http.ListenAndServe()`. As
you likely recall, a websocket connection starts out as an HTTP connection, which we can then upgrade. This will
take place inside our handler `handleWebsocket`, which you can think of as the client's main "entrypoint"
into our server. 

Also note however before we create the handler, we'll also define `upgrader`, which is a struct from our `gorilla` package 
we are instantiating. And we are defining it with the most lax permissions possible, we are essentially saying - 
always and indiscriminately upgrade to a websocket `connection` when a client requests it.
This is fine in a development environment, but is of course something you'd likely want to address before moving to production.

```go
var upgrader = websocket.Upgrader{
// Allow connections from any origin for development
CheckOrigin: func(r *http.Request) bool {
return true
},
}

// handleWebSocket handles WebSocket connections
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Log the new connection
	fmt.Println("New WebSocket connection established")

	
	// Simple message reading loop
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Log the received message
		log.Printf("Received message: %s", message)

		// Echo the message back to the client
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Printf("Error sending message: %v", err)
			break
		}
	}
}
```


Inside the `handleWebSocket()` function then we immediately call the `Upgrade()` function on the struct we just created.
Then I want you to shift attention to the infinite `for` loop. This is in some ways similar to our `for` loop from our
agent's `runLoop` - we want to enter this state indefinitely. This is of course because as mentioned, our websocket 
connection is persistent, so here we are essentially just entering into a state to receive and respond at any moment.

For now, I've furnished this with a basic `echo` functionality - meaning whatever we receive from the client, we just
send right back. This is just so we have some logic to test whether our websocket connection works, we'll replace this
later with something that makes more sense. 




## cmd/server/main.go

It might seem like everything is now set up, but there is one final, critical thing we need to do - call our `StartWebSocketServer()`
from our server's `main.go` to put the whole wheel in motion.


```go
websocket.StartWebSocketServer()
```





## test
So let's run our server first, and we should see that it's listening on `8080`.

![lab06](../img/lab06a.png)

And now, since we don't yet have a client that connect to the server, I'm going to use `websocat` to do so.

![lab06](../img/lab06c.png)

You can see that we were able to connect, and each time we write something it's echo'ed right back at us.

And on the server side we can confirm that we're receiving this message.

![lab06](../img/lab06b.png)


___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "lab05.md" >}})
[|NEXT|]({{< ref "lab07.md" >}})