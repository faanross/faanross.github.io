---
showTableOfContents: true
title: "Part 3 - Installation Method 2: Compiling From Source"
type: "page"
---


## **PART 3: INSTALLATION METHOD 2 - COMPILING FROM SOURCE (ADVANCED)**

### **Why Compile from Source?**

Before we dive into compilation, let's understand when and why you'd want to compile Zeek from source instead of using packages.

**When to compile from source:**

```
┌──────────────────────────────────────────────────────────────┐
│           REASONS TO COMPILE FROM SOURCE                     │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ✓ LATEST FEATURES                                           │
│    Get the absolute latest Zeek version, including newest    │
│    features and bug fixes not yet in package releases        │
│                                                              │
│  ✓ OPTIMIZATION                                              │
│    Compile with CPU-specific optimizations for better        │
│    performance on your exact hardware                        │
│                                                              │
│  ✓ CUSTOMIZATION                                             │
│    Enable/disable specific features, add custom patches,     │
│    modify source code for research purposes                  │
│                                                              │
│  ✓ LEARNING                                                  │
│    Understand how Zeek is built, what components it uses,    │
│    how different parts fit together                          │
│                                                              │
│  ✓ UNSUPPORTED PLATFORMS                                     │
│    When packages aren't available for your distribution      │
│    or you need to run on unusual platforms                   │
│                                                              │
│  ✗ COMPLEXITY                                                │
│    More complex, more opportunities for problems             │
│    Requires managing dependencies manually                   │
│    Updates require recompiling                               │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

For production environments, package installation is usually preferred due to its simplicity and reliability. But for learning, performance-critical deployments, or when you need cutting-edge features, source compilation is valuable.

**Note:** If you already installed Zeek via package manager above, you don't need to compile from source, this section is included so you understand the process and can use it when needed.

### **Installing Build Dependencies**

Compiling Zeek requires various development tools and libraries. Let's install everything we need:

```bash
# Development tools
sudo apt install -y build-essential cmake git

# Required libraries
sudo apt install -y libpcap-dev libssl-dev python3-dev \
    swig zlib1g-dev libmaxminddb-dev libgoogle-perftools-dev \
    libzmq3-dev

# Optional but recommended libraries
sudo apt install -y flex bison libkrb5-dev
```

**Understanding the dependencies:**

```
┌──────────────────────────────────────────────────────────────┐
│                  BUILD DEPENDENCIES                          │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  build-essential                                             │
│  • GCC compiler, make, and other essential build tools       │
│                                                              │
│  cmake                                                       │
│  • Build system generator used by Zeek                       │
│                                                              │
│  libpcap-dev                                                 │
│  • Packet capture library (the heart of traffic capture)     │
│                                                              │
│  libssl-dev                                                  │
│  • SSL/TLS library for parsing encrypted traffic metadata    │
│                                                              │
│  python3-dev                                                 │
│  • Python headers for building Python bindings               │
│                                                              │
│  swig                                                        │
│  • Tool for generating Python bindings to C++ code           │
│                                                              │
│  zlib1g-dev                                                  │
│  • Compression library for log file compression              │
│                                                              │
│  libmaxminddb-dev                                            │
│  • GeoIP database library for IP geolocation                 │
│                                                              │
│  libgoogle-perftools-dev                                     │
│  • Allows us to build Zeek with performance tools            │                                        
│                                                              │
│  libzmq3-dev                                                 │
│  • Required for ZeroMQ cluster backend support               │                                        
│                                                              │
│  flex, bison                                                 │
│  • Parser generators used in building analyzers              │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### **Downloading Zeek Source Code**

Zeek's source code is hosted on GitHub. Let's clone the repository:

```bash
# Create a directory for source code
mkdir -p ~/src
cd ~/src

# Clone the Zeek repository
git clone --recursive https://github.com/zeek/zeek
cd zeek

# Check out the latest stable release
git checkout v8.0.1  # Or whatever the latest stable version is
```

The `--recursive` flag is important - it ensures that Zeek's submodules (auxiliary repositories it depends on) are also cloned.

**Understanding the source tree:**

```bash
ls -la
```

You'll see:

```
zeek/
├── src/                    # C++ source code
│   ├── analyzer/          # Protocol analyzers
│   ├── packet_analysis/   # Packet processing
│   ├── script_opt/        # Script optimizer
│   └── ...
├── scripts/                # Zeek script language files
│   ├── base/              # Core scripts
│   ├── policy/            # Optional policies
│   └── site/              # Site customization templates
├── testing/                # Test suites
├── aux/                    # Auxiliary tools and libraries
├── cmake/                  # CMake build configuration
├── CMakeLists.txt          # Main build configuration
└── README                  # Documentation
```

### **Configuring the Build**

Now we configure how Zeek will be built using CMake:

