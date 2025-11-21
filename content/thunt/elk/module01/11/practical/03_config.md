---
showTableOfContents: true
title: "Configuration Deep Dive - elasticsearch.yml"
type: "page"
---

The heart of Elasticsearch configuration is `/etc/elasticsearch/elasticsearch.yml`. Let's explore this file systematically.

## Step 2.1: View the Default Configuration

```bash
# View the configuration file
sudo cat /etc/elasticsearch/elasticsearch.yml
```

```yaml
# ======================== Elasticsearch Configuration =========================
#
# NOTE: Elasticsearch comes with reasonable defaults for most settings.
#       Before you set out to tweak and tune the configuration, make sure you
#       understand what are you trying to accomplish and the consequences.
#
# The primary way of configuring a node is via this file. This template lists
# the most important settings you may want to configure for a production cluster.
#
# Please consult the documentation for further information on configuration options:
# https://www.elastic.co/guide/en/elasticsearch/reference/index.html
#
# ---------------------------------- Cluster -----------------------------------
#
# Use a descriptive name for your cluster:
#
#cluster.name: my-application
#
# ------------------------------------ Node ------------------------------------
#
# Use a descriptive name for the node:
#
#node.name: node-1
#
# Add custom attributes to the node:
#
#node.attr.rack: r1
#
# ----------------------------------- Paths ------------------------------------
#
# Path to directory where to store the data (separate multiple locations by comma):
#
path.data: /var/lib/elasticsearch
#
# Path to log files:
#
path.logs: /var/log/elasticsearch
#
# ----------------------------------- Memory -----------------------------------
#
# Lock the memory on startup:
#
#bootstrap.memory_lock: true
#
# Make sure that the heap size is set to about half the memory available
# on the system and that the owner of the process is allowed to use this
# limit.
#
# Elasticsearch performs poorly when the system is swapping the memory.
#
# ---------------------------------- Network -----------------------------------
#
# By default Elasticsearch is only accessible on localhost. Set a different
# address here to expose this node on the network:
#
#network.host: 192.168.0.1
#
# By default Elasticsearch listens for HTTP traffic on the first free port it
# finds starting at 9200. Set a specific HTTP port here:
#
#http.port: 9200
#
# For more information, consult the network module documentation.
#
# --------------------------------- Discovery ----------------------------------
#
# Pass an initial list of hosts to perform discovery when this node is started:
# The default list of hosts is ["127.0.0.1", "[::1]"]
#
#discovery.seed_hosts: ["host1", "host2"]
#
# Bootstrap the cluster using an initial set of master-eligible nodes:
#
#cluster.initial_master_nodes: ["node-1", "node-2"]
#
# For more information, consult the discovery and cluster formation module documentation.
#
# ---------------------------------- Various -----------------------------------
#
# Allow wildcard deletion of indices:
#
#action.destructive_requires_name: false

#----------------------- BEGIN SECURITY AUTO CONFIGURATION -----------------------
#
# The following settings, TLS certificates, and keys have been automatically
# generated to configure Elasticsearch security features on 21-11-2025 00:06:13
#
# --------------------------------------------------------------------------------

# Enable security features
xpack.security.enabled: true

xpack.security.enrollment.enabled: true

# Enable encryption for HTTP API client connections, such as Kibana, Logstash, and Agents
xpack.security.http.ssl:
  enabled: true
  keystore.path: certs/http.p12

# Enable encryption and mutual authentication between cluster nodes
xpack.security.transport.ssl:
  enabled: true
  verification_mode: certificate
  keystore.path: certs/transport.p12
  truststore.path: certs/transport.p12
# Create a new cluster with the current node only
# Additional nodes can still join the cluster later
cluster.initial_master_nodes: ["elastic"]

# Allow HTTP API connections from anywhere
# Connections are encrypted and require user authentication
http.host: 0.0.0.0

# Allow other nodes to join the cluster from anywhere
# Connections are encrypted and mutually authenticated
#transport.host: 0.0.0.0

#----------------------- END SECURITY AUTO CONFIGURATION -------------------------
```


