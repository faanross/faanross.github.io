---
showTableOfContents: true
title: "YAML-based Configuration Management System"
type: "page"
---
## Solutions
The **starting** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson02_Begin).
The **final** solution can be found [here](https://github.com/faanross/workshop_antisyphon_18092025/tree/main/Lesson02_Done).

## Overview
We saw in the previous lesson that we have a `Config` struct in our application acting as the type to house all our different configuration properties.

Now there are many different ways we can use this.

As we already saw, we can just define a struct literal right there in our `main` function and assign the values to it. This is the most direct, bare bones technique.

Another option would be to create a constructor which we could house in our `config` package, and then call it from `main` by passing all the desired values as arguments. This is on one hand a bit more contrived, but it keeps our `main` function clean and also is more conducive to the implementation of validation logic.

A way I prefer to typically handle configs however is by implementing a YAML-based system. I like this for the simple reason that I believe it creates a more user-friendly interface when we specify our desired config values. Instead of looking through our code where we're specifying the values, we have a clean, separate file written in YAML which is probably the closest you're gonna get to pure English in development.

Now, as was the case in the previous lesson when I showed how to create an embedded struct - in all honesty, creating a YAML-based implementation system is probably a little bit overkill for this application. This is because it comes with overhead - we now also have to implement a loader, which will read the YAML, create the struct, and unmarshall the YAML values into the struct.

I'm once again choosing to do this since I think it's a great touch when projects become larger and more complex, and so I wanted you to know how you can do this. Besides, it's really not that much effort, and as I just said, in larger projects this extra step is gonna pay off in terms of an improved user experience.



## What We'll Create
- Agent interface (`internals/config/loader.go`)
- YAML config (`./configs/config.yaml`)




___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "01_interfaces.md" >}})
[|NEXT|]({{< ref "../part_c/01_https_server.md" >}})