```bash
# Create a build directory (keeps source tree clean)
cd ~/src/zeek
mkdir build
cd build

# Configure the build
cmake .. \
    -DCMAKE_INSTALL_PREFIX=/opt/zeek \
    -DENABLE_PERFTOOLS=ON \
    -DZEEK_ETC_INSTALL_DIR=/opt/zeek/etc \
    -DCMAKE_BUILD_TYPE=Release
```

**Understanding the configuration options:**

```
┌──────────────────────────────────────────────────────────────┐
│                CMAKE CONFIGURATION OPTIONS                   │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  -DCMAKE_INSTALL_PREFIX=/opt/zeek                            │
│  └─ Where to install Zeek (same location as package install) │
│                                                              │
│  -DENABLE_PERFTOOLS=ON                                       │
│  └─ Enable performance profiling tools (useful for tuning)   │
│                                                              │
│  -DZEEK_ETC_INSTALL_DIR=/opt/zeek/etc                        │
│  └─ Where to place configuration files                       │
│                                                              │
│  -DCMAKE_BUILD_TYPE=Release                                  │
│  └─ Build optimized binaries (vs Debug with symbols)         │
│                                                              │
│  OPTIONAL FLAGS YOU MIGHT ADD:                               │
│                                                              │
│  -DCMAKE_CXX_FLAGS="-march=native"                           │
│  └─ Optimize for your specific CPU (best performance)        │
│                                                              │
│  -DENABLE_JEMALLOC=ON                                        │
│  └─ Use jemalloc for better memory performance               │
│                                                              │
│  -DINSTALL_AUX_TOOLS=ON                                      │
│  └─ Install additional utilities                             │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

CMake will check that all dependencies are present and generate build files. You'll see output about what features are enabled and disabled. If any required dependencies are missing, CMake will error out with a message telling you what's missing.

### **Compiling Zeek**

Now comes the actual compilation. This is CPU-intensive and will take a while:

```bash
# Compile using all available CPU cores
make -j$(nproc)
```

The `-j$(nproc)` flag tells make to use all your CPU cores in parallel. On a 4-core droplet, this compiles about 4x faster than single-threaded compilation.

**What to expect:**

Compilation will take 15-45 minutes depending on your hardware. You'll see thousands of lines of output as each source file is compiled:

```
[ 12%] Building CXX object src/CMakeFiles/zeek.dir/Analyzer.cc.o
[ 12%] Building CXX object src/CMakeFiles/zeek.dir/Conn.cc.o
[ 13%] Building CXX object src/CMakeFiles/zeek.dir/DbgHelp.cc.o
...
```

The percentage indicates progress. Go get coffee - this will take a while.

**If compilation fails:**

Don't panic. Compilation failures are usually due to missing dependencies. Read the error message carefully - it will tell you what's missing. Install the missing package and run `make` again. CMake remembers what it already compiled, so it won't start over from scratch.

### **Installing the Compiled Binaries**

Once compilation finishes successfully (you'll see `[100%] Built target zeek`), install Zeek:

```bash
sudo make install
```

This copies all the compiled binaries, scripts, and configuration files to `/opt/zeek/`. The installation takes just a minute or two.

**Verify the installation:**

```bash
/opt/zeek/bin/zeek --version
```

You should see version information, confirming Zeek is installed.

### **Post-Installation Setup**

Just like with package installation, add Zeek to your PATH and configure ZeekControl. The steps are identical to what we did in the package installation section, so refer back to those instructions.

### **Source vs. Package: Performance Comparison**

You might wonder: "Does compiling from source actually make Zeek faster?" The answer is: sometimes, but not always dramatically. Here's what to expect:

```
┌──────────────────────────────────────────────────────────────┐
│          PERFORMANCE COMPARISON: SOURCE VS PACKAGE           │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Scenario 1: Default compilation flags                       │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│  Package:  1000 Mbps max throughput                          │
│  Source:   1050 Mbps max throughput                          │
│  Gain:     ~5% (minimal difference)                          │
│                                                              │
│  Scenario 2: CPU-specific optimization (-march=native)       │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│  Package:  1000 Mbps max throughput                          │
│  Source:   1150 Mbps max throughput                          │
│  Gain:     ~15% (noticeable improvement)                     │
│                                                              │
│  Scenario 3: PGO (Profile-Guided Optimization)               │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│  Package:  1000 Mbps max throughput                          │
│  Source:   1250 Mbps max throughput                          │
│  Gain:     ~25% (significant, but complex to implement)      │
│                                                              │
│  VERDICT: For most deployments, packages are fine. Compile   │
│  from source when you need that last 10-25% performance or   │
│  need features not in package releases.                      │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```



---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./package.md" >}})
[|NEXT|]({{< ref "./docker.md" >}})

