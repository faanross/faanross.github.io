---
showTableOfContents: true
title: "Writing and Loading Your First Script"
type: "page"
---


## Let's Get Scripting!

Before we dive into type-specific sections and start writing our first actual script, let's make sure we understand the practical workflow of Zeek script development. 

## Setting Up Your Development Environment

### Choosing an Editor

Zeek scripts are plain text files with a `.zeek` extension (older scripts may use `.bro`). You can write them in any text editor, but some offer better support than others:

**Visual Studio Code (Recommended for beginners)**

- Install the "Zeek Language" extension for syntax highlighting
- Provides reasonable auto-completion and error detection
- Lightweight and cross-platform

**Vim/Neovim (Recommended for terminal users)**

- Install zeek.vim syntax plugin from: https://github.com/zeek/vim-zeek
- Fast, powerful, works over SSH
- Place syntax file in `~/.vim/syntax/zeek.vim`
- Add to `.vimrc`: `au BufRead,BufNewFile *.zeek set filetype=zeek`

**Emacs**

- Zeek mode available, search for "zeek-mode"
- Powerful but steeper learning curve

**Sublime Text / Atom**

- Various community plugins available
- Check package managers for "Zeek" or "Bro"

**Basic syntax highlighting is crucial** - it helps you spot typos, makes code more readable, and catches obvious errors before you even load the script.

### Script Linting and Validation

Before loading a script into Zeek, you can check it for syntax errors:

```bash
# Check script syntax without running it
zeek -a your-script.zeek

```

The `-a` flag performs a full parse without execution. This catches syntax errors, type mismatches, and undefined variables before you load the script into your running Zeek instance.

**Common errors you'll catch:**

- Missing semicolons
- Undeclared variables
- Type mismatches (adding a string to a count, etc.)
- Malformed regular expressions
- Incorrect function signatures

Get in the habit of running `zeek -a` before deploying any script. It saves time and prevents broken deployments.


## Where Scripts Live: The Zeek Directory Structure

Understanding where scripts go and how Zeek finds them is essential:

```
/opt/zeek/                          # Main Zeek installation
├── bin/                            # Zeek executables (zeek, zeekctl)
├── share/zeek/                     # Zeek's script library
│   ├── base/                       # Core scripts (protocols, frameworks)
│   ├── policy/                     # Optional detection scripts
│   └── site/                       # YOUR CUSTOM SCRIPTS GO HERE
│       ├── local.zeek              # Main site configuration
│       └── custom/                 # Organize your scripts here
└── logs/                           # Log output (when running)
```

**Key principle: Put your custom scripts in `/opt/zeek/share/zeek/site/` or subdirectories within it.**

This keeps your code separate from Zeek's built-in scripts, making updates easier and organization clearer.

### Creating Your First Script

Let's create a simple "hello world" script to verify the workflow:

```bash
# Navigate to the site directory
cd /opt/zeek/share/zeek/site/

# Create a directory for your custom scripts (optional but recommended)
sudo mkdir -p custom

# Create your first script
sudo nano custom/hello.zeek
```

In the editor, write:

```c
# hello.zeek - Your first Zeek script
# This script proves your workflow is working

event zeek_init()
{
    print "Hello from Zeek! Script loaded successfully.";
    print fmt("Zeek started at: %s", network_time());
}

event zeek_done()
{
    print "Zeek is shutting down. Goodbye!";
}
```

**Understanding this script:**

- `zeek_init()` fires once when Zeek starts - perfect for initialization and testing
- `zeek_done()` fires once when Zeek shuts down
- `print` outputs to stdout (visible in foreground mode) or logs
- `fmt()` formats strings like printf

Save the file (Ctrl+O, Enter, Ctrl+X in nano).


Let's now also test to ensure our script works:
```bash
zeek -a custom/hello.zeek
```

In this case - no output is good. If we had some error - try it yourself by introducing an intentional mangle - you'll see output related to it:

```bash
zeek -a custom/hello.zeek
error in ./custom/hello.zeek, line 13: syntax error, at or near "}"
```



## Loading Scripts: The local.zeek Method

As we touched on in Module 01, `local.zeek` file is our **site configuration hub**. It's loaded automatically by Zeek and is where you specify which scripts to load, set configuration options, and customize behaviour.

Edit `local.zeek`:

```bash
sudo nano /opt/zeek/share/zeek/site/local.zeek
```

