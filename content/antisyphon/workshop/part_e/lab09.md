---
showTableOfContents: true
title: "Client UI Command + Results Handling (Lab 09)"
type: "page"
---

## Overview

It's worth quickly reviewing our initial project overview image to quickly describe what we'll be doing for the remainder of the workshop.

![lab09](../img/lab09a.png)


And so starting at the bottom right, our intention is to be able to at then end of our workshop:
1. Select one of our 3 commands in our Client app (bottom left),
2. Which is then sent to our server and stored in a queue (top left),
3. Allowing our agent to check in to the `/commands` endpoint and retrieve it,
4. Our agent can then run the command and collect the output (top right),
5. The output is sent back to the server (top left),
6. It's then broadcast back to the client (bottom left),
7. After which it's once again displayed to us in the browser (bottom right).

As you can see, both the first (selecting command + client sending it), and the last (client receiving output from server) steps involve our client. Since these are the only remaining sections on Vue.js, I've decided to combine them together in this lab. That way all the JS is out of the way and we can spend the remainder of out time immersed in Go.


## Vue.js Architecture Overview
In Lab 07 I briefly mention how a Vue.js app is put together. I want to review + add to it, but before I do I just also need to be clear: this is ONE way of designing a Vue.js app architecture, but there are many. Specifically, this is the one that's inherited from `Vite`, the tool we used to generate our Vue.js project template.

So when we open our browser, all we really "see" is `index.html`. Now `index.html` itself contains very little, mostly it's a JS script called `main.js` running. When we looked inside of that, we saw that it is essentially just a proxy for our "main" Vue file called `App.vue`.

So when we create a Vue app, it's really all happening in App.vue. However, App.vue is in many ways our high-level "orchestrator". It's kind of like a `main.go` file - it serves as the entry-point, and it's where you express the highest-level logic of your app, but most of the nitty-grotty code actually goes in their own separate files.

And so with Vue, what we most often do is create individual components (inside of `/src/components`), and then import and arrange them in App.vue. Now as the program grows in complexity it gets a little more nuanced than that, but for now this is not a bad way to think of it.

And so the first thing we want to do then is create our components.

## ui/src/components/ConnectionStatus.vue

So create the following file, and then add the code below.

```vue
<template>
  <div class="connection-status">
    <div class="indicator" :class="{ connected: isConnected }"></div>
    <span>{{ isConnected ? 'Connected' : 'Disconnected' }}</span>
  </div>
</template>

<script>
import websocketService from '../services/websocket';

export default {
  name: 'ConnectionStatus',
  setup() {
    return {
      isConnected: websocketService.isConnected
    };
  }
};
</script>

<style scoped>
.connection-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
}

.indicator {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background-color: #ff3b30;
  transition: background-color 0.3s ease;
}

.indicator.connected {
  background-color: #34c759;
}
</style>
```

All this code does is it will give us a visual indicator of our connection status with word Connected/Disconnected + either a green or red circle. Before, we had to jump into the console to do that, well it's obviously not ideal to have to spend the time and effort doing this every time we want to ensure we ware connected, and so though this part is not "required" per se, it's quite handy.


## ui/src/components/CommandController.vue

This is are only other components, so create it and add the following code.

```vue
<template>
  <div class="command-controller">
    <div class="command-buttons">
      <button
          v-for="cmd in commands"
          :key="cmd"
          @click="executeCommand(cmd)"
          :disabled="!isConnected"
      >
        {{ cmd }}
      </button>
    </div>

    <div class="results-panel">
      <div class="results-header">
        <h3>Command Results</h3>
        <button @click="clearResults" class="clear-button">Clear</button>
      </div>

      <div class="results-container">
        <div v-if="messages.length === 0" class="no-results">
          No commands executed yet
        </div>

        <div v-for="(result, index) in messages" :key="index" class="result-item">
          <div class="result-header">
            <div class="command-name">$ {{ result.command }}</div>
            <div class="status-badge" :class="result.status">{{ result.status }}</div>
          </div>
          <pre class="output">{{ result.output }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import websocketService from '../services/websocket';

export default {
  name: 'CommandController',
  setup(props) {
    // Available commands
    const commands = ['pwd', 'whoami', 'hostname'];

    // Execute a command
    function executeCommand(command) {
      websocketService.sendCommand(command);
    }

    // Clear results
    function clearResults() {
      websocketService.messages.value = [];
    }

    return {
      commands,
      executeCommand,
      clearResults,
      isConnected: websocketService.isConnected,
      messages: websocketService.messages
    };
  }
};
</script>

<style scoped>
.command-controller {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.command-buttons {
  display: flex;
  gap: 10px;
}

.command-buttons button {
  padding: 8px 16px;
  border: none;
  background-color: #007aff;
  color: white;
  border-radius: 4px;
  cursor: pointer;
}

.command-buttons button:disabled {
  background-color: #97a5b5;
  cursor: not-allowed;
}

.results-panel {
  border: 1px solid #ddd;
  border-radius: 8px;
  overflow: hidden;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background-color: #3b3b3b;
  border-bottom: 1px solid #ddd;
}

.results-header h3 {
  margin: 0;
  font-size: 16px;
}

.clear-button {
  padding: 4px 8px;
  background: none;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.results-container {
  max-height: 400px;
  overflow-y: auto;
  padding: 12px;
}

.no-results {
  color: #8e8e93;
  font-style: italic;
  text-align: center;
  padding: 20px 0;
}

.result-item {
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #eee;
}

.result-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.command-name {
  font-family: monospace;
  font-weight: bold;
}

.status-badge {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
}



.output {
  background-color: #1d1d1d;
  color: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
  margin: 0;
  overflow-x: auto;
  font-size: 14px;
  line-height: 1.5;
}
</style>
```

