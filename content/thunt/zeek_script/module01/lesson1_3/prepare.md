---
showTableOfContents: true
title: "Part 1 - Preparing Your Installation Environment"
type: "page"
---
## **From Theory to Reality: Building Your First Zeek Sensor**

Welcome to what many students consider the most exciting lesson in Module 1 - the moment where abstract concepts become a working system. Over the next few hours, you're going to build a fully functional Zeek sensor from the ground up. You'll install the software, configure it for optimal performance, and capture your first network traffic. By the end of this lesson, you'll have a working platform for all the threat hunting and detection work that follows in later modules.

This lesson is different from the previous two. While Lessons 1.1 and 1.2 were primarily conceptual, this lesson is intensely practical. You'll be working in a terminal, executing commands, editing configuration files, and troubleshooting issues. This hands-on experience is crucial because Zeek isn't something you learn purely from reading - you need to feel how it works, see it in action, and develop the muscle memory that comes from actually operating the system.

We're going to take a methodical approach to installation. Rather than just running through a quick install script, we'll explore three different installation methods - package manager installation, compilation from source, and container deployment. Each method has different characteristics and use cases, and understanding all three will make you a more versatile Zeek operator. We'll examine Zeek's file system layout in detail so you understand where everything lives and why. We'll configure network interfaces for monitoring, set up high-performance packet capture, and implement best practices for production deployments.

This is a long lesson with many steps. Take your time, read carefully, and don't skip ahead. Each section builds on the previous one, and understanding why we're doing something is just as important as knowing how to do it.

Let's begin.

---


## **PART 1: PREPARING YOUR INSTALLATION ENVIRONMENT**

### **Understanding Your Installation Environment**

Before we install anything, let's understand the environment we're working with and prepare it properly. You can run Zeek on any platform that suits your needs: a cloud VM from any provider (Digital Ocean, AWS, Azure, etc.), a Type 1 or Type 2 hypervisor, or even a bare metal Linux installation. Use whatever you have available.

For this course, I'll be using a Digital Ocean droplet, but the principles apply regardless of where you're running your system. You should have a system with at least 4 vCPUs, 8GB RAM, and 80GB of storage. If you're using a cloud VM and haven't provisioned it yet, pause here and create it now.

Zeek runs on numerous Linux distributions, including Ubuntu, Red Hat, Rocky, and others. I'll be using Ubuntu 24.04 LTS in all examples because, in my opinion, it's currently the most stable OS for Zeek and will be supported until 2029. It strikes the right balance between modern software packages and being a long-term support release that won't surprise you with breaking changes.

You're welcome to use other versions like 22.04, or different distributions entirely, but be aware that if you deviate from Ubuntu 24.04, you may need to fill in any gaps or handle distribution-specific differences that arise during the course.


**Why Ubuntu 22.04 LTS specifically?**

```
┌──────────────────────────────────────────────────────────────┐
│         UBUNTU 24.04 LTS ADVANTAGES FOR ZEEK                 │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ✓ Long-term support until 2029                              │
│    (No need to upgrade for 3.5 years)                        │
│                                                              │
│  ✓ Modern kernel with AF_PACKET improvements                 │
│    (Better packet capture performance)                       │
│                                                              │
│  ✓ Recent compiler versions                                  │
│    (GCC 11+, needed for optimal Zeek compilation)            │
│                                                              │
│  ✓ Large community and extensive documentation               │
│    (Easy to find solutions to problems)                      │
│                                                              │
│  ✓ Official Zeek packages available                          │
│    (Can install via apt without compilation)                 │
│                                                              │
│  ✓ SystemD for service management                            │
│    (Modern, reliable daemon management)                      │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```


### **Initial System Setup and Security Hardening**

I'm going to assume you're logging into a fresh Ubuntu installation at this point. If you are not and your system is already set up and hardened, you are of course welcome to skip over any superfluous steps.

Before we install Zeek, we need to prepare the system properly. This isn't just about getting dependencies installed - it's about creating a secure, well-configured foundation that will support your security monitoring infrastructure.

**Step 1: Connect to Your VM**

Open your terminal and connect via SSH.

```bash
ssh root@YOUR_VM_IP
```

You'll be prompted to accept the server's SSH fingerprint the first time you connect. Type `yes` and press Enter. You should see a welcome message and a command prompt that looks like:

```
root@zeek-sensor:~#
```

**Step 2: Update the System**

