---
showTableOfContents: true
title: "The string Type: Text Data"
type: "page"
---




## The string Type: Text Data

The `string` type represents text - sequences of characters that make up hostnames, URLs, user agents, email addresses, file names, HTTP headers, DNS queries, and countless other pieces of textual data flowing through network protocols. In network security analysis, strings are everywhere because most application-layer protocols use human-readable text. Understanding how to work with strings effectively is essential for building detections that analyze protocol content rather than just packet headers.

### **Why Strings Matter in Security**

While lower-level network analysis focuses on IP addresses and port numbers, sophisticated security detection examines the actual content of communications. Is this URL trying to exploit a vulnerability? Does this User-Agent string match known malware? Is this hostname trying to masquerade as a legitimate domain? Does this HTTP request contain SQL injection attempts?

All of these questions require analyzing textual data. Attackers embed malicious code in URLs, disguise malware with convincing User-Agent strings, use domain names that look legitimate but contain subtle typos (typosquatting), and hide exploits in seemingly normal text fields. The `string` type gives you the tools to inspect, compare, search, and manipulate this textual content to uncover threats.

### **Basic Usage and Declaration**

Working with strings in Zeek is straightforward and similar to most programming languages:

```c
# String literals
local hostname: string = "www.example.com";
local user_agent: string = "Mozilla/5.0...";
local empty: string = "";

# Strings from events
event http_request(c: connection, method: string, original_URI: string,
                   unescaped_URI: string, version: string)
{
    local url: string = original_URI;
    local host: string = c$http$host;
}
```

String literals use double quotes. Most of the time, though, you won't be declaring string constants - you'll be extracting strings from network events. Every HTTP request contains multiple string fields: the method ("GET", "POST"), the URI, the hostname, the User-Agent header, cookies, and more. DNS queries contain domain names. SMTP traffic contains email addresses. These all come to you as `string` values ready for analysis.

### **Essential String Operations**

Zeek provides a rich set of operations for working with textual data:

**Concatenation** builds new strings by joining existing ones:

```c
# Concatenation
local full_url = "http://" + hostname + "/path";
```

This is useful for reconstructing full URLs, building log messages, or creating composite identifiers.

**Comparison** checks for exact matches:

```c
# Comparison
if ( hostname == "www.evil.com" )
    print "Matched malicious domain";
```

Exact comparison is the foundation of allow/deny lists - checking if a string matches a known good or known bad value.

**Length** tells you how many characters a string contains:

```c
# Length
local len = |hostname|;  
# Number of characters
```

String length is useful for detecting anomalies - excessively long URLs might indicate buffer overflow attempts, unusually short hostnames might be suspicious, and zero-length fields might indicate protocol violations.

**Substring checking** searches for patterns within strings:

```c
# Substring check
if ( "evil" in hostname )
    print "Suspicious string found";
```

The `in` operator checks if one string appears anywhere within another. This is simpler than regular expressions for straightforward substring matching.

**Case conversion** normalizes strings for comparison:

```c
# "WWW.EXAMPLE.COM" -> "www.example.com"
local lower = to_lower(hostname); 

# "www.example.com" -> "WWW.EXAMPLE.COM" 
local upper = to_upper(hostname);  
```

Case conversion is essential because attackers often use mixed case to evade simple string matching. Converting everything to lowercase before comparison prevents evasion through capitalization tricks.

**String formatting** creates readable messages:

```c
# String formatting
local message = fmt("User %s from %s accessed %s", 
                    username, src_ip, url);
```

The `fmt()` function works like printf in C or format strings in Python - placeholders like `%s` (string), `%d` (integer), `%f`(float) get replaced with actual values. This makes log messages and alerts readable and informative.

### **Regular Expressions: The Power Tool**

Regular expressions are pattern-matching mini-languages that let you express complex textual patterns concisely. They're incredibly powerful for security detection because attacks often follow recognizable patterns:

**Pattern matching for common exploits:**

```c
# Pattern matching - VERY powerful for detection
local url: string = "/admin/../../etc/passwd";

# Check for path traversal
if ( /\.\.[\/\\]/ in url )
    print "Path traversal detected!";

# Match SQL injection patterns
if ( /union.*select|or.*1=1|'; drop/i in url )
    print "SQL injection detected!";
```

