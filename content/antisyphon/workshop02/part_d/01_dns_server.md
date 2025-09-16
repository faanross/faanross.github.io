---
showTableOfContents: true
title: "DNS Server"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson06_Begin).

The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson06_Done).


## Overview
Though some of the details will differ, we'll now do the same essential thing for DNS as we just did for HTTPS. In this lesson we'll create the server, then we'll create the DNS agent (lesson 7), and finally we'll adjust our existing runloop to make it compatible with both HTTPS and DNS (lesson 8).

When we're done with that we'll have all our foundational communication logic in place, which sets us up perfectly for the development of a trigger signal, parsing, and transition logic in our final chapters.


## What We'll Create
- DNS Server (`internals/dns/server_dns.go`)


## Import Library

We'll use another external library for DNS:

```bash
go get github.com/miekg/dns
```


There are a few DNS libraries in Go, but imo this one reigns supreme. It's not only simple and straight-forward to use for cases where you want to keep things high-level (and thus "outsource" a lot of the low level logic to the library), but it allows you near complete control of all aspects of DNS objects and packets.

For example in crafting DNS requests, the library will literally allow you to set every single field of the packet header except for the Z-value.

We won't jump in that deep in this workshop, but I want you to get exposure to this library since in a number of my "more advanced" DNS tools (for [example](https://github.com/faanross/spinnekop), and [here](https://github.com/faanross/dns-packet-analyzer)), as well as other workshops/courses I have planned, having such complete control over DNS packet crafting allows for tremendous opportunities in creating novel and hard-to-detect DNS covert channel communication techniques.





## DNS Server

We'll once again use a struct to represent our DNS server, and create an accompanying constructor to instantiate it.

Let's create a new file `internals/dns/server_dns.go`.

```go
// DNSServer implements the Server interface for DNS
type DNSServer struct {
	addr   string
	server *dns.Server
}

// NewDNSServer creates a new DNS server
func NewDNSServer(cfg *config.Config) *DNSServer {
	return &DNSServer{
		addr: cfg.ServerAddr,
	}
}
```

Here our Server is even simpler since we don't need to reference our cert. Further, the `dns.Server` instance is a type from the `miekg/dns` library we just imported.

Note that, unlike HTTPS, we don't need a struct to represent a message. With DNS we're not sending a JSON in a response body, but rather the value of the IP itself in our DNS response will indicate whether the communication protocol should stay the same or should transition.



## Start()

Let's add our `Start()` method so our `DNSServer` will also satisfy the interface.

```go
// Start implements Server.Start for DNS
func (s *DNSServer) Start() error {
	// Create and configure the DNS server
	s.server = &dns.Server{
		Addr:    s.addr,
		Net:     "udp",
		Handler: dns.HandlerFunc(s.handleDNSRequest), 
	}

	// Start server
	return s.server.ListenAndServe()
}
```


Everything is pretty straightforward, note that we are also defining our handler as a method `handleDNSRequest`, which we can now create in the same file.


## handleDNSRequest

There's quite a bit more to explain here, have a look at the code first then I'll explain it below.

```go

// handleDNSRequest is our DNS Server's handler
func (s *DNSServer) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	// Create response message
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	// Process each question
	for _, question := range r.Question {
		// We only handle A records for now
		if question.Qtype != dns.TypeA {
			continue
		}

		// Log the query
		log.Printf("DNS query for: %s", question.Name)

		// For now, always return 42.42.42.42
		rr := &dns.A{
			Hdr: dns.RR_Header{
				Name:   question.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			A: net.ParseIP("42.42.42.42"),
		}
		m.Answer = append(m.Answer, rr)
	}

	// Send response
	w.WriteMsg(m)
}
```


We'll immediately create `m` by calling `new()` from the `miekg/dns` library to create a dns message. We'll then set two fields so that's it's set a response (since it's the server -> client), and we set it as the authoritative source.


Since there could potentially be multiple questions in the request we'll use a for loop to iterate through all of them. We know of course there are many different types of DNS records, but we can see here that we're ignoring everything that is not an A record.

We'll then create `rr`, which will be rolled into the `Answer` section of our message at the end. We're setting some basic required/expected values, and then crucially we're setting the answer itself as `42.42.42.42`. In our design this means "don't change", so it's functionally equivalent to the `change` field set to `false` in our DNS Server's JSON response body.


## Stop()

To satisfy the interface let's now also add `Stop()`, which will really just call the function from the `miekg/dns` library to stop our server.


```go
// Stop implements Server.Stop for DNS
func (s *DNSServer) Stop() error {
	if s.server == nil {
		return nil
	}
	log.Println("Stopping DNS server...")
	return s.server.Shutdown()
}
```



## Factory function update

Let's update our `NewServer` factory function so that it can call the `NewDNSServer` constructor.


```go
// NewServer creates a new server based on the protocol
func NewServer(cfg *config.Config) (Server, error) {
	switch cfg.Protocol {
	case "https":
		return https.NewHTTPSServer(cfg), nil
	case "dns":
		return dns.NewDNSServer(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %v", cfg.Protocol)
	}
}
```


Our `NewServer` function is now complete.



## Change config to dns

Since we now want to start up a DNS Server let's change the value in our `config.yaml` file.


```yaml
protocol: "dns"
```



## No changes to Server's main

That's it. We don't even need to make any changes to `main` - that's the beauty of an interface-based approach. The specifics are all abstracted away, we'll the only difference of course is that `cfg` now contains `dns` as the desired protocol, so the factory function will now return a DNS Server, and we can of course call `Start()` and `Stop()` on it.






## Test

Let's run our server.

```shell
❯ go run ./cmd/server
2025/08/11 15:11:43 Starting dns server on 127.0.0.1:8443
```


We don't have an agent yet, but we can use something like `dig` to query our server.
```shell
 dig @localhost -p 8443 www.thisdoesnotexist.com

; <<>> DiG 9.10.6 <<>> @localhost -p 8443 www.thisdoesnotexist.com
; (2 servers found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 29164
;; flags: qr aa rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
;; WARNING: recursion requested but not available

;; QUESTION SECTION:
;www.thisdoesnotexist.com.      IN      A

;; ANSWER SECTION:
www.thisdoesnotexist.com. 300   IN      A       42.42.42.42

;; Query time: 0 msec
;; SERVER: 127.0.0.1#8443(127.0.0.1)
;; WHEN: Mon Aug 11 15:12:08 EDT 2025
;; MSG SIZE  rcvd: 82

```

Note our answer section contains the IP `42.42.42.42`. Also note that our server is created in such a way where we can send any domain to it, it's agnostic and will for now always respond with `42.42.42.42`.

We can see our confirmation message in the server's output.

```shell
❯ go run ./cmd/server
2025/08/11 15:11:43 Starting dns server on 127.0.0.1:8443
2025/08/11 15:12:08 DNS query for: www.thisdoesnotexist.com.

```




___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../part_c/03_https_loop.md" >}})
[|NEXT|]({{< ref "02_dns_agent.md" >}})