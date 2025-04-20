---
showTableOfContents: true
title: "Overview of Our Project"
type: "page"
---
## Let's Begin at the End
I'd like to start with a quick overview of what we'll create by the end of today's workshop. I'll present the final 
diagram here, and then we'll deconstruct it during the remainder of this lecture.

![final_project](../img/final.png)

Now even this is a bit of a simplification, since I did not include _every_ single function and data structure on this diagram.
Instead, my goal was to illustrate some semblance of the lynchpin components, and how they fit together.
Now perhaps this seems overwhelming, but honestly by the end of our workshop this is going to actually seem incredibly trivial, and 
you'll be ready to take it to the next level, I promise.

For now however I'm going to show the progression as we go through each section, starting with our server.

## Server
![server](../img/server.png)

In our first section we'll create the foundation of our server. 
We'll first create a **Listener** that will bind to a port to create a socket, accept an incoming request and create a connection instance.

But a connection by itself is not very interesting. And so in order for it to "do something", we'll furnish it with a **router**, which allows
us to define **routes**. You can think of a route simply as defining a method + endpoint in relation to an action. So for example: if
a GET request (**method**) is sent to the root **endpoint** `/`, then call the **handler** `RootHandler` (**action**).

Since we won't yet have an agent at this stage, we'll use `curl` as a substitute to test our server's functionality. 

## Agent
![agent](../img/agent.png)

After we've created our basic server we'll create the foundation of our agent. We'll start by creating a **config** system,
which will allow us to configure important parameters such as target host, sleep, jitter, check-in endpoint etc. 

All the **agent's main logic** will live in a function called `runLoop()`, which connects, sends requests, processes responses, sleeps, then 
repeats that process until it's instructed to stop. 

Once that's in place we'll also learn about **UUID** and ensure our agent has its own **UUID**, and then create **middleware** on the server
side capable of intercepting requests and parsing the UUID so our server knows who it's communicating with.

## Client 
![client](../img/client.png)

In this section we'll set up a **basic client UI**. Now it won't do much at this point, but we want to put the structure
in place so that in the following section we can tie everything together.

We'll create our **Vue.js** Web UI frontend using **Vite**, which means we'll run a simple command to put an entire template
in place, complete with a local web server so we can run a development version of our client on our local host.

Once that's done we still need to connect our server to our new client UI, for this course we'll use websockets so that
our server can push real-time updates to our client - no need for polling. So on the server side we're going to set up
a websocket server, as well as an all-important receiver function called `handleWebsocket()`, which will serve as the
entrypoint for our client on the server side. On the client side we'll also add a simple js websocket implementation.

**PLEASE NOTE**: Since this is primarily an offensive security course, not a frontend dev course, we're going to blast
through any section that involves JS. I won't completely blackbox it, but unlike our sections on Golang where we'll actually
build out our logic line-by-line, I'm going to C+P the code and very briefly cover it at a high-level. This is the only way that
I could include us having a decent-enough frontend for our C2 framework, while still getting everything done in the allotted
time. If you're positively Jonesing to learn Vue.js, there are numerous free courses on Youtube, pick one and start building
something as soon as is feasible - that's how I got started. [Here's one](https://www.youtube.com/watch?v=VeNfHj6MhgA) that will get you going nicely.



## Weaving It All Together
![final_project](../img/final.png)

Now that we've completed the "circuit" - client to server to agent, back to server to client - we can now weave everything
together. In this case, since the focus of this workshop is really to build out the entire framework instead of advanced
endpoint behaviour, our agent will have a very simple command interpreter of sorts that will be 
capable of running 3 core commands on the endpoint - `whoami`, `pwd`, and `hostname`. 

The first thing we need to do is actually give our agent the ability to run these commands, and capture the output as strings.

Once this is done we'll start at the client and work our way through the sequence as it would happen:
- We give our client the ability to allow us to select a command, and then send that to the server.
- The server then processes the command, and queues it.
- The agent checks in, receives the command from the queue.
- The agent then runs the command, captures the output, and returns the result to the server.
- Finally, our server sends it back to the client where it is rendered in-browser. 

## Conclusion
That's it! It's going to be a fun ride, so without any further delay, let's get right to it.


___
[|TOC|]({{< ref "../moc.md" >}})
[|NEXT|]({{< ref "setup.md" >}})