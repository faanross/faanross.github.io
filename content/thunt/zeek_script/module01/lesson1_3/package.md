---
showTableOfContents: true
title: "Part 2 - Installation Method 1: Package Manager"
type: "page"
---



## **PART 2: INSTALLATION METHOD 1 - PACKAGE MANAGER (RECOMMENDED FOR BEGINNERS)**

### **Understanding Package Manager Installation**

The simplest way to install Zeek is using Ubuntu's package manager, `apt`. Package manager installation has several advantages that make it ideal for beginners and for production deployments where standardization is important.

**Advantages of package installation:**

```
┌──────────────────────────────────────────────────────────────┐
│        PACKAGE MANAGER INSTALLATION BENEFITS                 │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ✓ SIMPLICITY                                                │
│    Single command installation, handles all dependencies     │
│                                                              │
│  ✓ INTEGRATION                                               │
│    Properly integrated with Ubuntu's system (systemd, etc.)  │
│                                                              │
│  ✓ UPDATES                                                   │
│    Security updates via normal system update process         │
│                                                              │
│  ✓ RELIABILITY                                               │
│    Pre-tested binaries known to work on Ubuntu 24.04         │
│                                                              │
│  ✓ UNINSTALL                                                 │
│    Clean removal if needed (apt remove zeek)                 │
│                                                              │
│  ✗ LIMITATIONS                                               │
│    • Slightly older version than latest source               │
│    • Less control over compilation options                   │
│    • Can't customize build flags for optimization            │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

For most users, especially those new to Zeek, package installation is the right choice. You'll get Zeek running quickly and can focus on learning to use it rather than troubleshooting build issues.

### **Step-by-Step Package Installation**

Let's install Zeek from the official package repository. Zeek maintains packages for various Linux distributions, including Ubuntu.

**Step 1: Add the Zeek Package Repository**

Ubuntu's default repositories include an older version of Zeek. To get a more recent version, we'll add Zeek's official repository:

```bash
# Install prerequisites for adding repositories
sudo apt install -y curl gnupg2

# Add Zeek's GPG key (verifies package authenticity)
curl -fsSL https://download.opensuse.org/repositories/security:zeek/xUbuntu_24.04/Release.key | \
    sudo gpg --dearmor -o /usr/share/keyrings/zeek-archive-keyring.gpg

# Add the repository to your sources list
echo "deb [signed-by=/usr/share/keyrings/zeek-archive-keyring.gpg] \
    http://download.opensuse.org/repositories/security:/zeek/xUbuntu_24.04/ /" | \
    sudo tee /etc/apt/sources.list.d/zeek.list

# Update package index to include the new repository
sudo apt update
```

**Understanding what just happened:**

The GPG key ensures that packages you download are authentic and haven't been tampered with. Think of it like a signature that verifies the package came from the Zeek project. The repository URL tells Ubuntu where to find Zeek packages. By adding this to your sources list, you've told apt to check this repository when looking for packages.

**Step 2: Install Zeek**

Now we can install Zeek with a single command:

```bash
sudo apt install -y zeek
```

This will download and install Zeek along with all its dependencies. You'll see a lot of packages being installed - these are libraries and tools that Zeek needs to function. The installation might take 5-10 minutes depending on your connection speed.

Watch the output as packages install. You should see messages like:

```
Reading package lists... Done
Building dependency tree... Done
The following NEW packages will be installed:
  zeek zeek-core zeek-core-dev libbroker-dev ...
...
Setting up zeek (8.0.1-0) ...
```

Note - the version number might be different depending on when you're reading this.

**Step 3: Verify the Installation**

Let's confirm Zeek installed correctly:

```bash
# Check Zeek version
/opt/zeek/bin/zeek --version
```

You should see output like:

```
zeek version 8.0.1
```

If you see this, congratulations! Zeek is installed. But we're not done yet - we need to configure it.

**Understanding the installation location:**

Package installation places Zeek in `/opt/zeek/`. Let's explore this directory structure:

```bash
ls -la /opt/zeek/
```





**Zeek directory structure:**

```
/opt/zeek/
├── bin/                    # Executables
│   ├── zeek                # Main Zeek binary
│   ├── zeek-cut            # Log parsing tool
│   ├── zeekctl             # Cluster management tool
│   └── zkg                 # Package manager
│
├── etc/                    # Configuration files
│   ├── node.cfg            # Node/cluster configuration
│   ├── networks.cfg        # Network definitions
│   └── zeekctl.cfg         # ZeekControl settings
│
├── include/                # Header files (for development)
│
├── lib/                    # Libraries
│
├── logs/                   # Log output (created on first run)
│   └── current/            # Current log files
│
├── share/zeek/             # Zeek scripts and policies
│   ├── base/               # Core functionality scripts
│   ├── policy/             # Optional policy scripts
│   └── site/               # Local site customizations
│       └── local.zeek      # Main configuration script
│
└── var/                    # Runtime data
    └── lib/zeek/           # State files

```

This directory structure is important - you'll be working with these directories throughout the course. Let's understand the key locations:

| Directory                    | Purpose        | When You'll Use It                               |
| ---------------------------- | -------------- | ------------------------------------------------ |
| `/opt/zeek/bin/`             | Executables    | Every time you run Zeek commands                 |
| `/opt/zeek/etc/`             | Configuration  | When setting up monitoring, defining networks    |
| `/opt/zeek/share/zeek/site/` | Custom scripts | When writing detection logic (most of your work) |
| `/opt/zeek/logs/`            | Log output     | When analyzing captured traffic                  |
| `/opt/zeek/var/`             | Runtime state  | When troubleshooting issues                      |

### **Setting Up Your PATH**

Typing `/opt/zeek/bin/zeek` every time you want to run a Zeek command is tedious. Let's add Zeek's bin directory to your PATH so you can just type `zeek`:

```bash
# Add Zeek to PATH in your .bashrc
echo 'export PATH=/opt/zeek/bin:$PATH' >> ~/.bashrc

# Reload your .bashrc to apply changes
source ~/.bashrc

# Test it
zeek --version
```

Now you should be able to run `zeek` from anywhere without specifying the full path. This is a small convenience that makes daily Zeek operations much more pleasant.

**For the root user too:**

Since you'll sometimes need to run Zeek commands with sudo, let's add it to root's PATH as well:

```bash
sudo bash -c 'echo "export PATH=/opt/zeek/bin:\$PATH" >> /root/.bashrc'
```

### **Installing ZeekControl**

ZeekControl is a shell script that makes managing Zeek installations easier. It provides commands to start, stop, and monitor Zeek. While you can run Zeek directly, ZeekControl simplifies operations, especially for cluster deployments.

The package installation should have included ZeekControl, but let's verify:

```bash
zeekctl
```

You should see the ZeekControl prompt:

```
Hint: Run the zeekctl "deploy" command to get started.

Welcome to ZeekControl 2.6.0-28

Type "help" for help.

[ZeekControl] >
```

Type `exit` to leave the ZeekControl shell for now. We'll return to it shortly.




---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./prepare.md" >}})
[|NEXT|]({{< ref "./compile.md" >}})

