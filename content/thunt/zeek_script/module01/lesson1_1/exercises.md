---
showTableOfContents: true
title: "Part 5 - Practical Exercises"
type: "page"
---
## **PART 5: PRACTICAL EXERCISES**

Now it's time to move from theory to practice. The exercises in this section are designed to familiarize you with the resources we've discussed and help you begin building your personal knowledge base about Zeek. Unlike later lessons where you'll be writing scripts and analyzing traffic, these exercises focus on navigation, research, and preparation. They're building the foundation you'll need for the hands-on work ahead.

### **Exercise 1: Deep Dive into Official Documentation**

Your first task is to become comfortable navigating Zeek's official documentation. This isn't about reading every page - that would take weeks. Instead, you're going to explore the documentation's structure and locate specific types of information. This skill will save you countless hours as you develop detection capabilities later in the course.

Open your web browser and navigate to [docs.zeek.org](https://docs.zeek.org). Take a few minutes to explore the main navigation menu and understand how the documentation is organized. You'll see sections for "Getting Started," "Scripting," "Frameworks," and reference documentation.


Now let's find specific information to understand how this documentation works. Navigate to the Scripting section and locate the page that describes Zeek's built-in data types. You should find descriptions of types like addr (representing IP addresses), port (representing network ports), time (representing timestamps), and many others. Read through the description of the addr type. Notice how the documentation explains not just what the type is, but how to use it, what operations it supports, and common patterns for working with it.

Next, let's explore the event reference. In many places throughout the documentation, you'll see references to events that Zeek generates. Find the reference documentation for events and browse through the list. It's quite extensive - Zeek generates events for dozens of protocols and network activities. Locate the `connection_state_remove` event. This event fires when a connection terminates, and you'll use it frequently in your detection scripts. Read its documentation and notice what information is provided to your script when this event fires.



Now let's look at the Framework documentation. Navigate to the Intelligence Framework documentation and read through the overview. Even though you won't be writing intelligence framework scripts yet, understanding how the framework operates will inform your threat hunting strategy. Notice how the framework allows you to specify indicator types, matching conditions, and actions to take when indicators are observed.

Finally, familiarize yourself with the reference documentation's structure. Find the page that documents the connection record -a data structure that represents a network connection. This record is passed to many events and contains fields describing the connection. Read through the field descriptions and note the wealth of information available about each connection: source and destination addresses and ports, protocol, service identification, duration, byte counts, connection state, and more.

Your deliverable for this exercise is to document your findings. Create a text file or document where you write down the following:

1. Three events that would be useful for detecting command-and-control traffic. For each event, write a sentence explaining why it would be useful. Think about what characteristics of C2 traffic these events would help you observe.
2. Five fields from the connection record that would be valuable for behavioural analysis. Explain what each field represents and how you might use it in threat hunting.
3. Two frameworks that seem particularly relevant for your threat hunting goals, with a brief note about what each framework does.

This exercise should take you about thirty minutes. Don't rush - the goal is to become comfortable finding information in the documentation, not to memorize everything you see.

### **Exercise 2: Exploring the Package Ecosystem**

Now let's explore the community package repository and identify tools that will support your threat hunting objectives. Open your browser and navigate to [packages.zeek.org](https://packages.zeek.org). You'll see a list of available packages, each with a brief description.



Start by using the search functionality to find packages related to specific topics.
- Search for "malware" and see what comes up. You should find packages designed to detect various types of malware behaviour.
- Search for "TLS" or "SSL" and examine the packages related to encrypted traffic analysis.
- Search for "HTTP" and look at packages focused on web traffic analysis.

Now let's investigate some specific packages in detail. Find the JA3 package and click through to its full description. You'll typically find a link to the package's GitHub repository where you can see the actual code, read more detailed documentation, and understand how to use it. Read about what JA3 fingerprinting is and how it works. Even if you don't understand all the technical details yet, grasp the core concept: JA3 creates a fingerprint of TLS clients based on the parameters they use, allowing identification of malware even in encrypted traffic.

Locate the BZAR package. Read about how it implements detections for ATT&CK techniques. Visit its GitHub repository and look at the README file. Notice how the package documentation explains what it detects and provides examples of the alerts it generates. This is the kind of documentation you should aim for when you eventually share your own scripts.


Your deliverable for this exercise is a list of five packages you plan to install and use with your Zeek sensor. For each package, document:

1. The package name and where to find it
2. What problem it solves or what capability it provides
3. Why it's relevant to your threat hunting objectives
4. Any prerequisites or dependencies it has

This exercise should take about twenty to thirty minutes. The goal is to understand what the community has already built so you can leverage that work rather than reinventing solutions.




### **Exercise 3: Threat Intelligence Feed Research**

Let's explore the threat intelligence feeds that you can integrate with Zeek. Understanding what intelligence is available will shape how you approach detection - many sophisticated detection strategies combine behavioural analysis with intelligence feed matching.

Start by visiting abuse.ch, one of the most valuable sources of free threat intelligence. Navigate to their URLhaus project and examine the feed. Click through to see examples of the malicious URLs they track. Notice the level of detail provided: the URL itself, the malware family it's associated with, when it was first observed, the threat type, and more. Download a sample of the feed data and examine the format. You'll notice it's structured in a way that can be consumed by automated tools like Zeek.

Now explore their Feodo Tracker feed, which focuses on C2 infrastructure. Look at the types of information provided about each C2 server: the IP address or domain, the port, the malware family, the first and last times it was observed online, and confidence levels. This is exactly the kind of intelligence that's valuable for detecting compromised systems communicating with C2 infrastructure.

Visit AlienVault's Open Threat Exchange at otx.alienvault.com. You'll need to create a free account to fully explore the platform. Once logged in, browse through recent "pulses" - curated collections of indicators related to specific threats. Look for pulses related to RATs. A typical pulse might include IP addresses used for C2, domains registered by the attackers, file hashes of malware samples, and YARA rules for detecting the malware. Notice how different indicator types work together to provide comprehensive intelligence about a threat.

Finally explore the MISP Project website at misp-project.org. While you won't set up a full MISP instance right now, understanding what MISP is and how organizations use it will inform your intelligence strategy. Read about how MISP enables structured threat intelligence sharing and how communities use it to collaborate on threat information. Note that many organizations and CERTs operate MISP instances that share intelligence feeds in formats Zeek can consume.

Your deliverable for this exercise is documentation of three threat intelligence feeds you plan to integrate with Zeek. For each feed, document:

1. The feed source and how to access it
2. What types of indicators it provides (IPs, domains, URLs, file hashes, etc.)
3. How frequently it's updated
4. How you'll integrate it with Zeek (we'll cover the technical details later, but make note of the feed format and delivery mechanism)
5. Why this feed is particularly valuable for your threat hunting objectives

This exercise should take about twenty to thirty minutes. The goal is to identify intelligence sources that will enhance your detection capabilities before you begin writing scripts.



---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./community.md" >}})
[|NEXT|]({{< ref "./validation.md" >}})

