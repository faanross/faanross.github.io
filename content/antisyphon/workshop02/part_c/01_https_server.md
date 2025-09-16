---
showTableOfContents: true
title: "HTTPS Server"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson03_Begin).

The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson03_Done).

## Overview
In the next 3 lessons we'll create our complete HTTPS logic, including the server (this lesson), agent (Lesson 4), and run loop (lesson 5). Thereafter we'll do the exact same thing for DNS.

For our HTTPS server: we'll once again use a struct to represent an instance of a HTTPS server. We'll then also create an accompanying constructor to instantiate it - it's of course this exact constructor that we'll then call from the factory function.

Additionally, if we can recall from our first lesson - we created a Server interface (aka "contract"), that had two methods - `Start()` and `Stop()`. So, in addition to our `Server` struct and constructor, we'll need to implement these, as well as a handler so that our server can actually "do something" once our agent connects to it.

That's about it, let's get cracking.


## What We'll Create
- HTTPS server (`internals/https/server_https.go`)
- Server's main entrypoint (`cmd/server/main.go`)


## Import Library

Though Go's standard library (`net/http`) has a server + router, I am a huge fan of Chi. It's an established + well-maintained, and though we won't get to it in this course, it's got incredible, flexible support for middleware implementation.

So let's add it with:

```bash
go get github.com/go-chi/chi/v5
```



## Let's Generate Some Certs

Since we'll be using HTTPS, we'll need some certs. In this case I'll generate some self-signed ones using openssl, of course if you have alternative source or method you prefer - by all means go ahead.

First I'll create a directory called `./certs`, then `cd` into it. Let's then run the following:


```bash
# Generate private key
openssl genrsa -out server.key 2048

# Generate certificate
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 365 \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
```




## Server Design

Each time the agent periodically checks in with the server, its going to respond with a JSON containing a single bool field: `change`. By default, change will be `false`, which means "don't change to DNS".

For now, since we essentially just want to build out all our HTTPS machinery before focusing on the transition logic, it's going to stay `false`. Then eventually we'll integrate the ability for the field to change to `true`, when the HTTPS agent thus receives the JSON containing `change=true,` it will transition to DNS.

So that's just a brief overview, for now let's focus on creating a simple server with a single endpoint which will respond with this JSON every time the endpoint is hit.


## Server struct

Let's create a new file in `internals/https/server_https.go` that'll house all the logic for our HTTPS server.

First we'll create a struct for the server itself


```go
// HTTPSServer implements the Server interface for HTTPS
type HTTPSServer struct {
	addr    string
	server  *http.Server
	tlsCert string
	tlsKey  string
}
```


Note that the `server` field is a custom type from the `net/http` library.


## Response struct

Let's now also create a struct that will represent our response message from the server to the agent.


```go
// HTTPSResponse represents the JSON response for HTTPS
type HTTPSResponse struct {
	Change bool `json:"change"`
}
```

This is going to be marshalled to JSON before going on the wire, notice how, similar to how we needed YAML tags in the previous lesson to allow for struct-YAML conversion, we now need JSON tags to achieve something similar.


## Server constructor

Now we can also add our constructor to the same file:

```go
// NewHTTPSServer creates a new HTTPS server
func NewHTTPSServer(cfg *config.Config) *HTTPSServer {
	return &HTTPSServer{
		addr:    cfg.ServerAddr,
		tlsCert: cfg.TlsCert,
		tlsKey:  cfg.TlsKey,
	}
}
```


Note that we're not yet assigning our `server` field - we'll do that in the actual `Start()` method.


## Start()

Now, in order for our HTTPS Server to satisfy the `Server` interface we need to create the `Start()` and `Stop()` methods for it. Let's first create `Start()`:

```go

// Start implements Server.Start for HTTPS
func (s *HTTPSServer) Start() error {
	// Create Chi router
	r := chi.NewRouter()

	// Define our GET endpoint
	r.Get("/", RootHandler)

	// Create the HTTP server
	s.server = &http.Server{
		Addr:    s.addr,
		Handler: r,
	}

	// Start the server
	return s.server.ListenAndServeTLS(s.tlsCert, s.tlsKey)
}
```

So a few things worth remarking - first, we can see we're using the Chi library here to create a router.

We then define our endpoint, this is slightly arbitrary but for simplicity's sake I've assigned it as a GET method hitting the root (`/`) endpoint. We can see that when it's hit, it'll call the `RootHandler` function. We'll create this soon enough.

