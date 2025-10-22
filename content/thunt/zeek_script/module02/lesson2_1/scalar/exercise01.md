---
showTableOfContents: true
title: "Scalar Types Practical Exercises"
type: "page"
---

## Exercise 1: Port Scan Detection

### Objective

Detect when a single host connects to many different ports on another host within a short time window - the signature of a port scan.

### The Attack Scenario

Port scanning is reconnaissance - an attacker probing which services are running on a target. Tools like [Nmap](https://nmap.org) systematically try connections to many ports to map the attack surface. This is almost always the first phase of an attack.

### The Detection Logic

We'll track how many unique destination ports each source IP contacts on each destination IP. When this count exceeds a threshold (let's say 20 ports), we alert.

**Key types used:**

- `count`: Track number of ports contacted
- `addr`: Identify source and destination IPs

### Write the Script

Create the detection script:

```bash
sudo nano /opt/zeek/share/zeek/site/custom/scan-detection.zeek
```

For complete code solutions see [HERE](https://github.com/faanross/zeek_scripting_course/blob/main/module02/lesson02_01/scalar/port_scan/scan-detection.zeek).

For complete code breakdown see [HERE](https://github.com/faanross/zeek_scripting_course/blob/main/module02/lesson02_01/scalar/port_scan/scan-detection-breakdown.md).


### Verify and Load our Script

**Verify our syntax:**
```bash
zeek -a /opt/zeek/share/zeek/site/custom/scan-detection.zeek
```

No news is good news.


**Load it in `local.zeek`:**
```bash
sudo nano /opt/zeek/share/zeek/site/local.zeek
```

**Add at the bottom:**

```zeek
@load ./custom/scan-detection.zeek
```



### Let's Now Run a Single Instance of Zeek

Please note that when you run a single instance of zeek (i.e. using zeek instead of zeekctl), it will write the logs to the directory you are running from NOT `/opt/zeek/logs/current/`.

So just create a new directory for our test to run from

```bash
mkdir test_scan
cd ./test_scan

# assuming your interface is eth0, if not run `ip a` to find out what it is
sudo zeek -C -i eth0 local.zeek
```


You should immediately see our debug output:
```bash
zeek@zeek-sensor:~/test_scan$ sudo zeek -C -i eth0 local.zeek
listening on eth0

=== SCAN DETECTION SCRIPT LOADED ===
Threshold: 20 ports
Tracking window: 5 minutes of inactivity
Custom log 'port_scan.log' initialized
Watching for port scans...
```

If you don't see it, it probably means you forgot to load it in `local.zeek`

I'll be running the scan from another machine, if you want to do the same run `ip a` to get your public IP.



### Generate the Attack

We'll use **Nmap** to perform a real port scan:

```bash
# Install nmap if needed
sudo apt install nmap
```

Let's scan ports 1 to 100:
```bash
sudo nmap -p 1-100 <public IP of target machine>
```



### Detection

You should see the scans appear in the terminal where Zeek is running and eventually the alert too:

```
*** SCAN DETECTED! 90.29.239.33 hit 20 ports on 136.97.124.83 ***
Alert written to notice.log and port_scan.log
```


And now we can also confirm it was written to both notice.log and port_scan.log. Just please keep in mind these files will be in `pwd`, NOT the daemon's log directories in `/opt`.


```bash
zeek@zeek-sensor:~/test_scan$ cat notice.log
#separator \x09
#set_separator	,
#empty_field	(empty)
#unset_field	-
#path	notice
#open	2025-10-19-08-14-22
#fields	ts	uid	id.orig_h	id.orig_p	id.resp_h	id.resp_p	fuid	file_mime_type	file_desc	proto	note	msg	sub	src	dst	p	n	peer_descr	actions	email_dest	suppress_for	remote_location.country_code	remote_location.region	remote_location.city	remote_location.latitude	remote_location.longitude
#types	time	string	addr	port	addr	port	string	string	string	enum	enum	string	string	addr	addr	port	count	string	set[enum]	set[string]	interval	string	string	string	double	double
1760876062.226688	-	-	-	-	-	-	-	-	-	Scanning::Port_Scan_Detected	90.29.239.33 scanned 20 ports on 136.97.124.83	-	90.29.239.33-	-	-	-	Notice::ACTION_LOG	(empty)	3600.000000	-	-	-	-	-
```

```bash
zeek@zeek-sensor:~/test_scan$ cat port_scan.log
#separator \x09
#set_separator	,
#empty_field	(empty)
#unset_field	-
#path	port_scan
#open	2025-10-19-08-14-22
#fields	ts	src	dst	port_count
#types	time	addr	addr	count
1760876062.226688	70.27.237.44	138.197.134.43	20
```

---







---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./first.md" >}})
[|NEXT|]({{< ref "./conclusion.md" >}})

