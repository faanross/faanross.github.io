---
showTableOfContents: true
title: "Minimal Client UI Setup with Vue.js (Lab 05)"
type: "page"
---

## Overview
We'll now do a very basic setup of our client (frontend UI). It won't do much at this point, but we just want a basic 
template in place so that we can start connecting everything together.

## npm
The first thing you'll need is `npm` - the Node Package Manager. For installation instructions see [this link](https://nodejs.org/en/download).
You can think of `npm` like `pip` for Python or `brew` for macOS; 
it's the standard tool for downloading and managing the reusable code libraries (packages) that JavaScript projects, 
including Vue.js itself and its related tools, depend on. 



## Initialize Our Project with Vite
First, let's move into our project's root directory, then run:

```
npm create vite@latest ui -- --template vue
```

We are using `npm` to create a new Vue.js project template called ``ui``, using a specific build tool called `Vite`.

`Vite` is a build tool for a variety of JS-frameworks, and it's going to generate a new project structure with the Vue template option.
So it will set up all the necessary folders, files, and configurations. But another awesome thing it does, it provides us with a local 
development server we can use for our development stage to view a live version of our project on our local host. 

Once we've run the command we'll now have a new folder called ui - this is where all files related to our frontend will 
be found. Cool things is the command's output will actually tell you the exact next commands you should run.

```shell
cd ui
npm install
npm run dev
```

So we `cd` into our new `ui` folder, we install all the dependencies for our project, and then after we run the 
final command you'll see a message that our development server is now live on `localhost:5173`. If you now go
there you'll see we already have a development server live with a standard template image.


![lab05](../img/lab05a.png)

## Minor Edits

For the remainder of this lab I'm just going to make a few minor tweaks to this set up. Nothing is critical, and 
you can feel free to ignore everything, it's just my personal preference of how I like things to be set up.

First, in `index.html`, I just want to change the title to the name of my framework.

```html
<title>OrlokC2</title>
```

Then, I'll also replace the contents of `style.css` with my own preference. Again, this completely cosmetic and not
required, I'm just lowkey obsessed with Dracula color theme and Jetbrains font, so I use it wherever possible (you 
may have noticed). 

```css
@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap');

:root {
  font-family: 'JetBrains Mono', system-ui, Avenir, Helvetica, Arial, sans-serif;
  line-height: 1.5;
  font-weight: 400;

  color-scheme: dark;
  color: #fffbfb; /* Dracula foreground */
  background-color: #232323; /* Dracula background */

  font-synthesis: none;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

a {
  font-weight: 500;
  color: #bd93f9; /* Dracula purple */
  text-decoration: inherit;
}
a:hover {
  color: #ff79c6; /* Dracula pink */
}

body {
  margin: 0;
  display: flex;
  place-items: start;
  min-width: 320px;
  min-height: 100vh;
  background-color: #212123; /* Dracula background */
}

h1 {
  font-size: 3.2em;
  line-height: 1.1;
  color: #f1fa8c; /* Dracula yellow */
}

button {
  border-radius: 8px;
  border: 1px solid transparent;
  padding: 0.6em 1.2em;
  font-size: 1em;
  font-weight: 500;
  font-family: inherit;
  background-color: #44475a; /* Dracula selection background */
  color: #f8f8f2; /* Dracula foreground */
  cursor: pointer;
  transition: border-color 0.25s, background-color 0.25s;
}
button:hover {
  border-color: #ff79c6; /* Dracula pink */
  background-color: #6272a4; /* Dracula comment */
}
button:focus,
button:focus-visible {
  outline: 2px solid #8be9fd; /* Dracula cyan */
}

.card {
  padding: 2em;
  background-color: #44475a; /* Dracula selection */
  border-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
}

#app {
  max-width: 1280px;
  margin: 0 auto;
  padding: 2rem;
  text-align: center;
  color: #f8f8f2; /* Dracula foreground */
}
```

Then in `src/components` I'll remove `HelloWorld.vue` since we won't be using it. Your IDE might whine about this since
it's currently being rendered on the site, but don't worry about it we'll now also change our `App.vue`.

So remove it's current contents, and replace with something like this (change text to whatever you like).

```vue
<template>
  <div class="app">
    <h1>Hi, I'm Mister Derp!</h1>
  </div>
</template>

<style>
.app {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}

h1 {
  text-align: center;
}
</style>
```

This just displays a simple placeholder message. 

![lab05](../img/lab05b.png)

## Conclusion

And that's it, we have a simple little placeholder page which does not do much, but at least our client is ready to
be integrated into our overall project.

Our issue now is that our client just exists in isolation - it's not connected to our server. And so in our next 2 
labs we'll add websocket components to both the server (Lab 06) and client (Lab 07) sides, which will allow them to
communicate to one another.

Let's go!

___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "../part_c/lab04.md" >}})
[|NEXT|]({{< ref "lab06.md" >}})