We'll now also assign the `server` field of our `HTTPSServer` struct, and since calling the library function `ListenAndServeTLS()` returns an error we can call it as the return value.


## RootHandler()

As we just saw, our endpoint will call the handler function called `RootHandler()`. This is what will send respond with the JSON with the change field, and where later we'll splice in some conditional logic to allow the value to change from `false` to `true` if the required circumstances are met.


```go
func RootHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Endpoint %s has been hit by agent\n", r.URL.Path)

	// Create response with change set to false
	response := HTTPSResponse{
		Change: false,
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}
```

Simple as - we instantiate a `HTTPSResponse` with the `Change` field set to `false`, then we encode (marshall) and send it at the bottom.


## Stop()

In order to satisfy the interface we'll also add a `Stop()` method for our server.



```go

// Stop implements Server.Stop for HTTPS
func (s *HTTPSServer) Stop() error {
	// If there's no server, nothing to stop
	if s.server == nil {
		return nil
	}

	// Give the server 5 seconds to shut down gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

```


We'll create a context with a 5-second limit to ensure we can force a shutdown if the server is hanging for any reason. Note again that we can call the `Shutdown()` function as the return value, since it itself returns an error.


Great, so that's it now for our HTTP Server - we've put all the pieces into place. We can now go ahead and create our Server `main` so we can test it, however before we do that let's first call the constructor we've just created in the factory function.



## Factory function

So back in `./internals/models/factories.go`, in the `NewServer` function in the case for `https` we no longer return an error, instead we go and call our `NewHTTPSServer` constructor.


```go
// NewServer creates a new server based on the protocol
func NewServer(cfg *config.Config) (Server, error) {
	switch cfg.Protocol {
	case "https":
		return https.NewHTTPSServer(cfg), nil
	case "dns":
		return nil, fmt.Errorf("DNS server not yet implemented")  
	default:
		return nil, fmt.Errorf("unsupported protocol: %v", cfg.Protocol)
	}
}
```


So now when we call `NewServer` and our `cfg.Protocol `is set to `https`, it'll return an instantiated `HTTPSServer` struct.




## Server's main

We can now create our Server's main in `./cmd/server/main.go`.

```go
const pathToYAML = "./configs/config.yaml"

func main() {
	// Command line flag for config file path
	configPath := flag.String("config", pathToYAML, "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create server using interface's factory function
	server, err := models.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start the server in own goroutine
	go func() {
		log.Printf("Starting %s server on %s", cfg.Protocol, cfg.ServerAddr)
		if err := server.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down server...")
	if err := server.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

}
```


Our initial code - our flag parsing and loading the configuration - is exactly the same as it is with our agent.

Thereafter we can now call our `models.NewServer()` factory function, which, in the case that our stated protocol is `https`, should return a `HTTPSServer` instance.

Now that we have that we can call `Start()` on it, but to ensure we don't block our main goroutine let's call it in it's own goroutine using the `go` keyword.

We'll then also add some basic signal handling so that we can block our main goroutine using `<-sigChan`. Since `Start()` is called in its own goroutine, if we did not do this (or something similar), our program would immediately exit since the `main` function will complete and thus exit. 

Finally, once we've indicated our intention to exit the program we'll progress to the final code, which will gracefully shut down our server using `Stop()`, before exiting the program altogether.

## Test

First let's run our server:

```shell
❯ go run ./cmd/server
2025/08/21 15:49:59 Starting https server on 127.0.0.1:8443
```

We can now use an application like curl to hit this endpoint. Alternatively you could use your browser - simply go to `https://localhost:8443/`. If you use the browser you'll first see a warning since we're using self-signed certs, you'll need to select proceed.

Once we hit the endpoint we should see our JSON with the single field `change`, and the value of `false`.

```shell
❯ curl -k https://localhost:8443/
{"change":false}
```

Note we need to use `-k` since we are using self-signed certs.

And on the server's side we can see a confirmation that our endpoint has been hit.

```shell
❯ go run ./cmd/server
2025/08/14 08:11:21 Starting https server on 127.0.0.1:8443
2025/08/14 08:11:46 Endpoint / has been hit by agent
```



## Conclusion

Great, so we have our HTTPS Server, now let's create our own agent which will be able to connect to it.








___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../part_b/02_yaml.md" >}})
[|NEXT|]({{< ref "02_https_agent.md" >}})