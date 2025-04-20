---
showTableOfContents: true
title: "Project Setup (Lab 00)"
type: "page"
---

## Overview
Before we start creating our application I'd like to run over how to start a new project (module)
in Go, as well as how to set up version control using `git`.


## IDE
**A quick note on IDE use for this course.**

I use [Goland](https://www.jetbrains.com/go/) from Jetbrains as my IDE - BIG fan. It is a paid-for IDE, and so if you're not 
quite ready to dish out some sweet moolah for something you've grown accustomed to getting for free - I gets it.
Note however that you could get a 30-day free trial to use Goland for this course if you so desire - _totally_ up to you.

If you're also a GUI kiddie like me and want a free alternative then VSCode with all the required
plug-ins for Go development is aight. I mean, it's Microsoft, so y'know, temper those expectations, but 
it'll mostly work okay-ish. 

Something else you could try out is [Zed](https://zed.dev), it's written in Rust and has good support for LLM integration,
as well as Go development. I never gave it a second look until I recently watched a tutorial by 
[IppSec](https://ippsec.rocks) in which he mentioned it being his IDE of choice. Since that dude is 
operating on some cosmic strata of hackery, I have to assume it's legit. So feel free to give it a go.

Finally, if you really wanna cool and leet then obviously a terminal editor is the way to go, both
NeoVim and Helix can be set up for Go development. If you're not used to the bindings yet, you might
want to set a week or so aside to just learn that, then meet us back here. 


## Modules and Packages
Before we actually set up our "project" I'm just going very briefly touch on the Go-specific nomenclature for what's
more general referred to as projects, libraries etc.

Think of Packages first. These are basically just folders containing Go files (.go files) that logically belong together – 
like, all your code for handling user stuff might live in a `user` package. It's Go's fundamental way of organizing code 
chunks within your project. You use the package keyword at the top of your Go files to declare which package they belong to - much more on this later.

Then you've got Modules. A module is essentially a collection of related Go packages that are versioned together as a single unit. 
It's what defines your project, or maybe a reusable chunk of code (what you'd typically call a 'library' in other languages).

So when we create a new "project" in Go, we are actually creating a new module. An awesome feature of a Go module is that it keeps 
track of all the other modules (your project's dependencies) that your code needs to run, 
and crucially, which specific versions are being used. 

So unlike in Python for example where you need to generate a `requirements.txt` with what packages you are using, and then 
someone that wants to run you application needs to use something like `pip` 
to download them in a unique virtual environment (to ensure versions don't clash), Go does all of this for you. And since
packages are assigned on a per-module basis, and not system-wide, we don't need to use virtual environments with Go.



## Setting Up Our Project
When one creates a new project in Goland it automatically initializes it as a Go module. Since not everyone will
be using Goland I want to illustrate how this could be done manually, it's probably a good idea to know how to do this
regardless of what IDE you are using.

Now at this point, aside from having an IDE, I'm also assuming you have Go installed. If not follow the guide [here](https://go.dev/doc/install). 

In your terminal, navigate to the parent folder you typically save your projects to, in my case for example this would be `~/repos`.
Now let's create a new directory, name it after your project (in my case `orlokC2`), and `cd` into it.

```
mkdir orlokC2

cd orlokC2
```

Our directory is currently empty, so let's go ahead and initialize it as a Go project.


```
go mod init github.com/faanross/orlokC2
```

You'll notice that I did not simply name it `orlokC2`, but instead the full path for the repo URL where I _intend_ to host it.
This is a convention, reason being is that by doing it this way will make it easy for other developers to use your code in the 
future, all they would need to do is run the command `go get github.com/faanross/orlokC2`. The `get` tool will automatically 
recognize `github.com/faanross/orlokC2` as the module path, and know _how_ and _where_ to download the source code.

That being said, if you are 100% positive you have 0 intentions of ever sharing your code with anyone, you can name it
whatever you want.

If you now run `ls` you'll notice we have a new file in our directory called `go.mod`. And if you `cat` out the contents
you'll see something like this:

```go
module github.com/faanross/orlokC2

go 1.23.3
```

Now this file will automatically be updated as we develop our application and import other modules, but for now it tell us
2 fundamental things - the name of our module, and the version of Go we are using.

## Initializing Version Control

I personally like to set up version control, and even pushing it to an online repo,
before I even get coding, but that is of course a preference. I'm going to show you how I typically go about doing this,
while keeping the explanations of what everything means to a minimum.

First, still in our project directory, let's initialize our project using `git`.

```
git init . 
```

You'll see some output messages, don't worry about that too much for now, the most important part is that the final
sentence begins with "Initialized empty Git repository in...".


## .gitignore and README

Next I like to manually create `.gitignore` and `README.md` myself.

```
touch .gitignore

touch README.md
```

Just in case you don't know what these are, I'll offer a brief explanation. 

`.gitignore` contains a list of everything 
that you DO NOT want `git` to push from your local to your online repo. This usually includes "artifacts of your environment" - 
for example I use Goland, and work on Mac OS, so both my IDE and OS create hidden files and folders in my directory containing
things like user preferences etc. It won't be a huge deal if these get pushed (in most cases), but still it's noise and not required.

Where this however becomes absolutely crucial is when you for example start using API tokens for microservices, or PSK for
authentication, or TLS certs for encryption etc. In other words, sensitive data that could be abused by others with ill intent.
So for now I typically just populate it with a template I've developed over time, if you've never created one you can use
a free service like [gitignore.io](https://www.toptal.com/developers/gitignore/) to generate a template based on your specific tools, 
environment, and project details.

`README.md` is what you see when you open a repo on Github, and typically follows a conventional format. You could deviate
from this again if you'd like, but you should check some key boxes - what is the application, what is its purpose,
requirements, how to install, how to use etc. For now we'll leave it blank since this is typically done once your 
application is done and ready to be shared with the rest of the world. 

## git add and commit
With that out of the way we can now stage everything in the directory, as well as `commit` the stages.

```shell
# stage everything in directory
git add .

# commit the stages
git commit -m "Initial project setup"
```

With staging we select all the changes you want to save, while committing permanently records those selected changes 
as a snapshot in our project's history with your descriptive message.

The next (optional) step is then to `push` this commit (i.e. the "snapshot") to an online repo, but since we've not yet
created this repo, now would be the time to do so.

## Creating an online repo
So go to `github.com` > Repositories > New. Give it a name, description (optional) and select either Public or Private.
This is up to you, but typically while something is a work-in-progress its good to keep it private, at least up
until the point its in working state and has a README that makes it clear what the purpose of the application is,
how to install it, use it etc.

Don't add a README, nor a gitignore (since we already did this), and if you know what a license entails and would like
to add it - go for it - otherwise just select `None`. You can now click `Create Repository`. 

Following this you can now head back to the terminal where we'll run our final commands:

```shell
# This command connects your local repository to a remote repository on GitHub.
git remote add origin git@github.com:faanross/orlokC2.git

# This command renames the current branch from `master` to `main`.
git branch -M main

# This command pushes your local `main` branch to the remote repository and sets it as the upstream branch.
git push -u origin main
```

After you've run this last command you can go back to the repo on Github, and hit refresh - you should now see your 
files live on your online repo. Without getting into the finer nuances and all situations you'll possibly encounter in
Github, in general, in about 95% of situations, you'll only ever need to run 3 commands.

Let's say you've been working a bit and would like to stage all your new files and changes to commit, and push it to
your repo, just run (in the root directory where the .git file can be located):

```go
git add .
git commit -am "some message here reflecting what changes/updates were just made"
git push
```

## Conclusion
And that's it, a super simple, concise summary of how to start a new Go module ("project"), how to set up version control, 
and how to synchronize your local repo with an online version. Now that all the groundwork has been laid let's
go ahead and get cracking on our actual application.

___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "overview.md" >}})
[|NEXT|]({{< ref "../part_b/lab01.md" >}})