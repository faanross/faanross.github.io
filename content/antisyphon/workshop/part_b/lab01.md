---
showTableOfContents: true
title: "Basic Listener, Handler, Router (Lab 01)"
type: "page"
---
## overview
In this first lab we'll create a basic listener, router, and handler:
-  a **LISTENER** is capable of binding to a specified port to create a socket. It can then listen for incoming requests and assuming all conditions are met will accept and create a connection.
- a **ROUTER** is what "allows things to be done" with that connection. 
    - It allows us to set up **ROUTES**, which connects a specific endpoint to for example a handler.
    - As a simple example, we can say that if a GET request is sent to `/heartbeat` (the endpoint), execute a function (handler) that will check if any jobs are waiting in the queue.
    - a **HANDLER** is thus the function that executes in response to a specific request method being received by a specific endpoint.

## cmd/server/main.go

Let's create our first file, following the convention of placing all our entry points in the `cmd` folder, 
we'll first create a directory called `cmd` in our root directory, inside of that we'll create a directory called `server`,
and then finally inside of that we'll create a file called `main.go`. This will contain the entrypoint to our server application.

Now in Goland for example this file will automatically be declared to be `package server` at the top, but since this
is intended to be our entrypoint, we want to change this to `package main`. 

Now before we even declare our main function let's define two global constants - the interface + port we want our listener to bind to.

```go
const serverAddr = "127.0.0.1"  
const serverPort = "7777"
```

For now we will only test locally, hence `127.0.0.1` - this can obviously be changed later were we to communicate across networks.
Note that during R&D some people also like using "all interfaces" (0.0.0.0), but in general this is not a great idea as it could
allow external hosts to probe. 

Let's now create our main function, and immediately inside of it we'll create a new router using Chi.

```go
func main() {

	r := chi.NewRouter()

}
```

Note that Go does have an excellent router within its standard `net/http` library, and in general the rule of thumb
is that if there's a choice between a standard and 3rd party library you should opt for the former. The reason being simply that
standard libraries tend to be more stable. But, sometimes 3rd libraries provide features that you don't get with standard libraries,
in which case you might want to use them.

Keep in mind that not all 3rd party libraries are created equal; factors like their development history, how recently they've been 
updated, the size and activity of their userbase, and whether they frequently introduce breaking changes should influence your choice. 
For this project, we'll be using the `chi` router. I've selected `chi` primarily because it offers fine-grained control over middleware 
implementation. As our project evolves, particularly in its role as a C2 framework, we will rely heavily on middleware for 
critical tasks such as authentication, request decoding, data decryption, and payload parsing.

Now if you've written the line above you'll likely immediately see some form of an error, depending on which IDE you are using.
The error will say something along the lines of "unresolved reference", which is just Go's fancy way of saying it can't find the
import. And that's because of course, as I just mentioned, this is not a standar library - meaning we need to import it.

So open your terminal, make sure you are in the root directory of our project, and run the following command: 

```shell
go get -u github.com/go-chi/chi/v5
```

By the way if you now peek inside of `go.mod`, you'll notice `chi` has been added. 

Back in `main.go`, the unresolved reference error should now be resolved. Next we'd like to set up our routes, but we're going
to define our actual routes in a new file, after which we'll come back here.


## internal/router/routes.go

Let's create a new folder in the root directory called `internal`, and inside of that we'll create `router`. We'll create 
two new files here in - `routes.go` and `handlers.go`.

Let's first dig into `routes.go`. Following the `package router` declaration and import statements, let's create our
sole function - `SetupRoutes()`.

```go
func SetupRoutes(r chi.Router) {
	r.Get("/", RootHandler)
}
```

We can see it takes a sole argument `r`, which is a chi router instance. And inside our function we'll create a
single route, which will call the `RootHandler` function whenever a GET request is sent to the root (`/`) endpoint. 

So let's go ahead and create our RootHandler function inside of `handlers.go`.


## internal/router/handlers.go

