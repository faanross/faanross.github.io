---
showTableOfContents: true
title: "LAB 00 - Project Setup"
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


## Setting Up Our Project
When one creates a new project in Goland it automatically initializes it as a go



- Now when you create a new go project it actually needs a few special files
- Usually when you create it in Goland it will just do this automatically, but in the interest of being agnostic I want to show you how to do this manually

- Now of course I'm assuming you already have go installed, if not then Google it, follow instructions
- I'm on mac and if you are too I recommend installing it via brew

- Create folder for our project

```
mkdir orlokC2

cd orlokC2
```


- And so our project is called orlokC2, but we're actually going to give our module (that's the go term for our entire project), another name based on the full repo URL

```
go mod init github.com/faanross/orlokC2
```

- Now if you 100% have 0 intentions of ever sharing your project, that's fine you could just name it orlokC2
- But doing it this way will make it easy for another Go project if they want to use your code, they will `import "github.com/yourusername/yourproject/somepackage"`.
- The Go tools parse this import path, recognize `github.com/faanross/orlokC2` as the module path, and know _how_ and _where_ to download the source code (using Git to clone from GitHub in this case).

- We can now see a new file we've created there