It's a lot of code, but as you can see about 75% of it is really just css. Also, just FYI, I recommend perhaps finishing the lab so that you can see what the final visual output is, then coming back and looking at the code + comments below. It's easier to deduce what's going on when you have the final outcome in front of you.

Essentially we visually creating 3 things.

**First, a button:**
```vue
      <button
          v-for="cmd in commands"
          :key="cmd"
          @click="executeCommand(cmd)"
          :disabled="!isConnected"
      >
        {{ cmd }}
      </button>
```

There is a button for each element in commands, which we can see below in the `script` section.
```js
setup(props) {
    // Available commands
    const commands = ['pwd', 'whoami', 'hostname'];

    // Execute a command
    function executeCommand(command) {
      websocketService.sendCommand(command);
    }

```


We can also see above that when we click a button (`@click`), it calls `executeCommand()`, which in turn calls `websocketService.sendCommand`, a function we'll create below to send the command to our server.


**Moving on we also have a results panel:**
```vue
    <div class="results-panel">
      <div class="results-header">
        <h3>Command Results</h3>
        <button @click="clearResults" class="clear-button">Clear</button>
    </div>
```

This just has a simple heading, and a clear button allowing us to clear the output history.


**Finally, we have a results container:**
```vue
      <div class="results-container">
        <div v-if="messages.length === 0" class="no-results">
          No commands executed yet
        </div>

        <div v-for="(result, index) in messages" :key="index" class="result-item">
          <div class="result-header">
            <div class="command-name">$ {{ result.command }}</div>
            <div class="status-badge" :class="result.status">{{ result.status }}</div>
          </div>
          <pre class="output">{{ result.output }}</pre>
        </div>
      </div>
```


This is where the output that is returned by our server in the final step (result.output) will be displayed.


## src/App.vue

In this case of course we don't have to create the file since `Vite` did it for us back in Lab 05. So now, remove all the placeholder content we added in that lab, and replace it with the following code.

```vue
<template>
  <div class="app">
    <header>
      <h1>Orlok C2 Command Center</h1>
      <ConnectionStatus />
    </header>

    <main>
      <CommandController />
    </main>
  </div>
</template>

<script>

import ConnectionStatus from './components/ConnectionStatus.vue';
import CommandController from './components/CommandController.vue';

export default {
  components: {
    ConnectionStatus,
    CommandController
  }
}
</script>

<style>
body {
  background-color: #1a1a1a;
  color: #f0f0f0;
  font-family: sans-serif;
  margin: 0;
  padding: 20px;
}

.app {
  max-width: 800px;
  margin: 0 auto;
}

header {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  margin-bottom: 30px;
}

h1 {
  font-size: 28px;
  margin: 0 0 16px 0;
  color: #e6e978;
}

main {
  margin-top: 20px;
}
</style>
```

Once again, most of it is css, and that which is not is extremely simple - we `import` our two components we created above, and then display them below a heading.

## src/services/websocket.js

The final thing we need to do is make some changes to `websocket.js`. In our current implementation is is capable of creating a connection, but we now also need it to be able to send a command, as well as receive the output.



```js
// WebSocket connection management service
import { ref } from 'vue';

// Reactive state
const isConnected = ref(false);
const messages = ref([]);

// WebSocket instance
let socket = null;

// WebSocket server URL - adjust if needed
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

        // Attempt to reconnect after a delay
        setTimeout(() => {
            console.log('Attempting to reconnect...');
            connect();
        }, 3000);
    };

    // Connection error event
    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        isConnected.value = false;
    };

    // Incoming message event
    socket.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            console.log('Received message:', data);

            // Add message to history if it's a command response
            if (data.type === 'response') {
                messages.value.push(data);
            }
        } catch (error) {
            console.error('Error parsing message:', error);
        }
    };
}

// Send a command to the server
function sendCommand(command) {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.error('Cannot send message, WebSocket not connected');
        return false;
    }

    const message = {
        type: 'command',
        command
    };

    socket.send(JSON.stringify(message));
    return true;
}

// Start connection when service is imported
connect();

// Export the service
export default {
    isConnected,
    messages,
    connect,
    sendCommand
};
```


You'll notice that our ability to receive an incoming message like our results (`socket.onmessage`) is inside of our `connect()` function. It deserializes the message, and if it's a `response`, then it pushed where it is of course eventually displayed in our `CommandController` component.

Our `sendCommand` function is separate, this is of course called from  `CommandController.vue` as we saw. We are creating an object called `message` with 2 fields (the type and the actual command i.e. `whoami` etc.), we serialize it, and then we send it.


So that's it for our client implementation, we can now test it to at least make sure we are able to send a message to our server, we won't find out if our ability to receive + display results work until the very end of course.

## Test
So run the server (`go run ./cmd/server`), ensure the client is running (`npm run dev `inside of `./ui`), and open the client in your browser. You should now see the following.


![lab09](../img/lab09b.png)


It might initially be disconnected, but it should automatically connect within 3 seconds since I've changed the logic inside of `websocket.js` to attempt connection every 3 seconds - see if you can find it.

Now, press any of the commands, for example pwd. Nothing on our front end will change since it's not receiving anything back from the server yet, however if we look inside of our server we'll see that we received a JSON message with our 2 fields - `type` and `command`.



![lab09](../img/lab09c.png)


## Conclusion
So we're now capable of getting our command from our client to our server, let's go it up at this point in our next lab.




___
[|TOC|]({{< ref "../moc.md" >}})
[|PREV|]({{< ref "lab08.md" >}})
[|NEXT|]({{< ref "lab10.md" >}})