```go
func RootHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("You hit the endpoint: %s\n", r.URL.Path)

    w.Write([]byte("I'm Mister Derp!"))
}

```

Here we now have two arguments - `w http.ResponseWriter, r *http.Request`. You'll always see these two 
arguments show up together in Go handlers. Think of it like a fundamental conversation: `r` is the incoming message – 
it's the actual request hitting our server, carrying all the details about what the client wants (like the URL they asked for, 
any data they sent, etc.). Then, `w` is our way to talk back – it's the tool we can use to craft and send your response.

We can see inside the handler that when this function is called two things will happen:
- On the server-side: We get a notification that the endpoint was hit + a timestamp.
- On the client-side: We receive a simple message - `I'm Mister Derp! `.


  



## cmd/server/main.go

So now our route and handler set up we're ready to circle back to main and finish our lab. After creating our router
instance, let's now call our function that will set up our route.

```go
func main() {

	r := chi.NewRouter() 
	  
	router.SetupRoutes(r)
}
```
Notice that, since it's part of `package router`, the keyword `router` precedes the function call. 

Next, though it's great for allowing fine-grained control that we seperated the server's interface and port with our 
two package-level declarations at the top, we now actually need to combine the two since the function we'll call to create
 our listener requires them combined as a single argument. We can do this quite easily with `fmt.Sprintf`.


```go
serverAddrPort := fmt.Sprintf("%s:%s", serverAddr, serverPort)
```

And now just before we go and create our listener let's just print to console confirming that's what we're doing. 
This might seem like we're getting our order wrong - should we not first bind to the port and only then print to console?
Actually, there's a good reason for this - we won't be able to print, or in fact do anything after we run our listener.
This is because at the moment we only have a single goroutine (thread). We'll discuss and address in our next section,
but for now you'll just have to trust me that the order is correct. 


```go
log.Printf("Starting HTTP server on %s", serverAddrPort)
```

Note that, in general, I prefer using the `log` package since it adds time-stamps to the output. If however you did 
not want those included you can opt for the `fmt` package. 

We can now finally call the `net/http` library method to run our listener with some basic idiomatic Go error-handling included.
```go
	err := http.ListenAndServe(serverAddrPort, r)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
```
<br>

So the entire `main()` should now look like this
```go
func main() {

	r := chi.NewRouter()

	router.SetupRoutes(r)

	serverAddrPort := fmt.Sprintf("%s:%s", serverAddr, serverPort)

	log.Printf("Starting HTTP server on %s", serverAddrPort)

	err := http.ListenAndServe(serverAddrPort, r)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```

## test
First, let's run our actual server. Now we could use `go build`, which would compile our program, and then we could run 
it (2 steps). However, when in a period of rapid development I prefer using `go run`, which is going to compile, execute,
and then delete the binary once we're done. In other words, it does everything we need in a single command, so for now
that's a save I'm happy to accept.

So in your root project folder run:
```shell
go run ./cmd/server
```

This will look for the entrypoint in that directory, which should run our server, leading to the following output:

![lab01](../img/lab01A.png)

We can see that the server application reports that it is running on our chosen interface and port.

Let's run `lsof` to actually confirm this.

![lab01](../img/lab01B.png)

And we can see that indeed we're listening on the chosen port.

Let's now use `curl` to hit our endpoint and see if we trigger the expected output.

![lab01](../img/lab01C.png)

We can see that we're able to connect, and we get the expected message on the client side.

Further, we can see below that on the server side, we also get our expected message. 

![lab01](../img/lab01D.png)




## conclusion
Great, so that's really the core foundation - listener, router, handler. There is however A LOT of weaknesses in our code here,
we're blocking our main thread, we have no mechanism for graceful shutdown etc. Now unfortunately this was one of those corners
that had to be cut when I distilled the course down to 4 hours. However, I did at the very least just give you some introduction
to what blocking means and how Goroutines can help us out, so let's quickly dip our toes in the next






___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../part_a/setup.md" >}})
[|NEXT|]({{< ref "lab02.md" >}})