The first thing you should always do on a new Linux system is update all packages to their latest versions. This ensures you have the latest security patches and bug fixes:

```bash
# Update the package index
apt update

# Upgrade all installed packages
apt upgrade -y

# Remove any packages that are no longer needed
apt autoremove -y
```

The `-y` flag automatically answers "yes" to any prompts, which is fine for a fresh system. The update process might take a few minutes depending on how many packages need updating. You'll see a lot of output scrolling by - this is normal.

**Understanding what just happened:**

The `apt update` command refreshes Ubuntu's list of available packages and their versions. Think of it like getting the latest catalog from a store. The `apt upgrade` command actually downloads and installs newer versions of packages that are already installed.

**Step 3: Set the Hostname**

Your server needs a meaningful hostname. Let's set it to something that describes its purpose:

```bash
# Set the hostname
hostnamectl set-hostname zeek-sensor

# Update the hosts file to match
echo "127.0.1.1 zeek-sensor" >> /etc/hosts

# Verify the change
hostname
```

You should see `zeek-sensor` printed. This hostname will appear in your command prompt and in various system logs, making it easier to identify your system.

**Step 4: Configure the Timezone**

Zeek's logs include timestamps, and you want these to be in a timezone that makes sense for your operations. Let's set it now:

```bash
# See available timezones
timedatectl list-timezones | grep -i america  # or Europe, Asia, etc.

# Set your timezone (example: Eastern Time)
timedatectl set-timezone America/New_York

# Verify
timedatectl
```

This ensures that when Zeek logs show "14:32:18", you know exactly what time that was in your local context. For a distributed team, you might prefer UTC, which avoids timezone confusion:

```bash
timedatectl set-timezone UTC
```

**Step 5: Create a Non-Root User**

Running everything as root is dangerous. A mistake or compromised script could destroy your entire system. Let's create a dedicated user for Zeek operations:

```bash
# Create a new user
adduser zeek

# Add the user to the sudo group (for administrative tasks when needed)
usermod -aG sudo zeek
```

You'll be prompted to set a password for the zeek user and provide some optional information (full name, etc.). The password is important - make it strong.

If you want to be able to SSH directly into the VM using the new user and private key (STRONGLY RECOMMENDED) you'll need to also perform the following step:

```shell
rsync --archive --chown=zeek:zeek ~/.ssh /home/zeek
```

This command copies your root user's SSH key configuration to the new user's home directory and sets the correct ownership, allowing you to log in as zeek with the same SSH key.

Now, we want to remove the ability to SSH in as root since this is a security liability, however before you do this I always recommend exiting and making sure you are able to log-in as the new user, otherwise you can get locked out completely.


```shell
# exits VM session
exit

# SSH in as new user
ssh zeek@YOUR_VM_IP
```

You should now log in as zeek and see the following:
```shell
zeek@zeek-sensor:~$
```

Notice the `$` instead of `#`. The dollar sign indicates you're a regular user, while the hash mark indicated root access. For the rest of this lesson, we'll work as the zeek user, using `sudo` when we need administrative privileges.

If for any reason this failed, log-in as `root` again and perform the steps above again, making sure you do everything exactly as is outlined there.

While here, let's also confirm that we do indeed have `sudo` rights:
```shell
sudo whoami
```

Answer should be `root`


Now finally we want to disable the ability to log-in as `root`. Additionally, I will disable password-based logins - it is MUCH more secure to login using a private key. If you are not currently doing this and you are using a VM, I strongly urge you to find any online guide (there are dozens) and set this up first before continuing.

Let's first open the SSH config, I will use `nano`, you are of course free to use `vi`/`vim` if you prefer:

```shell
sudo nano /etc/ssh/sshd_config
```

Find the line that says:
```shell
PermitRootLogin yes
```

And change it to:
```shell
PermitRootLogin no
```

Now find this line further down:
```shell
#PasswordAuthentication yes
```

And both uncomment + change to no

```shell
PasswordAuthentication no
```

Save the changes and exit the editor.

Finally let's restart the service for these changes to take place:

```shell
sudo systemctl restart ssh
```


### **Understanding Network Interface Configuration**

Before we install Zeek, we need to understand your network interfaces. Zeek captures packets from a network interface, so knowing what interfaces you have and how they're configured is essential.

**List your network interfaces:**

```shell
ip link show
```

You'll see output similar to this:

