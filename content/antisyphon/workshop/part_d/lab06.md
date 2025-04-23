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

## overview
- Right now we can connect to our vue application via our browser using the connection mediated by vite
- and of course our agent can connect to our server using the HTTP/1.1 connection we created

- But, there is no way for our server to speak to our web ui


- SO we need to now go and create this, and as we just explained Websockets is an ideal protocol since it's real-time, no need for polling


- We'll now build everything out on our server-side then in the next lab we'll get to our client UI to complete the connection



## Server-Side WebSocket Implementation

We'll need to add the Gorilla WebSocket package:

```
go get github.com/gorilla/websocket
```

## websocket_ui.go

- new file here
  `internal/websocket/wss.go`


- Define our WebSocket Port
- Create a struct representing the WSS instance
- And then a constructor to instantiate

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

- For now we'll only add the port here to our WSS struct, we'll build on it later on




- Now we'll also define a global instance and immediately implement it in our `StartWebSocketServer()` method


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

-  `var GlobalWSServer *WebSocketServer` is implementing what's known as the singleton pattern.
- The singleton pattern ensures you have exactly one instance of a particular object accessible from anywhere in your application.
- In this case, it's making a single WebSocket server available globally.

- In our function then we call the constructor and assign it to the global instance
- Note that we start our WSS in a separate goroutine
- We are of course erroring out on `GlobalWSServer.Start()` since we have not yet defined that method, let's do so now



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


- We set up a route we call `handleWebSocket` so when client (vue js frontend) hits /ws on port 8080, we'll define that shortly
- We print confirmation to console and then we start the actual server


- We can now define our handleWebsocket method, this is really where almost everything will happen
- Two things - upgrade our connection from HTTP to WS, and create our reading loop
- So you will get very familiar with this method, we'll use it throughout as the reading loop the entrypoint for the client UI in our server


```go
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}
	
	defer conn.Close()


	// Log the new connection
	fmt.Println("New WebSocket connection established")

```

- The first thing we do is Upgrade our connection, as I explained before WSS piggybacks off HTTP, so start the initial connection as HTTP, then we upgrade - that's what upgraded.Upgrade does

- We need to add this, let's do so right above `handleWebSocket()`

```go
var upgrader = websocket.Upgrader{
	// Allow connections from any origin for development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
```


- Now that we've upgraded we add the heart of this method, which is an infite for-loop used to read and process incoming requests
- For loop - persistent - we know whats what WSS are., so when we start our connection this is where it is the entire time


```go
	for {
		// Read message
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
```

- For now obvs we don't have our UI-side wss method yet so we can't test it
- But I just want to do something simple for a proof of concept
- So essentially we read any message, and then just echo it back to whoever is sending it

- Connects over HTTP but then upgrades to Websocket, as we explained
- Then the infinite for loop is classic websocket - creates this persistent state, but since all in own goroutine not blocking
- For now all we are going to do is echo the message back to client, this will just test to see if we indeed have ability to communicate 2 ways, we'll still change this entier part what goes on in for loop dramatically, this is really where A LOT of the logic is going to happen in terms of what type of message can we receive and what to do with them, send etc.




**ENTIRE METHOD FOR REFERENCE**

```go
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





## now of course we need to also start it in main.go

- Just call the function

```go
websocket.StartWebSocketServer()
```





## test
- We can run see it is listening


___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "lab05.md" >}})
[|NEXT|]({{< ref "lab07.md" >}})