Let's break down these patterns:

- `/\.\.[\/\\]/` matches ".." followed by either a forward slash or backslash - the classic path traversal pattern trying to escape directory boundaries
- `/union.*select|or.*1=1|'; drop/i` matches common SQL injection patterns: "union" followed eventually by "select", or "or" followed by "1=1", or "'; drop". The trailing `i` makes the match case-insensitive.

**Extracting parts of strings:**

```c
# Extract parts of strings
local email = "user@example.com";
local parts = split_string(email, /@/);
# parts[0] = "user", parts[1] = "example.com"
```

The `split_string()` function breaks a string into pieces based on a delimiter (here, the `@` symbol). This is useful for parsing structured text - email addresses, URLs, CSV data, or any delimited format.

**File extension checking:**

```c
# Check for suspicious file extensions
if ( /.exe$|.scr$|.bat$/ in filename )
    print "Executable file detected";
```

The `$` anchor matches the end of the string, ensuring these extensions are actually at the end of the filename, not embedded in the middle.

Regular expressions deserve deep study - they're one of the most powerful tools in your security detection arsenal. The patterns you can express range from simple substring matches to complex multi-condition rules that would take dozens of lines of code to implement manually.

### **Real-World Example: User-Agent Analysis**

User-Agent strings identify the browser or application making HTTP requests. Legitimate browsers have characteristic User-Agent formats, while malware, scanners, and automation tools often use distinctive patterns or unusual User-Agents. Let's build detection logic around this:

```c
# Detect suspicious user agents
global legitimate_browsers = set(
    "Mozilla",
    "Chrome", 
    "Safari",
    "Firefox",
    "Edge"
);

event http_request(c: connection, method: string, original_URI: string,
                   unescaped_URI: string, version: string)
{
    if ( !c$http?$user_agent )
        return;
    
    local ua = c$http$user_agent;
    local suspicious = T;
    
    # Check if it looks like a legitimate browser
    for ( browser in legitimate_browsers )
    {
        if ( browser in ua )
        {
            suspicious = F;
            break;
        }
    }
    
    # Also check for known malicious patterns
    if ( /curl|wget|python|powershell|scanner/i in ua )
        suspicious = T;
    
    if ( suspicious )
    {
        print fmt("Suspicious User-Agent: %s from %s",
                  ua, c$id$orig_h);
    }
}
```

**Understanding this detection:** We're analyzing every HTTP request's User-Agent header. First, we assume it's suspicious until proven otherwise. Then we check if it contains any of the strings typical of legitimate browsers - "Mozilla", "Chrome", etc. Most real browsers include "Mozilla" for historical compatibility reasons, so this catches most legitimate traffic.

However, we then apply an additional check for known tool patterns. Command-line tools like `curl` and `wget`, scripting languages like `python`, shells like `powershell`, and explicit scanner tools often appear in User-Agent strings when attackers use automation. If we detect these patterns, we flag it as suspicious regardless of browser string presence.

This isn't perfect - legitimate automation exists, and attackers can forge User-Agents - but it's a useful signal. Unusual User-Agents warrant closer inspection, especially when combined with other suspicious indicators like accessing sensitive paths or generating unusual traffic patterns.

### **String Safety and Sanitization**

Strings from network traffic are fundamentally **untrusted input**. Attackers control this data and may craft it maliciously. You must handle strings carefully to avoid security issues in your own scripts:

```c
# Always sanitize strings before using in external commands or logs
function sanitize_string(s: string): string
{
    # Remove potentially dangerous characters
    local safe = gsub(s, /[^a-zA-Z0-9._-]/, "_");
    return safe;
}

# Truncate long strings to prevent log bloat
function truncate_string(s: string, max_len: count): string
{
    if ( |s| <= max_len )
        return s;
    
    return s[0:max_len] + "...";
}
```

**Why sanitization matters:** If you're writing strings to log files, you need to ensure they don't contain characters that could break your log format or inject false log entries (like newlines). If you're passing strings to external programs (generally discouraged), you must prevent command injection. The `gsub()` function (global substitute) replaces all characters that don't match the safe pattern with underscores.

