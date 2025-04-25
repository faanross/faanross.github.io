---
showTableOfContents: true
title: "Client + Server Communication (Theory 8.1)"
type: "page"
---
## Overview
So we now have a fully functioning reflective loader that obfuscates it's payload using rolling XOR with the ability to derive a key from environmental factors. But up until this point, we've always
been loading our DLL from disk, which of course defeats the entire purpose of reflective loading. Yes, we said that we want to avoid using `LoadLibrary`, but the reason for wanting to avoid using
`LoadLibrary` is that it requires us to load the DLL from disk.

We are avoiding its use so that our DLL never has to touch disk. A reminder - AV/EDR is capable of scanning/detection both on disk and in memory. But whereas the former is relatively cheap (in terms of impact on system performance), the latter has a much bigger performance costs. So whereas most systems may likely scan every thing that touches disk, most systems have to be much selective in terms of what they choose to scan in-memory. Hence why are goal is to avoid our shellcode-containing DLL from touching disk.

In this module we'll now implement a simple Client + Server mechanism so that our DLL will avoid touching disk. We will place our loader on the target system, but our DLL will stay on a remote system. When our reflective loader runs it acts as a client, sends a request to a server, and the server thus send the DLL over to it, whereafter it is injected directly into memory. So it does touch disk initially, just not the disk of the target machine that matters.


In this section we'll focus on establishing basic, secure communication between our agent (client) and a server using Go's built-in networking capabilities.

## Client/Server Communication (Go) & HTTPS

To handle the download of the DLL payload, we'll build a client that makes requests and a server that responds with the payload. Go's standard library provides convenient tools for this, making our lives incredibly simple. We'll mostly use the `net/http` package.

## Go `net/http` Package

The `net/http` package contains extensive functionality for building both HTTP clients and servers.

We  can create simple web servers that listen for incoming HTTP requests on a specific port. We'll define **handler** functions that specify how the server should respond to requests made to different URL paths (**endpoints**). In our case of course the handler will retrieve, and then serve, our obfuscated payload.


The `net/http` package also allows us to easily create HTTP clients to make requests (GET, POST, etc.) to remote servers and process their responses. This will allow us to send the request to the server, and then process the response (i.e. the payload).

## HTTPS for Secure Communication

When transmitting potentially sensitive data like a DLL payload (even an obfuscated one), using plain HTTP is obviously insecure. The traffic could easily be intercepted and inspected by anyone monitoring the network (Man-in-the-Middle attack). We'll therefor use  **HTTPS (HTTP Secure)**, which is of course HTTP layered over **TLS (Transport Layer Security)**, formerly known as SSL. TLS provides encryption for the data in transit, ensuring confidentiality and integrity.


## Server-Side HTTPS Implementation

- We'll create a HTTP server using `http.ListenAndServeTLS`.
- We'll need to generate (using `openssl`), and provide as arguments  a **TLS certificate file** (`.crt` or `.pem`) and the corresponding **TLS private key file** (`.key` or `.pem`).
- It's of course imperative that the private key is kept secret on our server and is used to establish the secure connection.
- Because we'll use a self-signed cert we'll have to enable a special configuration in our client won't automatically trust the self-signed certificate - see below.

## Client-Side HTTPS Implementation
- When our Go `net/http` client connects to an HTTPS URL, it automatically attempts to perform a TLS handshake.
- By default, the client verifies the server's certificate against a list of trusted CAs known to the operating system. If the server presents a certificate that is expired, invalid, or signed by an unknown (untrusted) CA (like our self-signed certificate), the Go client will refuse to connect and return an error.
- To allow our client to connect to our server using a self-signed certificate, we'll set the `InsecureSkipVerify` field within the `tls.Config` struct to `true`.


In the next section we'll look at our specific communication protocol, including how the client identifies itself and how the server responds.


---
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../module07/key_lab.md" >}})
[|NEXT|]({{< ref "protocol.md" >}})