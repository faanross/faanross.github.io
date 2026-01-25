---
layout: course01
title: "Lesson 9: Dual Server Startup"
---


## Solutions

- **Starting Code:** [lesson_09_begin](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_09_begin)
- **Completed Code:** [lesson_09_end](https://github.com/faanross/antisyphon_course_c2_golang/tree/master/lesson_09_end)

## Overview

Now before we get to our agent-side implementation, there's actually one more thing we need to address here on the server.

Let's just think about this for a second:

- Our server runs say on HTTPS, and so does our agent
- We signal our intent to switch both to DNS by hitting the `/switch` on port 8080
- Our server then sends `true` to the agent
- (**TODO**) Our agent will interpret this value and reach out to connect to the server over DNS

But there is of course something missing here... We don't have a DNS server to respond. So a transition should also involve creating the new server. Now there are two ways to do this - the right way, and the expedited way.

The right way assumes we only have our HTTP server running (in the example above), then when the trigger is received it creates a DNS Server. Then, once it's confirmed the new connection with the agent has been established over DNS it will kill the HTTPS server.

But, there is a much simpler way - we can just start both servers when our application begins, and keep both open. This of course represents the expedited way, and it's what I'll opt to do in this situation since it'll save us quite a bit of work, and have the exact same outcome.

That being said, it's not great practice and it's not something that scales really well - it's a definite uptick in technical debt, the equivalent of sweeping dust under the rug.

That being the case, I do 100% encourage you to think about how to do it in **"the right way"**. This would be an excellent exercise to perform following the completion of this workshop, and by the end you will have all the fundamental building blocks to figure out how to do this.

OK, with that out of the way, let's just make a simple adjustment to our server's main so that we start not only a single server, but both HTTPS and DNS servers.

## Server's main

In our server's main, we currently create and start a single server based on the config. We want to change that to start both servers simultaneously.

Update `cmd/server/main.go`:

```go
package main

import (
	"log"
	"os"
	"os/signal"

	"c2framework/internals/config"
	"c2framework/internals/control"
	"c2framework/internals/server"
)

func main() {
	// Create server config directly in code
	cfg := &config.ServerConfig{
		Protocol:           "https",
		ListeningInterface: "127.0.0.1",
		ListeningPort:      "8443",
		TlsCert:            "./certs/server.crt",
		TlsKey:             "./certs/server.key",
	}

	// Load our control API
	control.StartControlAPI()

	// Create BOTH servers regardless of config
	log.Printf("Starting both protocol servers on %s:%s", cfg.ListeningInterface, cfg.ListeningPort)

	// Create HTTPS server
	httpsCfg := *cfg
	httpsCfg.Protocol = "https"
	httpsServer, err := server.NewServer(&httpsCfg)
	if err != nil {
		log.Fatalf("Failed to create HTTPS server: %v", err)
	}

	// Create DNS server
	dnsCfg := *cfg
	dnsCfg.Protocol = "dns"
	dnsServer, err := server.NewServer(&dnsCfg)
	if err != nil {
		log.Fatalf("Failed to create DNS server: %v", err)
	}

	// Start HTTPS server in goroutine
	go func() {
		log.Printf("Starting HTTPS server on %s:%s (TCP)", cfg.ListeningInterface, cfg.ListeningPort)
		if err := httpsServer.Start(); err != nil {
			log.Fatalf("HTTPS server error: %v", err)
		}
	}()

	// Start DNS server in goroutine
	go func() {
		log.Printf("Starting DNS server on %s:%s (UDP)", cfg.ListeningInterface, cfg.ListeningPort)
		if err := dnsServer.Start(); err != nil {
			log.Fatalf("DNS server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down both servers...")

	if err := httpsServer.Stop(); err != nil {
		log.Printf("Error HTTPS stopping server: %v", err)
	}

	if err := dnsServer.Stop(); err != nil {
		log.Printf("Error DNS stopping server: %v", err)
	}
}
```

## Test

We can test this real quick by running the server:

```shell
go run ./cmd/server
2025/08/24 09:35:43 Starting Control API on :8080
2025/08/24 09:35:43 Starting both protocol servers on 127.0.0.1:8443
2025/08/24 09:35:43 Starting HTTPS server on 127.0.0.1:8443 (TCP)
2025/08/24 09:35:43 Starting DNS server on 127.0.0.1:8443 (UDP)
```

We can see from the output that both servers were started on the same port 8443 - HTTPS on TCP, and DNS on UDP.

We can also confirm that this is actually case with `lsof`:

```shell
lsof -i :8443
COMMAND   PID     USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
server  65243 faanross    5u  IPv4 0xcdbcc35684b1111c      0t0  UDP localhost:pcsync
server  65243 faanross    6u  IPv4 0x3edc32002434b2af      0t0  TCP localhost:pcsync(LISTEN)
```

## Conclusion

Now we have both servers running simultaneously. This means when our agent transitions from HTTPS to DNS (or vice versa), the target server will already be listening and ready to handle the connection.

In the next lesson, we'll implement the agent-side transition logic!

---

[Previous: Lesson 8 - API Switch Client](/courses/course01/lesson-08) | [Next: Lesson 10 - Protocol Transition](/courses/course01/lesson-10) | [Course Home](/courses/course01)