**Why truncation matters:** Attackers sometimes send extremely long strings (kilobytes or even megabytes) to exploit buffer overflows or cause denial of service. Even if your Zeek script handles them safely, logging these enormous strings can fill your disk, slow down log processing, and make analysis difficult. Truncating strings to reasonable lengths (maybe 200-500 characters) keeps logs manageable while preserving enough context for analysis.

### **Why This Matters for Security**

The `string` type is your window into application-layer protocols - the actual content attackers manipulate. While network-layer analysis (IPs, ports, connection patterns) catches broad categories of threats, string analysis detects sophisticated attacks that operate within legitimate protocols: SQL injection hidden in HTTP parameters, cross-site scripting in URLs, path traversal in file paths, command injection in form fields, malware callbacks disguised as browser traffic.

Effective string handling combines multiple techniques:

- **Exact matching** for known bad values (malicious domains, exploit signatures)
- **Regular expressions** for pattern-based detection (attack techniques that vary in detail but follow recognizable patterns)
- **Substring searching** for simple indicators (keywords associated with threats)
- **Length checks** for anomaly detection (unusually long or short values)
- **Case normalization** to prevent evasion
- **Sanitization** to protect your own systems

As you develop more advanced Zeek scripts, you'll find yourself analyzing strings constantly - parsing URLs to extract suspicious components, correlating hostnames with threat intelligence, detecting encoded or obfuscated attack payloads, and building signatures for emerging threats. The `string` type and its associated operations are fundamental tools you'll use in almost every detection you build.



## Knowledge Check: string Type

**Q1: Why is the string type so important in network security analysis when lower-level analysis focuses on IPs and ports?**

A: Sophisticated attacks operate within legitimate protocols at the application layer. While IP/port analysis catches broad threats, string analysis detects attacks embedded in protocol content: SQL injection in HTTP parameters, cross-site scripting in URLs, path traversal in file paths, command injection in form fields, malware callbacks disguised as browser traffic. Attackers manipulate textual data, so analyzing strings is essential for catching these content-based threats.

**Q2: What's the difference between substring checking with the "in" operator and pattern matching with regular expressions? When would you use each?**

A: Substring checking (`"evil" in hostname`) simply tests if one string appears anywhere within another - it's straightforward and fast. Regular expressions (`/\.\.[\/\\]/ in url`) match complex patterns with wildcards, alternatives, and structure - they're more powerful but more complex. Use substring checking for simple, literal matches (known bad domains, specific keywords). Use regular expressions for pattern-based detection where attacks vary in details but follow recognizable patterns (SQL injection, path traversal, exploit signatures).

**Q3: Why should you always sanitize and truncate strings from network traffic before using them in logs or external commands?**

A: Strings from network traffic are untrusted input controlled by attackers. Without sanitization, they could contain characters that break log formats or inject false log entries (like newlines). Without truncation, attackers could send extremely long strings (kilobytes/megabytes) to fill your disk, slow log processing, or exploit buffer overflows. Sanitization prevents format breaking and injection; truncation prevents resource exhaustion while preserving enough context for analysis.

**Q4: What is the purpose of case conversion (to_lower/to_upper) in security detection, and why is it necessary?**

A: Case conversion normalizes strings for comparison because attackers often use mixed case to evade simple string matching. For example, "EvIl.CoM" would bypass a check for "evil.com" without case normalization. Converting everything to lowercase (or uppercase) before comparison prevents this evasion technique. It's necessary because domain names, URLs, and many protocol fields are case-insensitive, but string comparison is case-sensitive by default.

**Q5: Describe what makes User-Agent analysis valuable for detecting malicious activity.**

A: Legitimate browsers have characteristic User-Agent formats (containing "Mozilla", "Chrome", etc.), while malware, scanners, and automation tools often use distinctive patterns. Command-line tools (curl, wget), scripting languages (python), shells (powershell), and scanner tools often appear in User-Agents when attackers use automation. While not perfect (attackers can forge User-Agents and legitimate automation exists), unusual User-Agents are a valuable signal, especially when combined with other suspicious indicators like accessing sensitive paths or generating unusual traffic patterns.



---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./interval.md" >}})
[|NEXT|]({{< ref "./bool.md" >}})