As we saw before, already has content (network definitions, loaded scripts, etc.). It would be worth spending some time
at some point reviewing all the content to make sure you understand what everything does, and in general how the
script is put together and operates.

**For now however, let's import our new script:**


```c
# Load our custom hello world script
@load ./custom/hello.zeek
```

The `@load` directive tells Zeek to load and execute the specified script. The path `./custom/hello.zeek` is relative to the `site/` directory.

**Alternative loading methods:**

```c
# Load by absolute path (less common)
@load /opt/zeek/share/zeek/site/custom/hello.zeek

# Load entire directory (loads all .zeek files in it)
@load ./custom/

# Load with namespace (for organization)
@load custom/detection/scanning
```

## Testing Your Script

Now let's verify everything works:

### Method 1: Test in Foreground 

Use this to run an instance of Zeek to test it, which is recommended for R&D purposes.

```bash
# Run Zeek in foreground on a network interface
sudo zeek -i eth0 local.zeek

# Or run on a PCAP file for testing
sudo zeek -r /path/to/capture.pcap local.zeek
```

Once it's up and running you can hit Ctrl + C to stop. You should immediately see:

```bash
sudo zeek -i eth0 local.zeek
listening on eth0

Hello from Zeek! Script loaded successfully.
Zeek started at: 0.0

1760440271.105128 115 packets received on interface eth0, 0 (0.00%) dropped, 0 (0.00%) not processed
...
Zeek is shutting down. Goodbye!
```

We can see both the output when the script starts, and the output when Zeek shuts down.



### Method 2: Deploy with ZeekControl (Production method)

For running Zeek as a service (which you learned in the previous section), there's an important difference: **`print` statements don't appear when Zeek runs as a daemon**. 

This is of course because daemons run in the background, so don't have the ability to output to the active terminal.

So instead, we need a script that writes to Zeek's logging system. Let's create a daemon-friendly version:

```bash
# Create a new script for daemon mode
sudo nano /opt/zeek/share/zeek/site/custom/hello-daemon.zeek
```

```c
# hello-daemon.zeek

module HelloDaemon;

export {
    redef enum Log::ID += { LOG };
    
    type Info: record {
        ts: time &log;
        message: string &log;
    };
}

event zeek_init()
{
    Log::create_stream(HelloDaemon::LOG, [$columns=Info, $path="hello"]);
    
    Log::write(HelloDaemon::LOG, [
        $ts=network_time(),
        $message="Hello from Zeek daemon! Script loaded successfully."
    ]);
}

event zeek_done()
{
    Log::write(HelloDaemon::LOG, [
        $ts=network_time(),
        $message="Zeek daemon is shutting down. Goodbye!"
    ]);
}
```


**Test to make sure there are no errors:**

Get in the habit of doing this and develop the muscle memory - ALWAYS run `zeek -a` against a newly minted script.

```
zeek -a custom/hello-daemon.zeek
```


**Load it in local.zeek:**

```bash
sudo nano /opt/zeek/share/zeek/site/local.zeek
```

Comment out the old one and add the new one at the bottom:

```c
# Custom script to test
# @load ./custom/hello.zeek
@load ./custom/hello-daemon.zeek
```

**Now deploy with zeekctl:**

```bash
# Check configuration is valid
sudo zeekctl check

# If no errors, install the new configuration
sudo zeekctl install

# Restart Zeek to load new scripts
sudo zeekctl restart

# Check that Zeek is running
sudo zeekctl status
```

**Verify the script loaded and is working:**

```bash
cat /opt/zeek/logs/current/hello.log
```

**You should see output like**:
```bash
#separator \x09
#set_separator	,
#empty_field	(empty)
#unset_field	-
#path	hello
#open	2024-10-14-07-13-05
#fields	ts	message
#types	time	string
1728891185.123456	Hello from Zeek daemon! Script loaded successfully.
```


**To see the shutdown message**: 
Stop Zeek and check the log, remember Zeek will archive the log from the `current` directory to the timestamped directory, since mine will be different than yours you cannot blindly C+P this command - find your specific log.

```bash
sudo zeekctl stop

# using zcat since files are gzipped
zcat /opt/zeek/logs/2025-10-14/hello.07\:27\:23-07\:27\:24.log.gz
#separator \x09
#set_separator	,
#empty_field	(empty)
#unset_field	-
#path	hello
#open	2025-10-14-07-27-23
#fields	ts	message
#types	time	string
1760441243.306404	Zeek daemon is shutting down. Goodbye!
#close	2025-10-14-07-27-23
```

