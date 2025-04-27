 ---
showTableOfContents: true
title: "Adding Websocket to Client (Lab 07)"
type: "page"
---
## Overview
In this lab we'll also implement Websocket logic on our client-side. Since our client is written in `Vue.js`, we'll write it in JS.


## Implementation: websocket.js

First, create a new directory here - `/ui/src/services.` Then, inside of services, create a file called `websocket.js`, and add the following:

```javascript
// WebSocket connection management service
import { ref } from 'vue';

// Reactive state to track connection status
const isConnected = ref(false);

// WebSocket instance
let socket = null;

// WebSocket server URL
const wsUrl = 'ws://localhost:8080/ws';

// Connect to WebSocket server
function connect() {
    // Don't connect if already connected
    if (socket && (socket.readyState === WebSocket.CONNECTING || 
                   socket.readyState === WebSocket.OPEN)) {
        return;
    }
    
    // Create new WebSocket connection
    socket = new WebSocket(wsUrl);
    
    // Connection opened event
    socket.onopen = () => {
        console.log('WebSocket connected');
        isConnected.value = true;
    };
    
    // Connection closed event
    socket.onclose = () => {
        console.log('WebSocket disconnected');
        isConnected.value = false;
    };
    
    // Connection error event
    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
    };
}

// Start connection when service is imported
connect();

// Export the service
export default {
    isConnected,
    connect
};
```



## Code Breakdown
Given that this is JS, we won't break it down in too much detail, but it is of course worth getting some sense of what's going on here.

The first thing to point out is that we are connecting to the port + endpoint we just defined in Lab 06:
- `const wsUrl = 'ws://localhost:8080/ws';`
- `socket = new WebSocket(wsUrl);`

We only really have one major function here called `connect()`, which I'm assuming you're capable of guessing what it does. We also define `socket.onopen` and `socket.onclose` to print to console, this will allow us to confirm we are able to connect in the browser (i.e. client side).


One thing I want to point out is that we call the function in our script - `connect();`. As the accompanying comment suggests, this means that this function is called the moment it's imported. This beckons the question - where should we import it?

## Tracing Vue.js Execution

If we go to the root of our Vue.js directory, `ui`, you'll see a file there called `index.html`. When we open our UI in our browser, this is actually the only page our application exists of. Hence why Vue.js (at least in this setup) is known as a SPA - Single-Page Application.

If we look inside of it, we can see how it gets away with doing what it does with only one file:
```html
<script type="module" src="/src/main.js"></script>
```

In other words, it imports the JS file `/src/main.js`, and all our logic is contained within it.

So let's have a look to see what's going on inside of it:
```js
import { createApp } from 'vue'
import './style.css'
import App from './App.vue'

createApp(App).mount('#app')
```

Also not much, we have our css, and then we can see it's in turn also just really doing one thing - mounting `App.vue`. We'll touch on what's going on in there and how it relates to our final UI in Lab 09, for now I just wanted to give you a sense of the nested relationship of our application.

In any case, back to `websocket.js` - I said that it will automatically call the `connect()` function the moment it's imported, and so this is where we import it - in main.js. So add this line

```js
import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import './services/websocket.js'

createApp(App).mount('#app')
```

So now - when we load our page `index.html `is rendered, it runs `main.js`, which imports `websocket.js,` which then automatically calls `connect()`.

## Test
First, just be sure that your client application server is running, from `ui` run:
```
npm run dev
```

Open the client in your browser, and then also run the server. You should see it connect over websockets, if not just refresh your browser.

![lab07](../img/lab07a.png)


Additionally, since we defined `socket.onopen` on the client-side implementation, you should see the following message if you open the browser console. Depending on the exact browser you are using, you should find this under something like "Developer's Tools" or "Inspect Element".



![lab07](../img/lab07b.png)






___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "lab06.md" >}})
[|NEXT|]({{< ref "../part_e/lab08.md" >}})