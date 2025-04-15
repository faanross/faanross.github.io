---
showTableOfContents: true
title: "Project Setup (Lab 00)"
type: "page"
---

## Overview
This is going to be a very short lab, we'll just get going with a new Go project, 
and then we'll also implement basic version control using git.


## IDE
I use Goland from Jetbrains as my IDE - BIG fan. It is a paid-for IDE, and so if you're not 
quite ready to dish out some sweet moolah for something you've grown accustomed to getting for free - I gets it.
Note however that you could get a 30-day free trial to use Goland for this course if you so wish.

If you're also a GUI kiddie like me and want a free alternative then VSCode with all the required
plug-ins for Go development is aight. I mean, it's Microsoft, so y'know, temper those expectations. 
It'll mostly work fine. 

Something else you could try out is Zed, it's written in Rust and has good support for LLM integration,
as well as Go development. I never gave it a second look until I recently watched a tutorial by 
[IppSec](https://ippsec.rocks) in which he mention it being his IDE of choice. Since that dude is 
operating on some cosmic strata of hackery, I have to assume it's legit.

Finally, if you really want be cool and leet then obviously a terminal editor is the way to go, both
NeoVim and Helix can be set up for Go development. If you're not used to the bindings yet, you might
want to set a week or so aside to just learn that, then meet us back here. 

Before we actually set up our "project" I'm just going very briefly touch on the Go-specific nomenclature for what's
more general referred to as projects, libraries etc.

## Modules and Packages
Before we initialize our project, let's quickly cover some Go terms you'll hear a lot: packages and modules.

Think of Packages first. These are basically just folders containing Go files (.go files) that logically belong together – 
like, all your code for handling user stuff might live in a `user` package. It's Go's fundamental way of organizing code 
chunks within your project. You use the package keyword at the top of your Go files to declare which package they belong to - much more on this later.

Then you've got Modules. A module is essentially a collection of related Go packages that are versioned together as a single unit. 
It's what defines your project, or maybe a reusable chunk of code (what you'd typically call a 'library' or 'framework' in other languages).

So when we create a new "project" in Go, we are actually creating a new module. An awesome feature of a Go module is that

The key thing about a module is that it keeps track of all the other modules (your project's dependencies) that your code needs to run, 
and crucially, which specific versions of them. This is getting a bit abstract, let's just get going and I'll show you
exactly what I mean here.


## Setting Up Our Project
When one creates a new project in Goland it automatically initializes it as a Go module. Since not everyone will
be using Goland I want to illustrate how this could be done manually, it's probably a good idea to know how to do this
regardless.

Now at this point, aside from having an IDE, I'm also assuming you have Go installed. If not follow this guide [here](https://go.dev/doc/install). 

In your terminal, navigate to the parent folder you typically save your projects to, for me for example this would be `~/repos`.
Now let's create a new directory, name it after your project (in my case `orlokC2`), and `cd` into it.

```
mkdir orlokC2

cd orlokC2
```

Our directory is currently empty, so let's go ahead and initialize it as a Go project.


```
go mod init github.com/faanross/orlokC2
```

You'll notice that I did not simple name it `orlokC2`, but instead the full path for the repo URL where I _intend_ to host it.
This is a convention, and like all conventions feel free to break it, we could just call it `orlokC2` if we so desired.
The reason it a convention however is that by doing it this way will make it easy for other developers to use your code in the 
future then all the need to do is run the command `go get github.com/faanross/orlokC2`. The `get` tool will automatically 
recognize `github.com/faanross/orlokC2` as the module path, and know _how_ and _where_ to download the source code.

If you now run `ls` you'll notice we have a new file in our directory called `go.mod`. And if you `cat` out the contents
you'll see something like this:

```go
module github.com/faanross/orlokC2

go 1.23.3
```

Now this file will automatically be updated as we develop our application and import other modules, but for now it tell us
2 fundamental things - the name of our module, and the version of Go we are using.

## Initializing Version Control

I personally like to set up version control (in case of oopsies), and even pushing it to an online repo (for backup purposes),
before I even get coding, but that is of course a preference. I'm going to show how I like this (the how), but I'm going to 
keep the explanations of what each means and does (they what and why) to a minimum.

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

Just in case you don't know what these are, I'll offer a brief explanation. `.gitignore` contains a list of everything 
that you DO NOT want `git` to push from your local to your online repo. This usually includes "artifacts of your environment" - 
for example I use Goland, and work on Mac OS, so both my IDE and OS create hidden files and folders in my directory containing
things like user preferences etc. It won't be a huge deal if these get pushed (in most cases), but still it's noise and not required.

Where this however becomes absolutely crucial is when you for example start using API tokens for microservices, or PSK for
authentication, or TLS certs for encryption etc. In other words, sensitive data that could be abused by others with ill intent.
So for now I typically just populate it with a template I've developed over time, if you've never created one you can just use
a free service like [gitignore.io](https://www.toptal.com/developers/gitignore/) to generate one based on your specific tools, 
environment, and project details.

`README.md` is what you see when you open a repo on Github, and typically follows a conventional format, though again this 
is largely based on personal preference. For now we'll leave it blank, but we'll create a simple README once our project is
done. 

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
