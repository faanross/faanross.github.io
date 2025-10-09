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



#### **3. zeekctl.cfg - ZeekControl Settings**

Review and adjust these key settings:

```bash
sudo nano /opt/zeek/etc/zeekctl.cfg
```

**Settings to configure:**

```ini
# Email for notices (optional - change if you want email alerts)
MailTo = root@localhost

# Log rotation: how often to rotate logs (in seconds)
# 3600 = 1 hour (default)
LogRotationInterval = 3600

# Log expiration: how long to keep old logs
# 0 = never expire (default)
# Set to number of days if you want automatic cleanup
LogExpireInterval = 28

# Stats logging (keep enabled to track Zeek performance)
StatsLogEnable = 1
StatsLogExpireInterval = 0
```

**Recommended changes for learning/testing:**

```ini
LogRotationInterval = 3600    # Keep at 1 hour (good default)
LogExpireInterval = 28          # Delete logs after 28 days (saves disk space)
```


### **Configuring AF_PACKET for Better Performance**

Remember in Lesson 1.2 when we discussed packet acquisition methods? Let's configure Zeek to use `AF_PACKET` for better performance:

```bash
sudo nano /opt/zeek/etc/node.cfg
```

Modify your standalone configuration:

```ini
[zeek]
type=standalone
host=localhost
interface=eth0

# AF_PACKET configuration
lb_method=custom
lb_procs=2
pin_cpus=0,1
af_packet_fanout_id=23
af_packet_fanout_mode=AF_PACKET_FANOUT_HASH
af_packet_buffer_size=128*1024*1024
```

**Understanding AF_PACKET settings:**

```
┌──────────────────────────────────────────────────────────────┐
│              AF_PACKET CONFIGURATION                         │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  lb_method=custom                                            │
│  └─ Use custom load balancing (AF_PACKET fanout)             │
│                                                              │
│  lb_procs=2                                                  │
│  └─ Number of processes to split traffic across              │
│     Should be ≤ number of CPU cores available                │
│                                                              │
│  pin_cpus=0,1                                                │
│  └─ Pin processes to specific CPU cores for better           │
│     cache performance                                        │
│                                                              │
│  af_packet_fanout_id=23                                      │
│  └─ Fanout group ID (arbitrary number, must be unique        │
│     per interface)                                           │
│                                                              │
│  af_packet_fanout_mode=AF_PACKET_FANOUT_HASH                 │
│  └─ How to distribute packets (HASH maintains                │
│     connection affinity)                                     │
│                                                              │
│  af_packet_buffer_size=128*1024*1024                         │
│  └─ Size of packet buffer (128 MB here)                      │
│     Larger buffers reduce dropped packets under burst loads  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

**For a VM with 4 vCPUs, optimal settings might be:**

```ini
lb_procs=3              # Use 3 processes (leave 1 CPU for OS)
pin_cpus=0,1,2          # Pin to first 3 cores
```


### **Site-Specific Configuration: local.zeek**

The most important configuration file for customization is `/opt/zeek/share/zeek/site/local.zeek`. This file is where you enable/disable policies, load additional scripts, and customize Zeek's behavior. The default file includes useful comments. Let's review the following example configuration:

```zeek
##! Local site policy. Customize as appropriate.
##!
##! This file will not be overwritten when upgrading or reinstalling!

# Load base scripts
@load base/frameworks/software/vulnerable.zeek
@load base/frameworks/software/version-changes.zeek
@load base/frameworks/software/windows-version-detection.zeek

# Load protocols
@load protocols/conn/known-hosts
@load protocols/conn/known-services
@load protocols/dhcp/software
@load protocols/dns/detect-external-names
@load protocols/ftp/detect
@load protocols/ftp/software
@load protocols/http/detect-sqli
@load protocols/http/detect-webapps
@load protocols/http/software
@load protocols/ssh/detect-bruteforcing
@load protocols/ssh/geo-data
@load protocols/ssh/interesting-hostnames
@load protocols/ssh/software
@load protocols/ssl/known-certs
@load protocols/ssl/validate-certs

# Load file analysis
@load frameworks/files/hash-all-files
@load frameworks/files/detect-MHR

# Enable notice/alert logging to notice.log
redef Notice::policy += {
    [$action = Notice::ACTION_LOG]
};

# Set site-specific variables
redef Site::local_nets += {
    # Add your networks here (should match networks.cfg)
    10.0.0.0/8,
    172.16.0.0/12,
    192.168.0.0/16,
};

# Increase some resource limits for learning environment
redef table_expire_interval = 30min;
redef default_table_expire_func = function() {};
```

**Understanding the configuration:**

```
┌──────────────────────────────────────────────────────────────┐
│              LOCAL.ZEEK STRUCTURE                            │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  @load directives                                            │
│  └─ Include additional script packages                       │
│     Think of these like "import" statements                  │
│                                                              │
│  redef statements                                            │
│  └─ Override default values of variables                     │
│     Customize Zeek's behavior                                │
│                                                              │
│  SCRIPT CATEGORIES:                                          │
│                                                              │
│  base/frameworks/*                                           │
│  └─ Core functionality frameworks                            │
│                                                              │
│  protocols/*                                                 │
│  └─ Protocol-specific analysis and detection                 │
│                                                              │
│  policy/*                                                    │
│  └─ Optional detection policies                              │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

The scripts you load in `local.zeek` determine what analysis Zeek performs. The base installation is conservative - it provides visibility but doesn't enable every possible detection. As you progress through this course, you'll add more `@load` directives to enable specific detections.

Once you feel confident that you understand the different config settings open your own `local.zeek` script and make any edits you feel inspired to:

```bash
sudo nano /opt/zeek/share/zeek/site/local.zeek
```
Alternatively, you could leave things as they are for now - we will constantly turn back to this file and make edits
as the course progresses since it is the foundational script used for all scripting that follows.

**NOTE:** If you do decide to make some edits, it would be a good idea to make a backup copy of your default script using `cp`, just in case anything goes awry and you make breaking changes.




---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./docker.md" >}})
[|NEXT|]({{< ref "./start.md" >}})

