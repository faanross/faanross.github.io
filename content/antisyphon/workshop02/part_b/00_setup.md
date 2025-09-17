---
showTableOfContents: true
title: "Setup"
type: "page"
---

## Preface
Please note you don't need a special VM for this course since we'll be developing. I recommend using your base OS ("daily driver")
since that way you can just use this setup to continue your coding efforts following this course. If however you wanted to keep
things "seperated", you could def do it on a VM, it won't hurt, but since that's not required I'm not going to provide instructions.

There are essentially 4 things you need: go, an IDE, the course repo, and "extra apps" (optional). I will provide instructions on setting
these up for all three major OS - Windows, Darwin, and Linux. Note I am using Darwin, just my preference, but there should be no reason
you could not follow along using any OS.

## 1. Install Go Programming Language

### Windows
1. Visit https://go.dev/dl/
2. Download the Windows installer (`.msi` file)
3. Run the installer - just click "Next" through all the steps
4. **Verify:** Open Command Prompt and type: `go version`
    - You should see something like: `go version go1.23.x windows/amd64`

### macOS
1. Visit https://go.dev/dl/
2. Download the macOS installer (`.pkg` file)
3. Double-click to install - follow the prompts
4. **Verify:** Open Terminal and type: `go version`
    - You should see something like: `go version go1.23.x darwin/amd64`

### Linux
1. Visit https://go.dev/dl/
2. Download the Linux archive (`.tar.gz` file)
3. Open Terminal and run:
   ```bash
   sudo rm -rf /usr/local/go
   sudo tar -C /usr/local -xzf go1.23.x.linux-amd64.tar.gz
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
   source ~/.bashrc
   ```
4. **Verify:** Type `go version`
    - You should see something like: `go version go1.23.x linux/amd64`

## 2. Set Up Your Code Editor

### Option A: GoLand (Recommended - What I'll Use)
- Download from: https://www.jetbrains.com/goland/
- It's a paid tool, but you can get a **30-day free trial** (no credit card needed)
- You can also use the code `MAX_JETBRAINS` (please keep this on the d-low) to get a 6-month license
- Just install and open - it configures Go automatically

### Option B: VS Code (Free Alternative)
1. Download VS Code from: https://code.visualstudio.com/
2. Install it on your system
3. Add Go support:
    - Open VS Code
    - Click the Extensions icon (square icon on left sidebar)
    - Search for "Go" (by Google)
    - Click Install
    - Also search for "YAML" and install it (for configuration files)

## 3. Download the Workshop Repository

### Easy Method (No Git Required):
1. Visit: https://github.com/faanross/workshop_antisyphon_18092025
2. Click the green "Code" button
3. Click "Download ZIP"
4. Extract the ZIP file to a folder you can remember (like your Desktop)
5. Open this folder in your IDE

### Alternative Method (If You Have Git):
```bash
git clone https://github.com/faanross/workshop_antisyphon_18092025
cd workshop_antisyphon_18092025
```

## 4. Install Network Tools

We'll need tools to test HTTP and DNS endpoints during the workshop.

### For HTTP Testing (curl or browser)

#### Windows
- **Built-in:** You can use PowerShell's `Invoke-WebRequest` or just your web browser
- **Or install curl:** Open PowerShell as Administrator and run:
  ```powershell
  winget install curl.curl
  ```

#### macOS
- **Already installed!** Just open Terminal and type `curl --version` to verify

#### Linux
- **Usually installed!** If not: `sudo apt install curl` (Ubuntu/Debian) or `sudo yum install curl` (RedHat/Fedora)

### For DNS Testing (dig or nslookup)

#### Windows
- **Built-in:** Use `nslookup` in Command Prompt (already installed!)
- Example: `nslookup google.com`

#### macOS
- **Built-in:** Both `dig` and `nslookup` are already installed
- Example: `dig google.com` or `nslookup google.com`

#### Linux
- **Install dig:**
    - Ubuntu/Debian: `sudo apt install dnsutils`
    - RedHat/Fedora: `sudo yum install bind-utils`

## Quick Test Checklist ✓

Run these commands to make sure everything works:

1. [ ] Go is installed: `go version`
2. [ ] Your IDE opens and recognizes Go files
3. [ ] You can access the workshop repository files
4. [ ] HTTP tool works: `curl https://google.com` (or open https://google.com in browser)
5. [ ] DNS tool works: `nslookup google.com` (Windows) or `dig google.com` (Mac/Linux)

## Troubleshooting

- **"Command not found"** - Try closing and reopening your terminal/command prompt
- **Permission errors** - On Mac/Linux, add `sudo` before install commands
- **Windows security warnings** - It's okay to allow Go and the tools through Windows Defender

## You're All Set! 

If you completed the checklist above, you're ready for tomorrow's workshop. Don't worry if something doesn't work perfectly - we'll have time at the beginning to help with any setup issues.
Also as mentioned we'll have two awesome people - Dezzy and Hermon (h,k) - that have offered to help people that are having issues.

See you tomorrow!


___
[|TOC|]({{< ref "../moc.md" >}})
[|NEXT|]({{< ref "01_interfaces.md" >}})