You'll see many commented-out options. Let's understand the critical ones.

## Step 2.2: Essential Configuration Parameters

Open the file for editing, since we will be removing many `#`:

```bash
sudo nano /etc/elasticsearch/elasticsearch.yml
```

#### Cluster Configuration

```yaml
# ======================== Cluster Settings ========================
# Give your cluster a descriptive name
cluster.name: threat-hunting-elk

# Node name (use hostname or descriptive name)
node.name: elk-node-01

# Node roles - for single node, use all roles
node.roles: [ master, data, ingest, ml ]
```

**Explanation**:

- **cluster.name**: All nodes with the same cluster name will try to join together. Choose something meaningful.
- **node.name**: Identifies this specific node. Useful when you scale to multiple nodes.
- **node.roles**: Defines what this node does:
    - `master`: Can be elected as master, manages cluster state
    - `data`: Stores data and executes queries
    - `ingest`: Processes documents before indexing
    - `ml`: Runs machine learning jobs (if licensed)

#### Network Configuration

```yaml
# ======================== Network Settings ========================
# Bind to specific IP (0.0.0.0 for all interfaces, or specific IP)
network.host: 0.0.0.0

# HTTP port for REST API
http.port: 9200

# Transport port for node-to-node communication
transport.port: 9300
```

**Explanation**:

- **network.host**: `0.0.0.0` allows connections from any IP. For production, use specific IP addresses.
- **http.port**: The REST API port you'll use with curl.
- **transport.port**: Used for cluster communication (not needed for single node, but good to know).

**Security Warning**: Binding to 0.0.0.0 in production requires proper firewall rules. For our lab, it's acceptable.

#### Discovery and Cluster Formation

```yaml
# ======================== Discovery Settings ========================
# For single-node setup
discovery.type: single-node
```

**Explanation**: `single-node` tells Elasticsearch not to look for other nodes. For multi-node clusters, you'd use different settings (covered in Module 1.2).

#### Path Configuration

```yaml
# ======================== Paths ========================
# Path to directory where data is stored
path.data: /var/lib/elasticsearch

# Path to log files
path.logs: /var/log/elasticsearch
```

**Explanation**: These paths were set during installation. Only change if you have specific storage requirements (e.g., separate SSD for data).


#### Security Configuration (Elasticsearch 8.x Default)

```yaml
#----------------------- BEGIN SECURITY AUTO CONFIGURATION -----------------------
#
# The following settings, TLS certificates, and keys have been automatically
# generated to configure Elasticsearch security features on 21-11-2025 00:06:13
#
# --------------------------------------------------------------------------------
# Enable security features
xpack.security.enabled: true

xpack.security.enrollment.enabled: true

# Enable encryption for HTTP API client connections, such as Kibana, Logstash, and Agents
xpack.security.http.ssl:
  enabled: true
  keystore.path: certs/http.p12

# Enable encryption and mutual authentication between cluster nodes
xpack.security.transport.ssl:
  enabled: true
  verification_mode: certificate
  keystore.path: certs/transport.p12
  truststore.path: certs/transport.p12
```

**Explanation**: Elasticsearch 8.x enables security by default, including TLS encryption. The certificates are auto-generated during installation.

**For Lab Purposes**: If you want to disable security temporarily to simplify initial learning (NOT RECOMMENDED for production):

```yaml
# ONLY FOR LAB - Disable security
xpack.security.enabled: false
xpack.security.http.ssl.enabled: false
xpack.security.transport.ssl.enabled: false
```

**Important**: We'll proceed with security enabled for this guide. Disabling is your choice for simplified learning, but understand the security implications.

### Step 2.3: Save and Exit

After making your changes:

- Press `Ctrl+X` to exit
- Press `Y` to confirm save
- Press `Enter` to confirm filename



---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./02_install.md" >}})
[|NEXT|]({{< ref "./04_jvm.md" >}})

