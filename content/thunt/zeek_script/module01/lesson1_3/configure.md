---
showTableOfContents: true
title: "Part 5 - Configuration + Initial Setup"
type: "page"
---



## **PART 5: CONFIGURATION AND INITIAL SETUP**

Now that Zeek is installed (via whichever method you chose), let's configure it for use. Configuration is where you tell Zeek what to monitor, how to behave, and what policies to enforce.

### **Understanding Zeek's Configuration Files**

Zeek's configuration is split across several files, each serving a specific purpose. Let's explore them:

```bash
cd /opt/zeek/etc
ls -la
```

You'll see:

```
node.cfg          # Node/cluster configuration
networks.cfg      # Network definitions
zeekctl.cfg       # ZeekControl settings
```

Let's understand each file:

#### **1. node.cfg - Defining Your Zeek Nodes**

This file defines what Zeek nodes you're running. For a standalone installation, it's simple:

```bash
sudo nano /opt/zeek/etc/node.cfg
```

You'll see a default configuration:

```ini
# node.cfg - Standalone Zeek configuration

# This is a complete standalone configuration.  Most likely you will
# only need to change the interface.
[zeek]
type=standalone
host=localhost
interface=eth0

## Below is an example clustered configuration. If you use this,
## remove the [zeek] node above.

#[logger-1]
#type=logger
#host=localhost
#
#[manager]
#type=manager
#host=localhost
#
#[proxy-1]
#type=proxy
#host=localhost
#
#[worker-1]
#type=worker
#host=localhost
#interface=eth0
#
#[worker-2]
#type=worker
#host=localhost
#interface=eth0
```

**Understanding the configuration:**

```
┌──────────────────────────────────────────────────────────────┐
│                  NODE.CFG PARAMETERS                         │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  [zeek]                                                      │
│  └─ Name of this node (arbitrary, but descriptive)           │
│                                                              │
│  type=standalone                                             │
│  └─ Single-instance deployment (not a cluster)               │
│     Other options: worker, manager, proxy, logger            │
│                                                              │
│  host=localhost                                              │
│  └─ Hostname or IP of the machine running this node          │
│     For standalone on same machine, localhost is fine        │
│                                                              │
│  interface=eth0                                              │
│  └─ Network interface to capture packets from                │
│     This should match your actual interface name             │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

**Verify your interface name:**

Make sure you specify the correct interface:

```bash
ip link show | grep -E "^[0-9]+:" | awk '{print $2}' | tr -d ':'
```

This shows all your interfaces. Use the one connected to the network you want to monitor (probably `eth0`).

**For a cluster configuration,** node.cfg would be more complex:

```ini
# Example cluster configuration (don't use this unless you have multiple machines)

[manager]
type=manager
host=192.168.1.10

[logger]
type=logger
host=192.168.1.11

[proxy-1]
type=proxy
host=192.168.1.12

[worker-1]
type=worker
host=192.168.1.20
interface=eth1

[worker-2]
type=worker
host=192.168.1.21
interface=eth1
```

Since we're running standalone, the simple configuration is what you need.



#### **2. networks.cfg - Defining Your Networks**

This file tells Zeek which IP ranges are "local" to our monitoring efforts. This is important for determining the direction of connections (inbound vs outbound) and for various detection logic.

First, let's run `ip a` to confirm our current public IP (look for it under the `eth0` header):

```
136.177.84.33/20
```

However, pay attention that in this case it's showing us the `/20` subnet, which includes 4,096 IPs belonging to OTHER DigitalOcean customers. We don't want to classify traffic to those other customer servers as "local".

Rather we'll be using `/32` to ensure Zeek knows only our specific IP is considered "local".

```
136.177.84.33/32
```

Now let's open our config to add this IP:

```bash
sudo nano /opt/zeek/etc/networks.cfg
```


As of Zeek 8.0.1 you'll see the following in the file:
```ini
# List of local networks in CIDR notation, optionally followed by a descriptive
# tag. Private address space defined by Zeek's Site::private_address_space set
# (see scripts/base/utils/site.zeek) is automatically considered local. You can
# disable this auto-inclusion by setting zeekctl's PrivateAddressSpaceIsLocal
# option to 0.
#
# Examples of valid prefixes:
#
# 1.2.3.0/24        Admin network
# 2607:f140::/32    Student network
```

As the comments at the top indicate, one no longer needs to add private network subnets (older versions of Zeek did require this). The comment states they're already automatically included via `Site::private_address_space`. Further, since our VM has a public IP and isn't monitoring an internal RFC1918 network, these aren't relevant to our current setup in this specific case.

Simply add the following line at the bottom, then save the file when done.

```ini
136.177.84.33/32    My Droplet
```

**Why this matters:** Without defining your droplet's IP as local, Zeek can't properly determine connection direction:

- Inbound: External IP → `136.177.84.33` (someone connecting TO our server)
- Outbound: `136.177.84.33` → External IP (our server connecting OUT)

This directional context is critical for many Zeek detection scripts.






---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./docker.md" >}})
[|NEXT|]({{< ref "./start.md" >}})

