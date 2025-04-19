---
showTableOfContents: true
title: "Brief Introduction to Goroutines (Lab 02)"
type: "page"
---

## Overview
In our previous lab I alluded to the fact we could not take any action once we called the `ListenAndServe()` function, since this function was "blocking". What does that even mean? Now unfortunately I cannot give the beautiful and somewhat complex topic of concurrency it's proper treatment here and now, but I'm going to do my best to give you some inkling.

I want you to imagine that when we start our program, we automatically get one "worker". Now picture this worker any way you want, but I mean it literally - it's some digital homunculus that goes through your code, line-by-line, and it is behind "executing" the code.

Now usually, this homunculus does job A, then job B, then job C; which is fine. They're a very fast and efficient worker, and it all gets done so quickly you don't even notice.

But sometimes there are situations that there are many different jobs, some of them could be really demanding, and some of them fracture into other sub-jobs. And so in these cases, of course, we'd benefit from having more than one worker. But aside from this, sometimes there's a job that's so involved, that once a worker starts doing it, that's it, they're completely tied up and unable to do anything else.

So if we had job A, job B, and job C and job B is such a job, well then our worker will never even get to job C. In this case, we call job B "blocking", meaning simply that it "blocks" our worker from doing anything else.

Now we don't actually use the term workers, rather they are called "threads", and one more thing you should know about a thread is that it's a system resource that is issued, and managed, by the OS.

When you start an application, you always get one thread automatically, this is typically called our main thread. And then if you want more than one thread, that is if you want extra workers either to make your program more efficient, or to ensure it does not get blocked if such operations exists, then you need to manually create more such threads.

So most languages can do this, this is known as concurrency. But, Go's handling of concurrency is unique, and quite frankly friggin awesome, for three main reasons.

**First**, it gives us a number of tools in the concurrency toolbox, some of which are unique to the language. These include: goroutines, select statements, channels, mutexes, and waitgroups.

**Second**, it's syntax is RIDICULOUSLY simple as compared to most other languages.

**Third**, it's concurrency system is remarkable efficient, for a reason I will explain below.

Remember when I said above our "worker" is a thread, that's created and managed by the OS? Now in Go, this is not true. Any application written and compiled in Go has a special runtime statically linked to it, and though yes the OS does issue our application a thread like with any other application (otherwise our main() function could not execute), the Go runtime takes this single thread and magically turns it into many (possibly thousands) of little "mini-workers" called Goroutines.

So in Go, execution is not done by threads, but by Goroutines. And its an awesome and extremely efficient system because, in the vast majority of situations an entire OS thread is overkill. Imagine you want to go to the store to buy a bag of peanuts, but all you have is an ounce of gold. You show up at the checkout and present it, and now they have to go and call the manager, and find some extra wads of cash just to be able to to give you enough change.

Wouldn't it have been much better if you could exchange the bar of gold for a whole pile of $10 bills beforehand, and then simply take a single bill to the store to go buy some peanuts? That's essentially what's happening here - Go takes the ounce of gold (OS thread), and divides into potentially thousands of $10 bills (Goroutines), allowing for much more efficient transactions, and for many more transactions to ultimately be able to occur at the same time.

## Blocking problem

Even though I've explained the blocking problem above, I want to actually show you, so let's add the final comment at the end of our `main()` function after calling `http.ListenAndServe()`.

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

**ADD IMAGE HERE**

And so this time around we do execute our _print_ statement, but also executed immediately - curious. Well, the issue now is
that, funnily enough, nothing is blocking our main goroutine. Our main goroutine no longer handles the http.ListenAndServe()
function - that's done by our newly minted goroutine. So it moves on, handles the print statement, and thereafter since it
is then free with nothing else to do it finished our main() function, which in turn signals to our application to exit.

## Intentional Block
Now there are much better ways to do this then we're about to. Typically, we want to combine channels with signal handling
so that we can intentionally block our `main()` function before it ends with the ability to gracefully kill it using a
signal (like Ctrl + C). But, for now we're going to use a simple little "hack" - an empty `select` statement.

In some ways its analogous to an empty for loop - it just creates an endless hole of logical continuity. Like I said, not
very elegant, but it's gonna get the job done for now. So add this right the to bottom of your code.

```go
func main() {

	r := chi.NewRouter()

	router.SetupRoutes(r)

	serverAddrPort := fmt.Sprintf("%s:%s", serverAddr, serverPost)

	log.Printf("Listening on %s\n", serverAddrPort)

	go func() {
		err := http.ListenAndServe(serverAddrPort, r)
		if err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	fmt.Println("This code will NEVER execute!")

	select {}
}
```

And if we run it again we'll see we're all good - the statement will execute _and_ our program will continue to run.

![lab02](../img/lab02b.png)


## Conclusion
Ideally we'd now also add channels + signal handling to allow our goroutines and listeners to all stop gracefully. Alas,
that luxury is not available to us in our allotted time - I decided to cut it since it might be the "right way" to do things,
but not doing it still works just fine. However, if you're keen to learn how to do it, which I do encourage, see the Director's Cut. 


___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "lab01" >}})
[|NEXT|]({{< ref "../part_c/lab03.md" >}})