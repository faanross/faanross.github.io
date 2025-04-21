---
showTableOfContents: true
title: "Brief Introduction to Goroutines (Lab 02)"
type: "page"
---

## Overview
In our previous lab I alluded to the fact we could not take any action once we called the `ListenAndServe()` function, 
since this function was "blocking". But what does that even mean? 
Now unfortunately I cannot give the beautiful and somewhat complex topic of concurrency it's proper treatment here and now
since our time is limited. But I'll do my best, and for the deep dive please see the Director's Cut.

I want you to imagine that when we start our program, we automatically get one "worker". 
Now picture this worker any way you want, but I mean it literally - it's some digital homunculus that goes 
through your code, line-by-line, and it's the driving force behind "executing" the code.

Now usually, this homunculus does job A, then job B, then job C; which works out just fine. It's a very fast and efficient worker, and it all gets done so quickly you don't even notice.

But sometimes there are situations that there are many different jobs, some of them could be really demanding, 
and some of them fracture into other sub-jobs. And so in these cases, of course, we'd benefit from having more than one worker
so things don't move too slowly for our liking.

But aside from this, what I really want to discuss here are specific type of jobs that's so involved, 
that once a worker starts doing it, that's it, they're 
completely tied up and unable to do anything else.

So if we had job A, job B, and job C and job B is such a job, well then our worker will never even get to job C. In this case, we call job B "blocking", meaning simply that it "blocks" our worker from doing anything else.

Now we don't actually use the term workers, rather they are called "threads", and one more thing you should know about a thread is that it's a system resource that is issued, and managed, by the OS.

When you start an application, you always get one thread automatically, this is typically called our main thread. And then if you want more than one thread, that is if you want extra workers either to make your program more efficient, or to ensure it does not get blocked if such operations exists, then you need to manually create more such threads.

Most languages can do this, this is known as concurrency. But, Go's handling of concurrency is unique, and quite frankly friggin awesome, for three main reasons.

**First**, it gives us a number of tools in the concurrency toolbox, some of which are unique to the language. 
These include: goroutines, select statements, channels, mutexes, and waitgroups. We'll cover all of them except for waitgroups in
this workshop.

**Second**, it's syntax is RIDICULOUSLY simple when compared to other languages' handling of concurrency.

**Third**, it's concurrency system is remarkable efficient, for a reason I will explain next.

Remember when I said above our "worker" is a thread that's created and managed by the OS? Now in Go, this is not true. 
Application written in Go are statically linked with a special runtime during compilation, and though yes the OS does issue 
our application a thread like with any other application (otherwise our `main()` function could not execute), the Go runtime takes this single thread and magically turns it into many (possibly thousands) of little "mini-workers" called Goroutines.

So in Go, execution is not done by threads, but by Goroutines. And its an awesome and extremely efficient system 
because, in the vast majority of situations, an entire OS thread is complete overkill. Imagine you want to go to the 
store to buy a bag of peanuts, but all you have is an ounce of gold. You show up at the checkout and present it, 
and now they have to go and call the manager, and find some extra wads of cash just to be able to give you enough change.
This is a very inefficient way to buy a bag of peanuts since an oz of gold is overkill.

Wouldn't it have been much better if you could exchange the bar of gold for a whole pile of $10 bills beforehand, and then simply take a single bill to the store to go buy some peanuts? That's essentially what's happening here - Go takes the ounce of gold (OS thread), and divides into potentially thousands of $10 bills (Goroutines), allowing for much more efficient transactions, and for many more transactions to ultimately be able to occur at the same time.

## Blocking problem

Even though I've explained the blocking problem above, I want to actually show you, so let's add the following comment at the end of 
our `main()` function after calling `http.ListenAndServe()`.

```go
func main() {

	r := chi.NewRouter()

	router.SetupRoutes(r)

	serverAddrPort := fmt.Sprintf("%s:%s", serverAddr, serverPost)

	log.Printf("Listening on %s\n", serverAddrPort)

	err := http.ListenAndServe(serverAddrPort, r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

	// ADD THIS
	fmt.Println("This code will NEVER execute!")

}
```

