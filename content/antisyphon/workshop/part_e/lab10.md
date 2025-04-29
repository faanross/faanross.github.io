---
showTableOfContents: true
title: "Server Receives + Queues Commands (Lab 10)"
type: "page"
---
## Overview
So at this point we receive the serialized object in our server, and our goal is now to prepare it for the agent to retrieve it.

Let's just first discuss conceptually what that will look like. Once we receive the JSON object (in `handleWebsocket`) we need to deserialize it. But deserialize it into what? In Go we'll typically use struct to handle custom data types internally, and so we'll need to decide what that looks like.

Then, once we have it in a struct, we need to the command in some form of a queue. Why is that required? Because of course we don't have a persistent connection between the server and agent, and our server cannot initiate exchange with the client.

So from the moment our server receives the command from our client, some time will inevitable pass before our agent check ins with a HTTP request. So we need somewhere to "store" the command before our agent checks in - hence why we'll create some sort of queue.


## internal/websocket/message.go

As I mentioned, we want to create some custom data type (i.e. struct) to deserialize our command into when we receive it from the client. So let's create a new file called `message.go` in our `websocket` package.

Now, we could go ahead and create some struct specifically for our commands (as they pass through the server from the client to the agent), however, the server will later of course also act as a proxy to return results from the agent to the client. So with this in mind, we can be a bit more forward-thinking and create a single, "universal" struct we can use for both purposes.

```go
package websocket

// MessageType defines the types of messages exchanged
type MessageType string

const (
	CommandMessage MessageType = "command"
	ResponseMessage MessageType = "response"
)

// Message represents the general message structure
type Message struct {
	Type      MessageType `json:"type"`
	Command   string      `json:"command,omitempty"`
	Output    string      `json:"output,omitempty"`
	AgentUUID string      `json:"agentUUID,omitempty"`
}
```

At the top we're just defining the two `type`s of `Message` structs - `command` (client -> agent), and `response` (agent -> client).

Then we define what our struct looks like. Note that, aside from `Type`, our fields include the keyword `omitempty`. This just means that if that field is empty in the struct form, DO NOT create an empty field when serializing it to JSON, rather just don't create a field at all. This helps us employ the "universal" nature of this struct since, depending on whether it's a `command` or `response`, the resulting JSON will only contain the fields relevant to its purpose.

## internal/websockets/wss.go

We can now head back here to our existing file containing all the websocket logic. Once inside, scroll down to the `handleWebSocket` handler, which we established earlier serves as the entrypoint for our client into our server.

Further, if you recall earlier all the main logic happens inside of a `for`-loop, and we previously added some contrived `echo` functionality there just to ensure we were able to test that our server's websocket logic was indeed working. So now, since we'll change the guts completely, simple remove everything that was inside of the `for` loop.

```go
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {

	// ALL PREVIOUS CODE OUTSIDE OF FOR LOOP REMAINS THE EXACT SAME
	
	for {
		// REMOVE ALL OF THIS!!
	}
}
```

Add the following code (again, just sticking to inside of the for loop for now).

```go
	for {
		var msg Message
		
		err := conn.ReadJSON(&msg)
		if err != nil {
			// Connection closed or some error
			break
		}

		log.Printf("New Message received: %+v", msg)
	
		// QUEUE THE COMMAND

	}

```

So first we instantiate a new `Message` struct - the one we just created above - before deserializing the message we received from the client directly into it.

We then print it to console, just so we can ensure we're keeping tabs on the flow of our application in case something breaks.

After this I've just placed a placeholder comment for now. This is where we want to queue the command contained within `msg`, however the function does not yet exist, so let's go and create it before we circle back here.

## internal/websocket/command_handler.go

So we'll create a new file called `command_handler.go` inside of `websocket`. First off, we want to create the struct that will hold our commands.

```go
package websocket


type CommandQueue struct {
	// Queue of commands for any agent
	PendingCommands []string
	mu  sync.Mutex  
}
```

We can see the string slice (`[]string`) which is where we will add our commands to. Note that in Go we call these slices and not arrays. Arrays do exist, but we need to specify their size. So for example if we said `[8]string` we are declaring an array of size 8. The issue with that in this case is the size is fixed, though we can change the values of elements, we cannot remove or add any elements to it.

This confines the use of arrays to instances where you know exactly how many elements you need. But in general, when we want to create a "container" to store an indeterminate amount of values, as is the case here, we use slices.

After that we declare a mutex, I'll explain what this is but a bit further down as I think it will make more sense to do this at the point of its use.

Let's now instantiate our struct.

```go
var AgentCommands = CommandQueue{
	PendingCommands: make([]string, 0),
}
```


Here you can see now we are not using a constructor, but a "struct literal". Think of it like - we are not using a function to instantiate our struct, we are just "literally" declaring an instance called `AgentCommands`.

We're using `make()` to create a string slice, and to start it will have no elements (`0`). As I already explained above, with a slice we can add and subtract elements at run-time.

Note that we don't need to instantiate our mutex field, this is because it is already a valid, unlocked mutex ready for use. This is an example of Go's broader design philosophy where types are designed to have useful zero values whenever possible.

Great so now we have our queue (`PendingCommands`), so let's create a function to add a command to it.

```go
// QueueCommand adds a command to the queue
func (cq *CommandQueue) QueueCommand(command string) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	cq.PendingCommands = append(cq.PendingCommands, command)

	log.Printf("Command queued: %s", command)

}
```

You can see that it's receiver function attached to our `CommandQueue` struct, and it takes a single argument, which is of course the command we want to place in the queue.


At the top we are using our mutex to `Lock()`, and then defer the `Unlock()` - let's unpack this.

First, what is a mutex? Well, it really deserves an entire article of its own, but in short: it's the ability to ensure thread safety. So whenever we work with goroutines and structures that we write to we want to use mutexes to ensure only a single goroutine can access it.

So for example here we want to add a string (`command`) to a byte slice (`PendingCommands`), we lock it using a mutex, because if two goroutines happen to write to the byte slice at the exact same moment, one would overwrite the other and we'd lose a command. So it essentially ensures there is no competition between the two goroutines.

But it seems odd that we lock, and then immediately unlock. Should we not wait right until the end before calling `Unlock()`? Well, that's actually exactly what we're doing with `defer`, which means - please run this command, but only the moment when this function finishes. So it's calling it, but putting a delay on it essentially.

And after we can see we use `append()` to actually add the command to the queue, and we just print this to screen so we can confirm during our test it worked.

So now that that's done we can circle back to `handleWebsocket` to actually call it.


## internal/websocket/wss.go

Picking up where we left off before, let's call `AgentCommands.QueueCommand` and pass not the entire `msg` struct, but the `Command` field.

```go

	// Simple message reading loop
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			// Connection closed or some error
			break
		}

		log.Printf("New Message received: %+v", msg)

		// HERE IS THE NEW QUEUE
			
		if msg.Type == CommandMessage {
			AgentCommands.QueueCommand(msg.Command)
		}

	}

```


That's it, we should now be able to queue our command after receiving it from the client, so let's test it out.

## Test

We can once again start our server, then our client, choose any command and we should see the print statement we added in `QueueCommand()` be called with the `command` string only.


![lab10](../img/lab10.png)





___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "lab09.md" >}})
[|NEXT|]({{< ref "lab11.md" >}})