```
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN mode DEFAULT group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP mode DEFAULT group default qlen 1000
    link/ether 12:34:56:78:9a:bc brd ff:ff:ff:ff:ff:ff
3: eth1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP mode DEFAULT group default qlen 1000
    link/ether e6:a3:0a:eb:a4:bc brd ff:ff:ff:ff:ff:ff
```

Let's understand what each interface does:

```
┌──────────────────────────────────────────────────────────────┐
│                    NETWORK INTERFACES                        │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  lo (Loopback Interface)                                     │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│  • Virtual interface for local communication                 │
│  • Traffic: 127.0.0.1 to 127.0.0.1                           │
│  • Purpose: Inter-process communication on same machine      │
│  • Monitoring value: None (don't capture from lo)            │
│                                                              │
│  eth0 (Public Network Interface)                             │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│  • Primary external network interface                        │
│  • Connected to the internet                                 │
│  • Traffic: Internet-facing communications                   │
│  • Monitoring value: Primary focus for this course           │
│                                                              │
│  eth1 (Private Network Interface)                            │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│  • Private/internal network interface                        │
│  • Connected to local network                                │
│  • Traffic: Internal/private communications                  │
│  • Monitoring value: Not our focus for now, but useful       │
│    for detecting lateral movement and east-west traffic      │
│    in more advanced investigations                           │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

For this course, we'll primarily monitor **eth0** to capture internet-facing traffic. The eth1 interface handles private network traffic, which becomes important in advanced scenarios where you're investigating internal threats, lateral movement, or east-west traffic patterns within your infrastructure.




**Check your IP configuration:**

```bash
ip addr show eth0
```

Output will look like:

```
2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP group default qlen 1000
    link/ether 12:34:56:78:9a:bc brd ff:ff:ff:ff:ff:ff
    inet 167.99.123.45/20 brd 167.99.127.255 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fe80::1034:56ff:fe78:9abc/64 scope link 
       valid_lft forever preferred_lft forever
```

This tells you:

- **MAC address**: `12:34:56:78:9a:bc` (Layer 2 hardware address)
- **IPv4 address**: `167.99.123.45` with a /20 subnet mask
- **IPv6 address**: A link-local address starting with `fe80::`
- **State**: UP (interface is active and operational)

### **Understanding the Dual-NIC Concept (Theory)**

In an ideal production deployment, you would have two network interfaces:

```
┌──────────────────────────────────────────────────────────────┐
│              IDEAL DUAL-NIC CONFIGURATION                    │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│                    Your Zeek Sensor                          │
│                                                              │
│       ┌──────────────────────────────┐                       │
│       │                              │                       │
│       │  ┌──────────┐  ┌──────────┐  │                       │
│       │  │Management│  │Monitoring│  │                       │
│       │  │   NIC    │  │   NIC    │  │                       │
│       │  │  (eth0)  │  │  (eth2)  │  │                       │
│       │  └────┬─────┘  └────┬─────┘  │                       │
│       └───────┼─────────────┼────────┘                       │
│               │             │                                │
│               │             │                                │
│               ▼             ▼                                │
│         SSH Access    SPAN/TAP Port                          │
│         Administration   (Monitoring)                        │
│                                                              │
│  Management NIC (eth0):                                      │
│  • Has IP address                                            │
│  • Used for SSH, logging output, management                  │
│  • Generates administrative traffic                          │
│  • Connected to management network                           │
│                                                              │
│  Monitoring NIC (eth2):                                      │
│  • NO IP address (promiscuous mode)                          │
│  • Receives copy of monitored traffic (SPAN/mirror)          │
│  • Never transmits (receive-only)                            │
│  • Connected to monitoring network                           │
│                                                              │
│  WHY SEPARATE?                                               │
│  • Prevents sensor traffic from being captured               │
│  • Improves security (monitoring NIC can't be attacked)      │
│  • Cleaner separation of concerns                            │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

**However**, since most cloud VMs only one network interface, this is what I'll be demonstrating. This is fine for learning and even for many production scenarios where you're monitoring the same network you're connected to. The important thing is understanding the concept so you can implement dual-NIC configurations when appropriate.

For our purposes, we'll capture traffic from eth0, which means Zeek will see:

- Traffic generated by your VM (SSH connections, package downloads, etc.)
- Traffic from any services you run on the VM
- Any traffic you generate for testing purposes

This is perfect for learning. In a later exercise, we'll generate test traffic and watch Zeek capture and analyze it.

---


[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "../lesson1_2/validation.md" >}})
[|NEXT|]({{< ref "./package.md" >}})