And now if you'd like, just remember to start Zeek again using `sudo zeekctl start`.



## Development Workflow: The Iteration Cycle

Here's a quick recap of the overall workflow we'll be using constantly:

**1. Write/Edit Script**

```bash
sudo nano /opt/zeek/share/zeek/site/custom/your-script.zeek
```

**2. Check Syntax**

```bash
cd /opt/zeek/share/zeek/site/
zeek -a custom/your-script.zeek
```

**3. Test in Foreground (Fast iteration)**

```bash
# Run on live interface
sudo zeek -i eth0 local.zeek

# Or run on test PCAP
sudo zeek -r test.pcap local.zeek
```

**4. Generate Test Traffic**

```bash
# In another terminal, trigger your detection
curl http://example.com
# or whatever triggers your script
```

**5. Observe Output**

- Watch stdout in the terminal where Zeek is running
- Check logs in current directory: `ls *.log`

**6. Fix/Improve Script** - Go back to step 1

**7. Deploy to Production (when satisfied)**

```bash
sudo zeekctl check
sudo zeekctl install
sudo zeekctl restart
```

## Common Pitfalls and Solutions

**Problem: "error loading script" or "file not found"**

- Check the path in your `@load` directive
- Verify the file exists: `ls -la /opt/zeek/share/zeek/site/custom/`
- Make sure path is relative to `site/` directory
- Check for typos in filename

**Problem: Script loads but doesn't seem to do anything**

- Add `print` statements to verify script is executing
- Check you're monitoring the right interface: `ip addr` to list interfaces
- Verify traffic is flowing: `sudo tcpdump -i eth0 -c 10`
- For event handlers, make sure events are actually firing

**Problem: Syntax error messages**

- Read the error carefully - Zeek tells you line number and what's wrong
- Common issues: missing semicolons, incorrect types, typos in variable names
- Use `zeek -a` before loading to catch errors early

**Problem: Changes don't appear after editing**

- With zeekctl: Did you run `zeekctl install` and `zeekctl restart`?
- With foreground: Did you stop and restart Zeek?
- Zeek doesn't auto-reload - you must restart after changes

## Best Practices for Script Development

**1. Start Simple**

- Write minimal code first (just print statements)
- Verify it loads and runs
- Add complexity incrementally

**2. Use Descriptive Names**

```c
# Good
local suspicious_connection_count: count = 0;

# Bad
local x: count = 0;
```

**3. Comment Your Code**

```c
# Track failed SSH attempts per IP
# Alert when threshold exceeded
global ssh_failures: table[addr] of count;
```

**4. Test with Known Data**

- Keep test PCAP files for common scenarios
- Generate controlled test traffic
- Verify detections fire when they should (and don't when they shouldn't)

**5. Log Your Detections**

```c
# Don't just print - write to notice framework or custom logs
# We'll cover this more later, but for now:
print fmt("DETECTION: Suspicious activity from %s", ip);
```

**6. Organize Your Scripts**

```
site/
├── local.zeek              # Main config
└── custom/
    ├── scanning/           # Scan detection scripts
    ├── bruteforce/         # Brute force detection
    ├── exfiltration/       # Data exfil detection
    └── util/               # Helper functions
```

**7. Version Control Your Scripts**

```c
cd /opt/zeek/share/zeek/site/custom/
git init
git add .
git commit -m "Initial detection scripts"
```

This lets you track changes, roll back mistakes, and collaborate.

## Verifying Your Setup

Before moving to the type-specific exercises, verify your environment:

```c
# 1. Check Zeek is installed and version
zeek --version

# 2. Check you can write to site directory
sudo touch /opt/zeek/share/zeek/site/custom/test.txt
sudo rm /opt/zeek/share/zeek/site/custom/test.txt

# 3. Verify your hello.zeek loads without errors
cd /opt/zeek/share/zeek/site/
zeek -a custom/hello.zeek

# 4. Test it runs
sudo zeek -i eth0 local.zeek
# Press Ctrl+C after seeing hello message

# 5. Check zeekctl works
sudo zeekctl status
```

If all of these work, you're ready to start building real detections!

**LET'S DO IT!**


---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./bool.md" >}})
[|NEXT|]({{< ref "./exercise01.md" >}})

