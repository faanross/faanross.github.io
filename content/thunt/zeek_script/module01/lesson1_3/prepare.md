---
showTableOfContents: true
title: "Part 1 - Preparing Your Installation Environment"
type: "page"
---
## **From Theory to Reality: Building Your First Zeek Sensor**

Welcome to what many students consider the most exciting lesson in Module 1 - the moment where abstract concepts become a working system. Over the next few hours, you're going to build a fully functional Zeek sensor from the ground up. You'll install the software, configure it for optimal performance, and capture your first network traffic. By the end of this lesson, you'll have a working platform for all the threat hunting and detection work that follows in later modules.

This lesson is different from the previous two. While Lessons 1.1 and 1.2 were primarily conceptual, this lesson is intensely practical. You'll be working in a terminal, executing commands, editing configuration files, and troubleshooting issues. This hands-on experience is crucial because Zeek isn't something you learn purely from reading - you need to feel how it works, see it in action, and develop the muscle memory that comes from actually operating the system.

We're going to take a methodical approach to installation. Rather than just running through a quick install script, we'll explore three different installation methods - package manager installation, compilation from source, and container deployment. Each method has different characteristics and use cases, and understanding all three will make you a more versatile Zeek operator. We'll examine Zeek's file system layout in detail so you understand where everything lives and why. We'll configure network interfaces for monitoring, set up high-performance packet capture, and implement best practices for production deployments.

This is a long lesson with many steps. Take your time, read carefully, and don't skip ahead. Each section builds on the previous one, and understanding why we're doing something is just as important as knowing how to do it.

Let's begin.

---


## **PART 1: PREPARING YOUR INSTALLATION ENVIRONMENT**

### **Understanding Your Installation Environment**

Before we install anything, let's understand the environment we're working with and prepare it properly. You can run Zeek on any platform that suits your needs: a cloud VM from any provider (Digital Ocean, AWS, Azure, etc.), a Type 1 or Type 2 hypervisor, or even a bare metal Linux installation. Use whatever you have available.

For this course, I'll be using a Digital Ocean droplet, but the principles apply regardless of where you're running your system. You should have a system with at least 4 vCPUs, 8GB RAM, and 80GB of storage. If you're using a cloud VM and haven't provisioned it yet, pause here and create it now.

Zeek runs on numerous Linux distributions, including Ubuntu, Red Hat, Rocky, and others. I'll be using Ubuntu 24.04 LTS in all examples because, in my opinion, it's currently the most stable OS for Zeek and will be supported until 2029. It strikes the right balance between modern software packages and being a long-term support release that won't surprise you with breaking changes.

You're welcome to use other versions like 22.04, or different distributions entirely, but be aware that if you deviate from Ubuntu 24.04, you may need to fill in any gaps or handle distribution-specific differences that arise during the course.


**Why Ubuntu 22.04 LTS specifically?**

```
┌──────────────────────────────────────────────────────────────┐
│         UBUNTU 24.04 LTS ADVANTAGES FOR ZEEK                 │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  ✓ Long-term support until 2029                              │
│    (No need to upgrade for 3.5 years)                        │
│                                                              │
│  ✓ Modern kernel with AF_PACKET improvements                 │
│    (Better packet capture performance)                       │
│                                                              │
│  ✓ Recent compiler versions                                  │
│    (GCC 11+, needed for optimal Zeek compilation)            │
│                                                              │
│  ✓ Large community and extensive documentation               │
│    (Easy to find solutions to problems)                      │
│                                                              │
│  ✓ Official Zeek packages available                          │
│    (Can install via apt without compilation)                 │
│                                                              │
│  ✓ SystemD for service management                            │
│    (Modern, reliable daemon management)                      │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```






---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "../lesson1_2/validation.md" >}})
[|NEXT|]({{< ref "./package.md" >}})

