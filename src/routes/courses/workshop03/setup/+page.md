---
layout: workshop03
title: "Setup Guide"
---

## Preface

Please note you don't need a special VM for development since we'll be writing code on your "daily driver" machine. Using your base OS means you can continue using this setup for future Go development after the workshop.

There are essentially 4 things you need: **Go**, an **IDE**, the **workshop repo**, and (ideally) a **test machine** for running the final payload. I will provide instructions for all three major operating systems - Windows, macOS, and Linux. Note I am using macOS, but there should be no reason you couldn't follow along using any OS.


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

You can use any code editor or IDE you prefer - no specific one is required for this workshop.

### Option A: GoLand (Recommended - What I'll Use)
- Download from: https://www.jetbrains.com/goland/
- It's a paid tool, but you can get a **30-day free trial** (no credit card needed)
- Just install and open - it configures Go automatically

### Option B: VS Code (Free Alternative)
1. Download VS Code from: https://code.visualstudio.com/
2. Install it on your system
3. Add Go support:
    - Open VS Code
    - Click the Extensions icon (square icon on left sidebar)
    - Search for "Go" (by Google)
    - Click Install

### Option C: Other Editors
- **Zed** (https://zed.dev) - Written in Rust, great LLM integration, good Go support
- **NeoVim** or **Helix** - If you're comfortable with terminal editors, both can be configured for Go development


## 3. Download the Workshop Repository

### Easy Method (No Git Required):
1. Visit: https://github.com/faanross/workshop_antisyphon_23012026
2. Click the green "Code" button
3. Click "Download ZIP"
4. Extract the ZIP file to a folder you can remember (like your Desktop)
5. Open this folder in your IDE

### Alternative Method (If You Have Git):
```bash
git clone https://github.com/faanross/workshop_antisyphon_23012026
cd workshop_antisyphon_23012026
```

Note: Git is optional but recommended. If you don't have it installed, the ZIP download method works perfectly fine.


## 4. Test Machine for Running the Final Payload

### Important Context

The final step of this workshop is to actually run our reflective loader and execute `calc.exe` on a Windows machine. This is the payoff for all the work we'll do - seeing your shellcode execute in memory!

**However**, if you're unable to set up a test machine, you can still complete 99% of the workshop. Everything up to the final execution step will work on any development platform. So don't let this be a blocker - set it up if you can, but don't stress if you can't.

### Why a Separate Machine?

The final payload we build is essentially what malware looks like - it's a reflective DLL loader. Windows Defender (and other AV solutions) will flag and block it. Running this on your main development machine would require disabling your security software, which I **strongly recommend against**.

Instead, use one of these options:

### Option A: Dedicated Physical Machine
If you have an extra Windows machine on your LAN (old laptop, spare PC), this is the simplest option:
- Install Windows (any modern version)
- Disable Windows Defender (see instructions below)
- Connect it to the same network as your dev machine

### Option B: Windows Virtual Machine
Run a Windows VM on your development machine:
- **VMware**, **VirtualBox**, or **Parallels** (Mac) all work
- Create a Windows 10/11 VM
- Disable Windows Defender inside the VM (see instructions below)

### ⚠️ Important Exception: Apple Silicon (M-series) Mac Users

If you're developing on an M1, M2, M3, or M4 Mac, **you cannot use a Windows VM** for this workshop. The shellcode we build targets AMD64 (x86-64) architecture, which is incompatible with ARM-based virtualization.

**Your options:**
- Use a physical x86-64 Windows machine (dedicated host on your LAN)
- Borrow or repurpose an old Intel-based computer

This is an architecture limitation, not a software one - ARM cannot execute x86-64 shellcode even through emulation for our purposes.


### Disabling Windows Defender

For the test/victim machine only (never your main machine!), you'll need to disable Windows Defender. This is a two-step process:

**Step 1: Temporarily Disable via Windows Settings**

You need to disable Defender through Windows settings first, otherwise it will block the download and execution of the remover tool.

1. Open **Windows Security** (search for it in the Start menu)
2. Go to **Virus & threat protection**
3. Click **Manage settings** under "Virus & threat protection settings"
4. Turn OFF **Real-time protection**
5. Turn OFF **Tamper Protection** (if present)

This is a temporary disable - Windows will re-enable it eventually, which is why we need Step 2.

**Step 2: Deep Disable with Defender Remover**

Now that Defender is temporarily disabled, we can run the tool that permanently disables it.

**Recommended Tool:** [Windows Defender Remover](https://github.com/ionuttbara/windows-defender-remover/releases)

1. Go to the releases page linked above
2. Download the latest release for your Windows version
3. Run it **as Administrator** on your test machine only
4. Follow the on-screen instructions
5. Reboot when prompted

This tool performs a "deep disable" of Windows Defender, preventing it from re-enabling itself. Again - only do this on a dedicated test machine, never on your main system.


## Quick Test Checklist

Run these commands to make sure your **development** environment is ready:

- [ ] Go is installed: `go version`
- [ ] Your IDE opens and recognizes Go files
- [ ] You can access the workshop repository files
- [ ] (Optional) Test machine is set up with Defender disabled


## Troubleshooting

- **"Command not found"** - Try closing and reopening your terminal/command prompt
- **Permission errors** - On Mac/Linux, add `sudo` before install commands
- **Windows security warnings** - It's okay to allow Go and the tools through Windows Defender (on your dev machine)
- **Can't disable Defender on test VM** - Make sure you're running the remover tool as Administrator


## You're All Set!

If you completed the development setup (Go + IDE + repo), you're ready for the workshop. The test machine is a nice-to-have for the final step, but don't let it hold you back from attending.

See you at the workshop!
