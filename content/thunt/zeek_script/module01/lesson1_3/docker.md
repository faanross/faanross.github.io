---
showTableOfContents: true
title: "Part 4 - Installation Method 3: Docker Container"
type: "page"
---

## **PART 4: INSTALLATION METHOD 3 - DOCKER CONTAINER (MODERN ALTERNATIVE)**

### **Understanding Containerized Deployment**

Containers offer a modern approach to deploying applications. A Docker container packages Zeek with all its dependencies in an isolated environment. This has interesting advantages and some trade-offs.

**Container benefits and limitations:**

```
┌──────────────────────────────────────────────────────────────┐
│            CONTAINERIZED ZEEK DEPLOYMENT                     │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ✓ ISOLATION                                                 │
│    Zeek runs in its own environment, won't conflict with     │
│    other software on your system                             │
│                                                              │
│  ✓ PORTABILITY                                               │
│    Same container image runs identically on any Docker host  │
│                                                              │
│  ✓ VERSION CONTROL                                           │
│    Pin specific Zeek versions, easy to test upgrades         │
│                                                              │
│  ✓ RAPID DEPLOYMENT                                          │
│    Pull image and run - no compilation or dependency mgmt    │
│                                                              │
│  ✗ PACKET CAPTURE COMPLEXITY                                 │
│    Containers add network abstraction layers that can        │
│    complicate packet capture and monitoring                  │
│                                                              │
│  ✗ PERFORMANCE OVERHEAD                                      │
│    Small performance penalty from containerization           │
│    (usually <5%, but matters at high packet rates)           │
│                                                              │
│  ✗ LEARNING CURVE                                            │
│    Requires Docker knowledge in addition to Zeek knowledge   │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

Containers are excellent for development, testing, and certain production scenarios. However, for high-performance packet capture, native installation often performs better. Let's explore containerized deployment so you understand the option.




### **Installing Docker**

First, install Docker on your Ubuntu droplet:

```bash
# Update package index
sudo apt update

# Install prerequisites
sudo apt install -y apt-transport-https ca-certificates curl software-properties-common

# Add Docker's GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | \
    sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Add Docker repository
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] \
    https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | \
    sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io

# Add your user to docker group (avoid needing sudo for docker commands)
sudo usermod -aG docker $USER

# Log out and back in for group changes to take effect
exit
```

After logging back in:

```bash
# Verify Docker installation
docker --version
docker run hello-world
```

You should see Docker version information and a "Hello from Docker!" message.

### **Running Zeek in a Container**

The simplest way to run Zeek in Docker is using an official or community image:

```bash
# Pull a Zeek Docker image
docker pull zeek/zeek:latest

# Run Zeek in a container
docker run -it --rm \
    --net=host \
    --cap-add=NET_RAW \
    --cap-add=NET_ADMIN \
    -v /opt/zeek/logs:/usr/local/zeek/logs \
    zeek/zeek:latest
```

**Understanding the docker run options:**

```
┌──────────────────────────────────────────────────────────────┐
│               DOCKER RUN OPTION BREAKDOWN                    │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  -it                                                         │
│  └─ Interactive terminal (you can interact with container)   │
│                                                              │
│  --rm                                                        │
│  └─ Remove container when it exits (don't leave debris)      │
│                                                              │
│  --net=host                                                  │
│  └─ Use host's network stack directly (needed for packet     │
│     capture)                                                 │
│                                                              │
│  --cap-add=NET_RAW                                           │
│  └─ Give container ability to capture raw packets            │
│                                                              │
│  --cap-add=NET_ADMIN                                         │
│  └─ Give container network administration capabilities       │
│                                                              │
│  -v /opt/zeek/logs:/usr/local/zeek/logs                      │
│  └─ Mount host directory into container so logs persist      │
│     after container exits                                    │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### **Creating a Custom Zeek Container**

For real deployments, you'll want to create a custom container with your specific configuration and scripts. Let's create a Dockerfile:

```bash
# Create a directory for your Zeek container
mkdir ~/zeek-container
cd ~/zeek-container

# Create a Dockerfile
cat > Dockerfile << 'EOF'
FROM ubuntu:24.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    gnupg2 \
    && rm -rf /var/lib/apt/lists/*

# Add Zeek repository and install
RUN curl -fsSL https://download.opensuse.org/repositories/security:zeek/xUbuntu_24.04/Release.key | \
    gpg --dearmor -o /usr/share/keyrings/zeek-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/zeek-archive-keyring.gpg] \
    http://download.opensuse.org/repositories/security:/zeek/xUbuntu_22.04/ /" \
    > /etc/apt/sources.list.d/zeek.list && \
    apt-get update && \
    apt-get install -y zeek && \
    rm -rf /var/lib/apt/lists/*

# Add Zeek to PATH
ENV PATH="/opt/zeek/bin:${PATH}"

# Expose common Zeek ports (if running in daemon mode)
EXPOSE 47760-47770

# Set working directory
WORKDIR /opt/zeek

# Default command
CMD ["/opt/zeek/bin/zeekctl", "deploy"]
EOF

# Build the container
docker build -t my-zeek:latest .
```

This creates a custom Zeek container image that you can deploy with your specific configuration.

### **When to Use Containers vs Native Installation**

**Use native installation when:**

- Maximum performance is critical (high-bandwidth monitoring)
- You're deploying on bare metal or dedicated VMs
- You want the simplest possible setup
- You're managing infrastructure with traditional tools

**Use containers when:**

- You need to deploy multiple Zeek versions simultaneously
- You're in a Kubernetes or container-orchestrated environment
- You want to isolate Zeek from other system components
- You're doing development and testing (easy to spin up/tear down)
- You're deploying in cloud-native architectures

For this course, we'll primarily use native installation since it's simpler and provides better performance for learning. But knowing how to containerize Zeek gives you flexibility for future deployments.




---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./compile.md" >}})
[|NEXT|]({{< ref "./configure.md" >}})

