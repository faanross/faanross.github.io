---
title: "Section 1: Setting Up Our Virtual Environment"
description: ""
date: 2023-08-12T02:01:58+05:30
type: course
---


`|` [Course Overview](https://www.faanross.com/posts/course01/) `|` [Return to Section 0](https://www.faanross.com/course01/prebanter) `|` [Proceed to Section 2](https://www.faanross.com/course01/02_attack/) `|`

***

&nbsp;  


{{< figure src="/img/gif/thehacker.gif" title="" class="custom-figure" >}}

# 1.1. Introduction

In this section we'll set up the three VMs we'll need for the course - Windows 10 (Victim), Kali Linux (Attacker), and Ubuntu 20.04 (Post-Mortem Analysis). First we'll download the iso images and use them to install the operating systems. Then, depending on the specific VM, we'll perform some configurations as well as install extra software.



***

&nbsp; 

# 1.2. Requirements

I do want to give you some sense of the hardware requirements for this course, however I also have to add that I am not an expert in this area. ***AT ALL.*** So I'll provide an overview of what we'll be running, as well as what I think this translates to in terms of host resources (ie your actual system). But please - if you disagree with my estimation and believe you can get the same results by adapting the process, then please do so. After all - this is the *way of the hacker*. 

{{< figure src="/img/gif/tripleram.gif" title="" class="custom-figure" >}}

As mentioned above, we'll create 3 VMs in total, however, at any one moment there will only be a `maximum of 2 VMs running concurrently`. For each of these VMs I recommend the following system resources:
- min 2 (ideally 4) CPU cores
- min 4 (ideally 8) GB RAM
- around 60 GB HD space (allocated)

So based on this, that is roughly 2x the above + resources for your actual host system, you would likely need something along the lines of:
- 8 CPU cores (12+ even better)
- 16 GB RAM (32+ even better)
- 200 GB free HD space

{{< figure src="/img/gif/beefcake.gif" title="" class="custom-figure" >}}

Now I understand this requirement is rather beefy, but consider:
- You don't have to use a single system to run the entire VLAN - you could create an actual physical network, for ex with a Raspberry Pi cluster, and run the VMs on that. Or mini-pcs, or refurbished clients - really for a few hundred dollars you could more than easily be equipped to run a small network. I don't want to sound insensitive to a few 100 dollars, but I'm gonna level with you: `if you want to learn cybersecurity then there is no better investment than having localized resources to create virtual simulations`. 
- In case you don't want to invest up-front but don't mind paying some running costs: You can also use a service like [Linode](https://www.linode.com) and simply rent compute via the cloud. You can then install your VMs on that, and have access to them for as long as you care to foot the bill.

Finally I want to mention that beyond the hardware, `everything we will use is completely free`. This course ain't upselling a full course, and every piece of software is freely available. The sole exception has free alternatives, which I'm about to discuss with you right now. 

***

&nbsp; 


# 1.3. Hosted Hypervisor
So in the off-chance you don't know: a hosted (type 2) hypervisor is the software that allows us to run virtual machines on top of our base operating system. It's kinda like *Inception* - it allows us to create systems within our systems. 

{{< figure src="/img/gif/inception.gif" title="" class="custom-figure" >}}

For this course I'll be using [VMWare Workstation](https://store-us.vmware.com/workstation_buy_dual), which as of writing costs around $200. However you could also do it with either [VMWare Player](https://www.vmware.com/ca/products/workstation-player.html), or [Oracle Virtualbox](https://www.virtualbox.org/wiki/Downloads), both of which are free. 

I've used both `VMWare Player` and `VirtualBox` in the past, they mostly work well but running into some issues from time-to-time should not be completely unexpected. That being said, the problems I encountered were all, in hindsight, opportunities to learn. Frustrating - *feck yesh*. Enriching - sure. 

Since I switched over to `VMWare Workstation` my experience has been significantly more stable, so if you do have the money and are committed to this path as a career I would definitely consider getting it. That being said I don't wanna come across as some corporate shill, so really the choice is totally up to you.

{{< figure src="/img/gif/makechoice.gif" title="" class="custom-figure" >}}

Note that if you decide to not use `VMWare Workstation` then some of the details of the setup might be different. When that occurs it'll be up to you to figure out how to adapt it for your situation - Google, ChatGPT, StackExchange, common sense etc. Again, use the opportunities when things don't happen exactly "as they should" to learn. As a wise emperor once said - ***The impediment to action advances action. What stands in the way becomes the way.*** 

`So at this point please take a moment to download and install the hypervisor of your choice.`

Once that's done with feel free to proceed...

{{< figure src="/img/gif/pleasego.gif" title="" class="custom-figure-2" >}}

***

&nbsp; 

# 1.4. VM Images

Now that you have your hypervisor up and running the next thing we need to do is install our actual virtual machines. There are a few ways to do this, you can for example simply download the entire VM and simply import it into your hypervisor. This does usually mean that the file you'll be downloading will be quite large, so we'll opt for another approach - using iso files. You can think of an iso file simply as a "virtual copy" of the installation disc. So instead of importing the completed VM, we will be installing the VM ourselves using the iso image. 

So please go ahead and download the following 3 iso's:
* For the victim we'll use [Windows 10 Enterprise Evaluation 32-bit](https://info.microsoft.com/ww-landing-windows-10-enterprise.html). Note that MS will want you to register (it's free), so do so to download the iso OR [click here](https://techcommunity.microsoft.com/t5/windows-11/accessing-trials-and-kits-for-windows/m-p/3361125) to go to a Microsoft Tech Community post with direct download links. 
* For the attacker we'll use [Kali Linux](https://www.kali.org/get-kali/#kali-installer-images).
* For post-mortem analysis we'll be using [Ubuntu Linux Focal Fossa](https://releases.ubuntu.com/focal/). The reason being is in future courses we'll be using *RITA*, which, as of writing, runs best on *Focal Fossa*. 

Once you've successfully downloaded all three iso images we are ready to proceed. 

***

&nbsp; 

# 1.5. VM 1: Windows 10 aka "The Victim" 

{{< figure src="/img/gif/screamdrew.gif" title="" class="custom-figure" >}}
 
# 1.5.1. Installation

1. In VMWare Workstation goto `File` -> New Virtual Machine. 
2. Choose `Typical (recommended)`, then click `Next`. 
3. Then select `I will install the operating system later` and hit `Next`.

{{< figure src="/img/course01/image001.png" title="" class="custom-figure-2" >}}

4. Select `Microsoft Windows`, and under Version select `Windows 10`. 
5. Here you are free to call the machine whatever you'd like, in my case I am calling it `Victim`. 
6. Select 60 GB and `Split virtual disk into multiple files`. 
7. Then on the final screen click on `Customize Hardware`.

{{< figure src="/img/course01/image002.png" title="" class="custom-figure-2" >}}

8. Under `Memory` (see left hand column) I suggest at least 4096 MB, if possible given your available resources then increase it to 8192 MB. 
9. Under `Processors` I suggest at least 2, if possible given your available resources then increase it to 4.
10. Under `New CD/DVD (SATA)` change Connection from Use Physical Drive to `Use ISO image file`. Click `Browse…` and select the location of your Windows 10 iso file. 
11. Once done click `OK` on the bottom to exit out of the Hardware options dialog box. 

You should now see your VM in your Library (left hand column), select it and then click on `Power on this virtual machine`. If you don't see a Library column on the left simply hit `F9` which toggles its visibility.

Wait a short while and then you should see a Windows Setup window. Choose your desired language et cetera, select `Next` and then click on `Install Now`. Select `I accept the license terms` and click `Next`. Next select `Custom: Install Windows only (advanced)`, and then select your virtual HD and click Next.

Once its done installing we’ll get to the setup, select your region, preferred keyboard layout etc. Accept the `License Agreement` (if you dare - ***mwhahaha!***). Now once you reach the `Sign in` page don’t fill anything in, rather select `Domain join instead` in the bottom left-hand corner.

{{< figure src="/img/image006a.png" title="" class="custom-figure-2" >}}

Choose any username and password, in my case it'll be the highly original choice of `User` and `password`. Then choose 3 security questions, since this is a "burner" system used for the express purpose of this course don't overthink it - randomly hitting the keyboard a few times will do just fine. Turn off all the privacy settings, and for `Cortana` select `Not Now`.

Windows will now finalize installation + configuration, this could take a few minutes, whereafter you will see your desktop.

# 1.5.2. VMWare Tools
Next we'll install VMWare Tools which for our purposes does two things. First, it ensure that our VMs screen resolution assumes that of our actual monitor, but more importantly it also gives us the ability to copy and paste between the host and the VM. 

So just to be sure, at this point you should be staring at a Windows desktop. Now in the VMWare menu bar click `VM` and then `Install VMWare Tools`. If you open `Explorer` (in the VM) you should now see a `D:` drive. 

{{< figure src="/img/image008.png" title="" class="custom-figure-2" >}}

Double-click the drive, hit `Yes` when asked if we want this app to make changes to the device. Hit `Next`, select `Typical` and hit `Next`. Finally hit `Install` and then once done `Finish`. You'll need to restart your system for the changes to take effect, but we'll shut it down since we need to change a setting. So hit the Windows icon, Power icon, and then select `Shut Down`.

Right-click on your VM and select `Settings`. In the list on the LHS select `Display`, which should be right at the bottom. On the bottom - deselect `Automatically adjust user interface size in the virtual machine`, as well as `Stretch mode`, it should now look like this:

{{< figure src="/img/image009.png" title="" class="custom-figure-2" >}}

Go ahead and start-up the VM once again, we'll now get to configuring our VM.

# 1.5.3. Deep disable MS Defender + Updates

I call this 'deep disable' because simply toggling off the switches in `Settings` won't actually fully disable Defender and Updates. You see, Windows thinks of you as a younger sibling - it feels the need to protect you a bit, most of the time without you even knowing. (Unlike Linux of course which will allow you to basically nuke your OS kernel if you so desired.) 

{{< figure src="/img/winlin.png" title="" class="custom-figure" >}}

And just so you know why it is we're doing this...

We are disabling Defender so that the AV won't interfere with us attacking the system. Now you might think well this represents an unrealistic situation since in real-life we'll always have our AV running. Thing is, this is a simulation - we are simulating an actual attack. 

Yes, the AV might pick up on our mischievous escapades here since we are using a well-known and widely-used malware framework. But, if you are being attacked by an actual threat actor worth their salt they likely won't be using something so common. It's not that much of a stretch to assume they will be capable of using analogous technologies that our AV will not pick up on. Thus, by turning off Defender this is what we are simulating. 

And as for updates, we disable this because sometimes we can spend all this time configuring and setting things up and then one day we boot up our VM up, Windows does it's whole automatic update schpiel, and suddenly things are broken. This is thus a small time investment to hedge against extreme potential frustration. *So worth it*. 

{{< figure src="/img/do_it_now.gif" title="" class="custom-figure" >}} 

1. **Disable Tamper Protection**
    1. Hit the `Start` icon, then select the `Settings` icon.
    2. Select `Update & Security `.
    3. In LHS column, select `Windows Security`, then click `Open Windows Security`.
    4. A new window will pop up. Click on `Virus & threat protection`.
    5. Scroll down to the heading that says `Virus & threat protection settings` and click on `Manage settings`.
    6. There should be 4 toggles in total, we are really interested in disabling `Real-time protection`, however since we are here just go ahead and disable all of them. 
    7. Note that Windows will warn you and ask if you want to allow this app to make changes to the device, hit `Yes`.
    8. All 4 toggle settings should now be disabled.

{{< figure src="/img/image010.png" title="" class="custom-figure-2" >}}
    
2. **Disable the Windows Update service**
    1. Open the Run dialog box by pressing Win+R.
    2. Type `services.msc` and press Enter.
    3. In the Services list, find `Windows Update`, and double-click it.
    4. In the Windows Update Properties (Local Computer) window, under the `General` tab, in the `Startup type:` dropdown menu, select `Disabled` - see image below.
    5. Click `Apply` and then `OK`.
    
 {{< figure src="/img/image011.png" title="" class="custom-figure-2" >}}

3. **Disable Defender via Group Policy Editor**
    1. Open the Run dialog box by pressing Win+R.
    2. Type `gpedit.msc` and hit enter. The `Local Group Policy Editor` should have popped up.
    3. In the tree on the LHS navigate to the following: `Computer Configuration` > `Administrative Templates` > `Windows Components` > `Microsoft Defender Antivirus`.
    4. In the RHS double-click on `Turn off Microsoft Defender Antivirus`.
    5. In the new window on the top left select `Enabled` - see image below. 
    6. First hit `Apply` then `OK`.

 {{< figure src="/img/image012.png" title="" class="custom-figure-2" >}}

4. **Disable Updates via Group Policy Editor**
    1. Still in `Local Group Policy Editor`, navigate to: `Computer Configuration` > `Administrative Templates` > `Windows Components` > `Windows Update`.
    2. In the RHS double-click on `Configure Automatic Updates`.
    3. Select `Disabled`, then click `Apply` and `OK`.

5. **Disable Defender via Registry**
    1. In the search bar on the bottom type `cmd`.
    2. On the top left, right under `Best match` you should see `Command Prompt`.
    3. Right-click and select `Run as administrator`, hit `Yes`.
    4. Copy and paste the following command below into your command prompt and hit enter.
    ```
    REG ADD "hklm\software\policies\microsoft\windows defender" /v DisableAntiSpyware /t REG_DWORD /d 1 /f
    ```
    
Almost there! We just need to boot into Safe Mode to make some final adjustments to the registry and then we are good to go.

6. **Reboot system in Safe Mode**
    1. Open the `Run` dialog box by pressing Win+R.
    2. Write `msconfig` and hit enter.
    3. Select the `Boot` tab.
    4. Under `Boot options` select `Safe boot`, ensure `Minimal` is selected - see image below. 
    5. Hit `Apply` first, the `OK`.
    6. Select `Restart`.
    
{{< figure src="/img/image014.png" title="" class="custom-figure-2" >}}

Once your system has restarted in `Safe mode`...

7. **Disable Defender via Registry**
    1. Open the `Run` dialog box by pressing Win+R.
    2. Write `regedit` and hit enter, this should bring up the `Registry Editor`.
    3. Below you will see a list of 6 keys. For each of these keys you will follow the same process: once the key is selected find the `Start` value in the RHS, double-click, change the value to `4` and hit `OK` - see image below.
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SYSTEM` > `CurrentControlSet` > `Services` > `Sense`
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SYSTEM` > `CurrentControlSet` > `Services` > `WdBoot`
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SYSTEM` > `CurrentControlSet` > `Services` > `WinDefend`
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SYSTEM` > `CurrentControlSet` > `Services` > `WdNisDrv`
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SYSTEM` > `CurrentControlSet` > `Services` > `WdNisSvc`
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SYSTEM` > `CurrentControlSet` > `Services` > `WdFilter`

{{< figure src="/img/image015.png" title="" class="custom-figure-3" >}}

8. **Disable Updates via Registry**
    1. Still in `Registry Editor` let's navigate to the following:
    - `Computer` > `HKEY_LOCAL_MACHINE` > `SOFTWARE` > `Microsoft` > `Windows` > `CurrentVersion` > `WindowsUpdate` > `Auto Update`
    2. Right-click the `Auto Update` key, select `New`, and then click `DWORD (32-bit) Value`.
    3. Name the new key `AUOptions` and press Enter.
    4. Double-click the new `AUOptions` key and change its value to `2`. Click `OK` - see image below.
    5. Close Registry Editor.

{{< figure src="/img/image016.png" title="" class="custom-figure-3" >}}

9. **Leave Safe Mode**
    1. All that's left to do is get back into our regular Windows environment.
    2. Open the `Run` dialog box by pressing Win+R.
    3. Write `msconfig` and hit enter.
    4. Select `Boot` tab.
    5. Deselect `Safe boot`, hit `Apply`, hit `OK`.
    6. Hit `Restart`.

And that, I can promise you, is by far the most boring part of this entire course. But the good news is that this is potentially the last time you have to do it. *Ever*. This is because we can now convert this VM to a template and clone it indefinitely in the future. 

But before we learn to do that, let's setup all the awesome tools we'll be using in this course. 

# 1.5.4. Sysmon 

You should now be back in the normal Windows environment looking at your desktop. Let's set up ***Sysmon*** - a simple, free, Microsoft-owned program that will *dramatically* improve our logging ability. 

{{< figure src="/img/lumberjack.gif" title="" class="custom-figure" >}}

Before we install ***Sysmon*** there's just one thing you need to know - in addition to downloading ***Sysmon*** itself, we also need a config file. One day when you get to *that* level you can even create your own config file, which will allow you to make it behave exactly how you want it to. 

But for now, since we are decidedly not yet there, let's download and use one made by some really smart people. Of late  I have heard a few trusted sources, included [Eric Conrad](https://www.ericconrad.com) prefer [this version from Neo23x0](https://github.com/bakedmuffinman/Neo23x0-sysmon-config) whose authors included another blue team giant, [Florian Roth](https://twitter.com/cyb3rops?ref_src=twsrc%5Egoogle%7Ctwcamp%5Eserp%7Ctwgr%5Eauthor). 

So first download the [config file](https://github.com/bakedmuffinman/Neo23x0-sysmon-config), then [go here to download Sysmon](https://download.sysinternals.com/files/Sysmon.zip). You should now have two zip files - the config you downloaded from Github, as well as the ***Sysmon*** zip file. Extract the ***Sysmon*** archive, the contents should look as follows:

{{< figure src="/img/image017.png" title="" class="custom-figure-3" >}}

Now also extract the zip file containing the config. Inside of the folder rename `sysmonconfig-export.xml` to `sysmonconfig.xml`. Now simply cut (or copy) the file and paste it in the folder containing ***Sysmon***. 

Great, everything is set up so now we can install it with a simple command. Open command prompt as administrator and navigate to the folder containing ***Sysmon*** and the config file - in my case it is `c:\Users\User\Downloads\Sysmon`. Run the following command:

```
Sysmon.exe -accepteula -i
```

This is what a successful installation will look like:

{{< figure src="/img/image018.png" title="" class="custom-figure" >}}

Now let's just validate that it's running. In the command prompt run the command `powershell` so we change over into a PS shell. Then, run the command `Get-Service sysmon`. In the image below we can see it is running - we are good to go!

{{< figure src="/img/image019.png" title="" class="custom-figure" >}}

That's it for Sysmon, now let's enable PowerShell ScriptBlock logging. 

# 1.5.5. PowerShell ScriptBlock Logging

For security purposes, another quick and easy switch we can flip is enabling PowerShell logging. This is great because one specific type of PowerShell logs (`ScriptBlock`) will record exactly what command was run in PowerShell. As we know, in-line with the `Living off the Land` paradigm, modern adversaries LOVE abusing PowerShell. It should this be clear why having logs of commands that were run in PowerShell could potentially be of huge benefit to us. 

Something to be aware of is that there are a few types of PowerShell logging: Module, ScriptBlock, Operational, Transcription, Core, and Protected Event. For the purposes of this course we will be activating `ScriptBlock`, as well as `Operational`. While activating the former tells PowerShell to log the commands, we also need to activate `Operational` so that the system is able to properly save the logs. 

NOTE: This entire process could be performed in the GUI using `Group Policy Editor`, we will however be performing it via PowerShell command line. You should ***always*** prefer this method to using the GUI when it comes to enabling logs. Not simply to look cool, nay, there is a very good practical reason for this.

{{< figure src="/img/verycool.gif" title="" class="custom-figure" >}}

Imagine for a moment you needed to activate this feature on 1000 stations. You could either do so by logging into each station individually and interacting with the `gpedit` GUI interface, which would likely take you a few days working at a ferocious pace for all 1000 stations. Alternatively, you could run a single command from a domain controller, which would take less than a minute for any amount of stations. 

This is an admittedly dramatic way of saying that performing administrative tasks using PowerShell commands scales well, while flipping GUI toggles does not scale at all. So invest your time early on learning the methods that won't break down the moment you need to do it at scale, it's so worth it. 

{{< figure src="/img/worth.gif" title="" class="custom-figure" >}}

Open PowerShell as an administrator:
1. First we'll set the execution policy to allow us to make the changes:
```
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope LocalMachine
```
2. We'll now create a new registry path for `ScriptBlockLogging`:
```
New-Item -Path HKLM:\Software\Policies\Microsoft\Windows\PowerShell\ScriptBlockLogging -Force
```
3. Now create a new DWORD property EnableScriptBlockLogging and set its value to 1:
```
New-ItemProperty -Path HKLM:\Software\Policies\Microsoft\Windows\PowerShell\ScriptBlockLogging -Name EnableScriptBlockLogging -Value 1 -PropertyType DWord -Force
```
4. And finally we'll enable Operational logging to ensure our ScriptBlock logs are saved properly:
```
wevtutil sl "Microsoft-Windows-PowerShell/Operational" /e:true
```

# 1.5.6. Install Software

And now we'll install three programs:
- We'll use **Process Hacker** for live memory forensics. 
- We'll use **winpmem** to create a memory dump for post-mortem memory forensics. 
- We'll use **Wireshark** to generate a pcap for traffic analysis.

You can download [Process Hacker here](https://processhacker.sourceforge.io/downloads.php). Once downloaded go ahead and install it.

You can download the latest release of [winpmem here](https://github.com/Velocidex/WinPmem/releases). Since its a portable executable there is no installation required, just download the `.exe` file and place it on the desktop. 

And finally the `WireShark` setup file can be [downloaded from here](https://2.na.dl.wireshark.org/win32/Wireshark-win32-3.6.15.exe). Once downloaded run setup, just keep all options per default, nothing fancy required. 

That's it friend. We are done with BY FAR the heaviest lifting in terms of VM setup - the remaining two will comparatively be a breeze. But before we get to that there's one very simple thing we can do that will make our lives much easier in the future - turning this VM into a template for cloning.

# 1.5.7. Creating a Template 

So why do we want to do this again? Well by turning this VM we just created into a template we are in essence creating an archetype (blueprint). Then, whenever we want this same "victim" system for any project or course we can simply clone it as many times as we want. 

{{< figure src="/img/mememe.gif" title="" class="custom-figure" >}}

Thus instead of repeating this entire, rather cumbersome process we can click a few buttons and have it ready to go in a few seconds. This is also useful if we ever "mess up" the VM, we can just come back to this starting point.


So just follow along with these few simple steps:
1. First shut down the VM.
2. In VMWare you should see the library pane on the LHS listing our VM. If you don't, hit `F9`, or go to `View` > `Customize` > `Library`.
3. Right-click on our VM (`Victim`), select `Snapshot` > `Take Snapshot`.
4. Name it anything you'd like, I will be calling it `Genesis`. Hit `Take Snapshot`.
5. Again right-click the VM and select `Settings`. 
6. On the top left we can see two tabs - `Hardware` and `Options`, select `Options`.
7. Go down to the bottom and select `Advanced`.
8. Select `Enable Template mode (to be used for cloning)`, hit `OK`.

{{< figure src="/img/image021.png" title="" class="custom-figure-3" >}}

9. Note you might want to rename this VM to something like `Victim Template`, so we are aware this is the template that we should not be using, but rather use for cloning. You can do this under `Settings` > `Options` > `General`.
10. Now let's create our first clone which we will actually be using in the course. Right-click on `Victim Template`, select `Manage` > `Clone`. Hit `Next`.
11. We'll select the snapshot we created and hit `Next`. 
12. Keep selection as `Create a linked clone` and hit `Next`. 
13. Give your clone a name, I will be calling it `Victim01`. Choose a location and hit `Next`.

That's it! You should now see both `Victim Template` and `Victim01` in your library.

The bad news - we still have two VMs to install. The good news - they will require minimal-to-no configuration, so at this point we're about 80% done with our VM setup. So let's get it done.

***

&nbsp; 

# 1.6. VM 2: Kali Linux aka "The Attacker" 
{{< figure src="/img/attacker.gif" title="" class="custom-figure" >}}

We'll be using Kali Linux to simulate the attacker. The great thing about Kali Linux is that everything we'll need comes pre-packaged, so we just have to install the actual operating system. 

1. In VMWare hit `File` > `New Virtual Machine...`
2. `Typical (recommended)` and hit `Next`. 
3. `I will install the operating system later` and hit `Next`.
4. Select `Linux`, and under Version select `Debian 11.x 64-bit`. (Note: Kali Linux is built on top of Debian Linux).
5. Again call the machine whatever you'd like, in my case I am calling it `Hacker`. 
6. Increase the Maximum disk size to 60 GB and select `Split virtual disk into multiple files`. 
7. Then on the final screen click on `Customize Hardware`.
8. Under `Memory` I suggest at least 4096 MB, if possible given your available resources then increase it to 8192 MB. 
9. Under `Processors` I suggest at least 2, if possible given your available resources then increase it to 4.
10. Under `New CD/DVD (SATA)` change Connection from Use Physical Drive to `Use ISO image file`. Click `Browse…` and select the location of your Kali Linux iso image.

{{< figure src="/img/kali.gif" title="" class="custom-figure-2" >}}

So now let's get to actually installing it:
1. Right-click on the VM and select `Power` > `Start Up Guest`.
2. Select `Graphical Install`.
3. Select language, country etc.
4. Choose any `Hostname`, leave `Domain name` blank, for Full name and username I chose `hacker`.
5. Create a password, again though OBVIOUSLY not a suggested real-world practice, in these simulations I tend to simply use `password` since it minimizes any operational friction. 
6. Choose a timezone.
7. Next select `Guided - use entire disk` and hit `Continue`.
8. The only disk should be selected, hit `Continue`.
9. Keep `All files in one partition (recommended for new users)`, hit `Continue`.
10. Keep `Finish partitioning and write changes to disk`, hit `Continue`.
11. Select `Yes` and `Continue`.
12. In `Software selection` keep the default selection and hit `Continue`. Kali will now start installing, just be aware this can take a few minutes, probably around 5 to 10.
13. Next it'll ask you about installing a GRUB boot loader, keep it selected as `Yes` and hit `Continue`. 
14. Select `/dev/sda` and hit `Continue`. More installing... 
15. Finally it will inform us it's complete, we can hit `Continue` causing the system to reboot into Kali Linux. Enter your username and password and hit `Log In`.
16. Let's shut down the VM, then right-click on it in the library and select `Settings`. Under `Display` deselect `Stretch mode` and hit `OK`.

And that's it for our attacker machine - feel free to repeat the Template-Cloning process we performed for our Windows 10 VM if you so desire.

***

&nbsp; 

# 1.7. VM 3: Ubuntu Linux 20.04 aka "The Analyst" 
# 1.7.1. Installation
{{< figure src="/img/analysis.gif" title="" class="custom-figure-3" >}}

And now finally we'll set up our Ubuntu VM.

1. In VMWare hit `File` > `New Virtual Machine...`
2. `Typical (recommended)` and hit `Next`. 
3. `I will install the operating system later` and hit `Next`.
4. Select `Linux`, and under Version select `Ubuntu 64-bit`.
5. Again call the machine whatever you'd like, in my case I am calling it `Analyst`. 
6. Increase the Maximum disk size to 60 GB and select `Split virtual disk into multiple files`. 
7. Then on the final screen click on `Customize Hardware`.
8. Under `Memory` I suggest at least 4096 MB, if possible given your available resources then increase it to 8192 MB. 
9. Under `Processors` I suggest at least 2, if possible given your available resources then increase it to 4.
10. Under `New CD/DVD (SATA)` change Connection from Use Physical Drive to `Use ISO image file`. Click `Browse…` and select the location of your Ubuntu Linux 20.04 iso image. Make sure `Connect at power on` is enabled.
Click `Close` then `Finish`.

{{< figure src="/img/fossa.gif" title="" class="custom-figure-2" >}}

So now let's install Focal Fossa:
1. Right-click on the VM and select `Power` > `Start Up Guest`.
2. Select `Try or Install Ubuntu`.
3. Once it boots up the GUI, select `Install Ubuntu`.
4. Select your keyboard and language, hit `Continue`.
5. Keep `Normal Installation` selected, deselect `Download updates while installing Ubuntu`.
6. Keep `Erase disk and install Ubuntu` selected, then hit `Install Now`. 
7. For the popup asking if you want to `Write the changes to disks?`, hit `Continue`.
8. Choose a timezone and hit `Continue`.
9. Now fill in your name and desired credentials, I'll be using `analyst` and `password`.
10. When it's complete you can power the system off. Go into settings, under `CD/DVD (SATA)` disable `Connect at power on`.
11. Then goto `Display`, disable `Stretch mode`.
12. Hit `OK`, start the VM up once again, and log in.

`NOTE: A few moments after logging in and answer Ubuntu's questions you'll be asked whether you want to upgrade. IMPORTANT: Do not do so, decline the offer.`

{{< figure src="/img/image029.png" title="" class="custom-figure-2" >}}

OK, that's it for the installation, now let's install the two programs we'll use in this course. Note I'm also going to install ***RITA*** here, we won't use it in this course but indeed in the next one. So feel free to skip this, or just get it done with now so that next time everything is gtg.

# 1.7.2. Install Software
# Volatility3
1. Either download the zip file from the repo [here](https://github.com/volatilityfoundation/volatility3), or run the command below from terminal to clone the repo:
```
git clone https://github.com/volatilityfoundation/volatility3.git
```
2. Next we'll need to install ***pip***, which is a package manager for ***Python*** (***Volatility*** is written in ***Python***). We'll do this so we can install all the required package dependencies. Run the following commands:
```
sudo apt update
sudo apt install python3-pip
```
3. Once that's complete we can install our package dependencies. Open a terminal and navigate to where you cloned ***Volatility***. Now simply run the following command:
```
pip3 install -r requirements.txt
```

# WireShark
1. Run the following command to update the packet repository cache:
```
sudo apt update
```
2. Now run the following command to install ***WireShark***:
```
sudo apt install wireshark
```

# RITA (Optional)

Here's the cool thing about installing ***RITA***: when we do so it will also automatically install ***Zeek***, which is another amazing tool for traffic analysis we'll be using in the future. 

1. Goto the [RITA Github repo](https://github.com/activecm/rita).
2. Scroll down to `Install` and follow the instructions using the `install.sh` script. During installation you will be asked a few questions, answer `y` and hit enter each time. 
3. Let's check the version of RITA to ensure installation was successful. First close your terminal, reopen, and then run the commands seen in image below. You should get similar results. 

{{< figure src="/img/image030.png" title="" class="custom-figure" >}}

OK. Do you know what time it is? 

Yeah it's time for all this installing and configuring to pay off - let's kick things off by performing the attack!

{{< figure src="/img/strangelove.gif" title="" class="custom-figure" >}}

&nbsp;  

***

`|` [Course Overview](https://www.faanross.com/posts/course01/) `|` [Return to Section 0](https://www.faanross.com/course01/prebanter) `|` [Proceed to Section 2](https://www.faanross.com/course01/02_attack/) `|`