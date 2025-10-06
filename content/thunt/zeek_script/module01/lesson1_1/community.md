---
showTableOfContents: true
title: "Part 4 - The Zeek Community and Ecosystem"
type: "page"
---

## **PART 4: THE ZEEK COMMUNITY AND ECOSYSTEM**

### **The Foundation: Official Documentation and Resources**

One of Zeek's greatest strengths is its vibrant community and extensive ecosystem of resources, scripts, and tools. As you embark on your Zeek learning journey, understanding how to navigate these resources is just as important as understanding the tool itself. The community has developed an enormous body of knowledge, shared scripts, and collaborative tools that you'll leverage throughout your career working with Zeek.

Let's start with the official documentation, which serves as your primary reference throughout this course and beyond.
The Zeek documentation site, located at [docs.zeek.org](https://docs.zeek.org/en/master/), is comprehensive and well-organized, but it can be overwhelming when you're first starting out. Understanding its structure will help you find information quickly as you develop your skills.


The documentation is organized into several major sections, each serving a different purpose. The [Getting Started](https://docs.zeek.org/en/master/get-started.html) section provides introductory material about Zeek's concepts and basic operation. This is where you'll find explanations of how Zeek works, tutorials for common tasks, and guidance on your initial deployment. As you progress through this course, you'll reference this section less frequently, but it's invaluable for building your mental model of how Zeek operates.



The "Scripting" section is where you'll spend most of your time as you develop detection capabilities. This section documents Zeek's scripting language in detail, including data types, operators, control structures, and the vast library of built-in functions. Every time you're writing a script and need to know how a particular function works or what data types are available, you'll come here. The scripting documentation also includes information about Zeek's frameworks - structured collections of functionality for common tasks like threat intelligence integration, file analysis, and logging.

The "Frameworks" section deserves special mention because it documents some of Zeek's most powerful capabilities. The Intelligence Framework, which allows you to integrate threat intelligence feeds and automatically match them against network activity, is documented here. The Logging Framework, which you'll use to create custom log files, has its own comprehensive documentation. The Input Framework, used to read data from external files and databases, is explained in detail. As you tackle more sophisticated detection scenarios, you'll rely heavily on these frameworks.

Finally, the reference documentation provides complete details about every built-in event, function, type, and variable in Zeek. When you're writing scripts and need to know exactly what parameters a particular event provides, or what fields are available in a connection record, this is where you'll find that information. The reference documentation is exhaustive but can be dry - it's meant as a lookup resource rather than tutorial material.

Beyond the official documentation, the Zeek project maintains several other critical resources. The GitHub repository at github.com/zeek/zeek contains Zeek's source code and serves as the hub for development activity. If you encounter a bug or want to request a feature, this is where you'll file an issue. The repository also includes the issue tracker's history, which can be valuable when you're troubleshooting problems-chances are someone else has encountered a similar issue before.



### **The Power of Community: Packages and Shared Scripts**

One of Zeek's most valuable aspects is the community's culture of sharing detection scripts and protocol analyzers. Rather than each organization building everything from scratch, the community has developed a rich ecosystem of packages that extend Zeek's capabilities. Understanding how to find, evaluate, and leverage these packages will dramatically accelerate your threat hunting capabilities.

The central hub for community packages is [packages.zeek.org](https://packages.zeek.org), which serves as a searchable repository of contributed scripts and analyzers. At the time of this writing, the repository contains hundreds of packages covering everything from protocol parsers for obscure protocols to sophisticated detection logic for specific attack techniques. Many of these packages represent hundreds or thousands of hours of development effort, freely shared with the community.



Let's walk you through some of the most important packages that you'll likely use in your threat hunting work, because understanding what's available will inform how you approach detection problems later in this course.

The JA3 package, developed by researchers at Salesforce, is one of the most widely deployed community packages. JA3 provides fingerprinting of SSL/TLS clients and servers based on the parameters used during the TLS handshake. This might sound esoteric, but it has profound implications for threat detection. Many malware families use consistent TLS parameters when establishing encrypted connections, creating unique fingerprints that can identify them even when the traffic itself is encrypted. By generating JA3 hashes for TLS connections observed on your network and comparing them against databases of known-malicious fingerprints, you can detect malware communications over encrypted channels. This is particularly valuable given the increasing prevalence of encrypted malware command-and-control traffic.

Similarly, the HASSH package provides fingerprinting for SSH communications. SSH is another protocol where implementation details create unique fingerprints. Attackers who establish SSH backdoors or use SSH for lateral movement often have distinctive SSH fingerprints that differ from legitimate administrative tools. HASSH enables you to detect these anomalies.

The BZAR package, maintained by MITRE, focuses on detecting specific attack techniques from the ATT&CK framework. BZAR includes detection logic for techniques like credential dumping, lateral movement using SMB, and various reconnaissance activities. Rather than building these detections from scratch, you can deploy BZAR and immediately gain coverage for a range of common attack techniques. As you become more proficient with Zeek, you might customize BZAR's detections or use them as templates for your own detection logic.

Several packages focus on detecting specific types of threats. The zeek-httpattacks package includes signatures and behavioural detections for web application attacks like SQL injection and cross-site scripting. The zeek-log4j package was rapidly developed in response to the Log4j vulnerability and can detect exploitation attempts. The zeek-EternalSafety package detects exploitation attempts related to the EternalBlue vulnerability that was used in the WannaCry ransomware attack.

Other packages extend Zeek's protocol coverage. While Zeek includes native support for many common protocols, it can't possibly include every protocol in existence. The community has developed packages for analyzing protocols ranging from industrial control systems (Modbus, DNP3, BACnet) to messaging protocols (MQTT, AMQP) to various proprietary protocols. If you need to monitor a protocol that Zeek doesn't support natively, there's a good chance someone in the community has already developed a parser for it.

The Zeek Package Manager, called `zkg`, makes it straightforward to discover and install these packages. The command `zkg list` shows all available packages, `zkg search <keyword> `helps you find packages related to specific topics, and `zkg install <package-name>` installs a package and makes it available to Zeek. We'll work with `zkg` hands-on in a later lesson, but it's important to know now that this ecosystem exists and that you don't need to build everything yourself.



### **Threat Intelligence Integration: Leveraging Community Feeds**

A critical component of effective threat hunting is threat intelligence - knowledge about indicators of compromise, attacker techniques, and emerging threats. Zeek's Intelligence Framework provides powerful capabilities for integrating threat intelligence into your monitoring, and the community has made numerous high-quality intelligence feeds available for free or at low cost.

Understanding what intelligence feeds are available and how they can enhance your detection capabilities will inform how you approach threat hunting throughout this course. When you write detection scripts later, you'll often combine behavioural analysis with intelligence feed matching to create highly accurate detections with low false positive rates.

Let's explore some of the most valuable intelligence feed sources that you'll want to integrate with your Zeek deployment.

[Abuse.ch](https://abuse.ch) operates several freely available feeds that are particularly valuable for network security monitoring. Their [URLhaus](https://urlhaus.abuse.ch) feed provides real-time intelligence about malicious URLs being used for malware distribution. When your users browse the web, Zeek can automatically check accessed URLs against this feed and alert if anyone attempts to visit a known-malicious site.

The [Feodo Tracker](https://feodotracker.abuse.ch) feed from Abuse.ch focuses on command-and-control infrastructure for banking trojans and other malware families. These are IP addresses and domains actively being used by malware to communicate with attacker-controlled servers. By loading this feed into Zeek's Intelligence Framework, you can detect if any systems on your network are communicating with known C2 infrastructure - a strong indicator of compromise.

The [SSL Blacklist](https://sslbl.abuse.ch) feed, also from Abuse.ch, tracks malicious SSL certificates. When Zeek observes SSL/TLS connections, it extracts certificate details and can compare the certificate fingerprints against this blacklist. Malware often uses distinctive certificates for its encrypted communications, making certificate-based detection particularly effective.

AlienVault's [Open Threat Exchange (OTX)](https://otx.alienvault.com) is a community-driven threat intelligence platform that aggregates indicators from thousands of contributors worldwide. Security researchers, malware analysts, and security teams share intelligence about threats they've observed in the form of "pulses" - collections of related indicators describing specific threats or campaigns. OTX provides APIs that allow you to programmatically retrieve intelligence and feed it into Zeek. The indicators include IP addresses, domains, URLs, file hashes, and more, all tagged with information about the associated threats.

The [MISP Project](https://www.misp-project.org) (Malware Information Sharing Platform) is an open-source threat intelligence platform widely used by CERTs, security teams, and intelligence organizations for sharing structured threat information. Many organizations operate MISP instances to share intelligence within trusted communities. MISP can export intelligence in formats that Zeek's Intelligence Framework can consume, enabling automated sharing of threat intelligence between organizations and tools.




Commercial threat intelligence providers also offer feeds compatible with Zeek. Companies like Recorded Future, ThreatConnect, and others provide curated intelligence feeds that can be integrated through Zeek's Intelligence Framework or through custom scripts. While these commercial feeds require subscriptions, they often provide intelligence that's earlier, more accurate, or more comprehensive than free community feeds.

Beyond external feeds, many organizations generate their own internal threat intelligence from incident response activities, malware analysis, and threat research. When your team analyzes a malware sample and identifies its C2 infrastructure, those indicators can be fed into Zeek to detect other systems that might be compromised by the same malware. When you investigate a phishing campaign and identify the attacker's infrastructure, that intelligence can be used to detect future attacks from the same adversary.



---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./ecosystem.md" >}})
[|NEXT|]({{< ref "./exercises.md" >}})

