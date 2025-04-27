---
showTableOfContents: true
title: "Agent Command Execution (Lab 08)"
type: "page"
---

## Overview
There's one final "building block" we need to put into place before we can start weaving everything together. In this lab we'll integrate the functionality into our agent to be able to run the following 3 commands:
- `whoami` to get user name
- `pwd` to get the present directory
- `hostname` for the host name

In addition to running the commands we also want to ensure we are able to capture the output as a string since we intend to send the result back to our server in a future lab.

## internal/agent/commands/commands.go

Let's create a new directory called `commands` inside of `internal/agent`, and inside of it create a file called `commands.go`.

The first function we'll create is our "command router" - it takes the specific command that arrived from the server, and then based on it's value decides which of our 3 functions to call.

```go
package commands

// Execute runs the specified command and returns the output
func Execute(cmd string) (string, error) {
	// Trim any whitespace
	cmd = strings.TrimSpace(cmd)

	// Check which command to run
	switch cmd {
	case "pwd":
		return Pwd()
	case "hostname":
		return Hostname()
	case "whoami":
		return WhoAmI()
	default:
		return "", fmt.Errorf("unknown command: %s", cmd)
	}
}
```

The first thing to note is that we're doing a bit of "code hygiene" immediately inside of the function - we call `TrimSpace()` on our input to ensure there are not trailing or leading whitespaces, which will lead to the command not being correctly matched in the switch statement that follows.

Speaking of, the switch statement takes the command, and calls the corresponding function + returns its return value. These functions don't exists yet, so let's go ahead and create them.

First off create `Pwd() `and `Hostname()` since they are both extremely simple and quite similar in structure.

```go
// Pwd returns the current working directory
func Pwd() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, nil
}
```


```go
// Hostname returns the machine hostname
func Hostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}
```


You can see for both of these we are using the Go standard library `os` package, which provides us with a platform-independent interface to operating system functionality. Meaning it will handles the platform-specific details internally, giving us a single wrapper function to call a specific OS API function in a cross-platform compatible manner.


```go
// WhoAmI returns the current user
func WhoAmI() (string, error) {
	cmd := exec.Command("whoami")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
```


So whereas with `Pwd()` and `Hostname()` we are leveraging Go's direct, platform-abstracted interface to OS functionalities, with `WhoAmI()` we delegate the task to an external command-line utility (`whoami`) and processes its output. In other words, we are calling the command directly via the operating system's command execution mechanism.


Great, so we have these 4 new functions in `commands.go` - `Execute()` is the wrapper that decides which of the 3 specific functions - `Pwd()`, `Hostname()`, and `WhoAmi()` - to call.

So that's all we have to do in this section, but I would like to test it, so let's head back to `agent.go` and create a contrived function that calls all of them, just so we are sure it actually works.


## internal/agent/agent/agent.go
At the bottom of `runLoop()`, just before we get to our Sleep(), let's call a new function called `CommandsTest()`.

```go
func (a *Agent) runLoop() {
	for {
		select {
		case <-a.stopChan:
			return
		default:
			// OTHER CODE HERE
			// Process response

			 // RIGHT HERE BEFORE SLEEP ADD THIS

			CommandsTest() // ADD THIS LINE

			time.Sleep(sleepTime)
		}
	}
}
```

And then at the bottom of the file add the actual function implementation, which is very simple - we're just providing each of our 3 command input strings to `Execute()`, and printing the result.

```go
func CommandsTest() {
	presentDir, _ := commands.Execute("pwd")
	fmt.Printf("pwd is %s\n", presentDir)

	currentUser, _ := commands.Execute("whoami")
	fmt.Printf("user is %s\n", currentUser)

	currentHost, _ := commands.Execute("hostname")
	fmt.Printf("host is %s\n", currentHost)
}
```

## test

Run the server, and then agent - we should now see each of the function's output on the agent side.

![lab08](../img/lab08a.png)






___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../part_d/lab07.md" >}})
[|NEXT|]({{< ref "lab09.md" >}})