![lab02](../img/lab02a.png)

So now you can see on the left that, despite running the server and even connecting to it (on the right), that comment
does indeed never execute. And as we explained, we have a single goroutine, and that function blocks it, so it never
gets to move beyond it and execute the next goroutine.

## Adding a Goroutine

As described above, since we have a blocking function (`http.ListenAndServe()`), we need to call it in its own new
goroutine. And to show you just how easy that is in Go:

```go
go func() {
err := http.ListenAndServe(serverAddrPort, r)
if err != nil {
    log.Fatalf("Server failed to start: %v", err)
}
}()
```

As we can see in this case since we want to include our error check in our new goroutine we wrap it all in an 
anonymous function (`func()`), which we precede with the keyword `go`. So now our main goroutine is not going to concern
itself with this blocking function, it has its very own goroutine, and so now we should be good and our print statement
should execute. Let's see...



![lab02](../img/lab02c.png)


And so this time around we do execute our _print_ statement, but our application also immediately exits. Curious.
Well, the issue now is that, funnily enough, nothing is blocking our main goroutine. Our main goroutine no longer handles 
the `http.ListenAndServe()`
function - that's done by our newly minted goroutine. So it moves on, handles the print statement, and thereafter since it
is then free with nothing else to do it finished our `main()` function, which in turn signals to our application to exit.

## Signal Handling

So what we want to do now is reintroduce some form of blocking at the end, but then also make it contingent upon
some action we can control to unblock, and thus exit. And the way we'll do this is with signal handling.

So let's add the following 2 lines right at the top of our sever's main function.

```go
sigChan := make(chan os.Signal, 1)  
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
```

Rather than break these 2 lines of code down bit by bit, I think it's easier if I just explain the entire thing
conceptually. 

As most of you probably know if you run almost any application and hit Ctrl + C, the application will immediately exit. 
This is of course how we have been killing our server process thus far, but the question then becomes - how is this happening?
We never added any logic to our application to tie those two keys with an instruction for our application to exit, so what gives?
Well, it's not application itself that is exiting. 

You see, whenever you run an application, and it's then the active process, the OS is still "listening", and keeping tabs of
what's going on. The process itself (i.e. our application) is itself of course being managed by the OS, and it's the OS
that detects when a user inputs Ctrl + C. This is called SIGTERM, and the OS "knows" that whenever the user hits Ctrl + C
they are asking it to immediately and abruptly close the current active process.

So what we're doing above is we're creating a new channel called sigChan. `Channels` are used for Goroutines to communicate
with one another, for now I just want you to think of this channel as a signal. When we create it, it's "deactivated",
but as soon as we "activate" it, it then serves as a trigger for us to be able to do something. 

So with this code we're saying: We don't want SIGTERM (i.e. Ctrl + C) to be interpreted by the OS as a request to
immediately kill the current process anymore. Instead, what we want is for our sigChan to be activated.

And so then what happens once sigChan is activated? Well let me show you the one other line of code we'll add, 
right at the bottom of our main() function.

```go


func main() {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// AFTER THE NEW GO FUNC()

	<-sigChan

	fmt.Println("Program will now exit.")

}

```

So you can see we added `<-sigChan`. What this does is it block the main goroutine, when it gets there it stops. But,
we have the ability to unblock it, and perhaps you've already guessed it. When we hit Ctrl + C and activate sigChan
it serves as a signal to unblock, and thus allow the main goroutine to proceed past it, finish execution of the main()
function, and thus end our application. 


And if we run it our application we'll see we're all good - the statement will execute _and_ our program will continue to run,
until we hit Ctrl + C to end it. 

![lab02](../img/lab02b.png)


## Conclusion
Ideally we'd now also add the ability to gracefully stop our listener and its goroutine. Alas,
that luxury is not available to us in our allotted time - I decided to cut it since it might be the "right way" to do things,
but not doing it still works just fine. 

However, if you're keen to learn how to do it, which I do encourage, see the Director's Cut. 


___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "lab01" >}})
[|NEXT|]({{< ref "../part_c/lab03.md" >}})