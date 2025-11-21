---
showTableOfContents: true
title: "Installing Elasticsearch 8.x"
type: "page"
---

## Understanding the Installation Process

Elasticsearch is distributed as a package that can be installed via package managers (APT for Debian/Ubuntu, YUM for RHEL/CentOS) or as a tarball. We'll use the APT repository method as it provides easy updates and proper service management.

## Step 1.1: Prepare Your System

First, update your system packages and install required dependencies:

```bash
# Update package lists
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install apt-transport-https wget gnupg2 -y
```

**Explanation**:

- `apt-transport-https` enables APT to retrieve packages over HTTPS
- `wget` is used to download files
- `gnupg2` handles GPG keys for package verification

## Step 1.2: Add Elasticsearch Repository

Import the Elasticsearch GPG key and add the repository:

```bash
# Download and add Elasticsearch GPG key
wget -qO - https://artifacts.elastic.co/GPG-KEY-elasticsearch | sudo gpg --dearmor -o /usr/share/keyrings/elasticsearch-keyring.gpg

# Add Elasticsearch repository
echo "deb [signed-by=/usr/share/keyrings/elasticsearch-keyring.gpg] https://artifacts.elastic.co/packages/8.x/apt stable main" | sudo tee /etc/apt/sources.list.d/elastic-8.x.list
```

**Explanation**:

- The GPG key verifies package authenticity
- We're specifying the 8.x branch to ensure we get Elasticsearch 8
- The `signed-by` parameter tells APT which key to use for verification

## Step 1.3: Install Elasticsearch

```bash
# Update package lists with new repository
sudo apt update

# Install Elasticsearch
sudo apt install elasticsearch -y
```

**Critical Note**: During installation, Elasticsearch 8.x will generate a superuser password and enrollment tokens. **Save this output immediately!** You'll see something like:

```
--------------------------- Security autoconfiguration information ------------------------------

Authentication and authorization are enabled.
TLS for the transport and HTTP layers is enabled and configured.

The generated password for the elastic built-in superuser is : YOUR_PASSWORD_HERE

If this node should join an existing cluster, you can reconfigure this with
'/usr/share/elasticsearch/bin/elasticsearch-reconfigure-node --enrollment-token <token-here>'
```

**Action Required**: Copy this password to a secure location. If you lose it, you'll need to reset it later.


Now execute the following statements to configure elasticsearch service to start automatically using systemd
```bash
 sudo systemctl daemon-reload
 sudo systemctl enable elasticsearch.service
```


Also note that you can start elasticsearch service by executing
```bash
 sudo systemctl start elasticsearch.service
```

## Step 1.4: Understanding Installation Locations

After installation, Elasticsearch files are distributed across several directories:

| Directory                   | Purpose       | Contents                                          |
| --------------------------- | ------------- | ------------------------------------------------- |
| `/etc/elasticsearch/`       | Configuration | elasticsearch.yml, jvm.options, log4j2.properties |
| `/var/lib/elasticsearch/`   | Data          | Indices, shards, cluster state                    |
| `/var/log/elasticsearch/`   | Logs          | Application and slow logs                         |
| `/usr/share/elasticsearch/` | Application   | Binaries, plugins, modules                        |

## Step 1.5: Configure System Settings

Elasticsearch requires specific system settings to function optimally:

```bash
# Increase max file descriptors
echo "elasticsearch - nofile 65535" | sudo tee -a /etc/security/limits.conf

# Increase max locked memory
echo "elasticsearch - memlock unlimited" | sudo tee -a /etc/security/limits.conf

# Disable swap (critical for performance)
sudo swapoff -a

# Make swap disable persistent (edit /etc/fstab and comment out swap lines)
sudo sed -i.bak '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab

# Increase virtual memory
sudo sysctl -w vm.max_map_count=262144

# Make virtual memory setting persistent
echo "vm.max_map_count=262144" | sudo tee -a /etc/sysctl.conf
```

**Explanation of Each Setting**:

1. **File Descriptors (nofile)**: Elasticsearch opens many files simultaneously (indices, shards, network connections). 65535 ensures we don't hit limits.

2. **Locked Memory (memlock)**: Prevents the OS from swapping Elasticsearch memory to disk, which would cause severe performance degradation.

3. **Swap Disabled**: Swapping kills Elasticsearch performance. It's better to crash than swap in production.

4. **Virtual Memory (max_map_count)**: Elasticsearch uses memory-mapped files. The default Linux limit (65530) is too low; 262144 is the recommended minimum.





---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./01_intro.md" >}})
[|NEXT|]({{< ref "./03_config.md" >}})

