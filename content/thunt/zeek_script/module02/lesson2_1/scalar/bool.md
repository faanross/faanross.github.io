---
showTableOfContents: true
title: "The bool Type: Boolean Values"
type: "page"
---








## **The bool Type: Boolean Values**

The `bool` type represents binary truth values - something is either true or false, yes or no, on or off. Booleans are the fundamental building blocks of logic and decision-making in programs. In Zeek, true is written as `T` and false as `F` (note the capital letters, unlike many languages that use lowercase `true` and `false`). 

While this might seem like a simple type, booleans are essential for expressing security logic, for example: 
- "Is this IP address suspicious?" 
- "Has the threshold been exceeded?" 
- "Should we alert on this behavior?"

### **Why Booleans Matter in Security**

Security detection is fundamentally about making decisions based on conditions. Every detection you write ultimately answers a yes/no question: "Is this traffic malicious?" "Does this behavior match a known attack pattern?" "Have multiple indicators aligned to suggest a threat?"

Booleans let you express these binary decisions clearly and combine them using logical operations. A sophisticated detection might check: "Is the source internal AND the destination external AND the connection is encrypted AND the volume is high AND the timing is suspicious?" Each of these conditions evaluates to a boolean, and you combine them with logical operators to make a final determination.

Boolean flags also let you track state over time: "Have we seen scanning from this IP?" "Is this connection still active?" "Has this alert already fired?" This state tracking is essential for building detections that aggregate information across multiple events.

### **Basic Usage and Declaration**

Working with booleans in Zeek is straightforward:

```c
# Boolean literals
local is_suspicious: bool = T;
local is_encrypted: bool = F;

# From comparisons
local is_local: bool = (ip in local_networks);
local exceeded_threshold: bool = (count > 100);

# Boolean operations
# AND
local detected = is_suspicious && exceeded_threshold;  

# OR
local flagged = is_local || is_external; 

# NOT 
local not_safe = !is_encrypted;  
```

**Direct assignment** gives you explicit true/false values using `T` and `F`. Remember these are capitalized in Zeek - lowercase won't work.

**From comparisons**, any comparison operation produces a boolean result. Testing if an IP is in a subnet, checking if a count exceeds a threshold, comparing strings - all of these evaluate to `T` or `F`.

**Boolean operators** let you combine simple conditions into complex logic:

- `&&` (AND) returns true only if both operands are true
- `||` (OR) returns true if either operand is true
- `!` (NOT) inverts the boolean value

These operators let you express sophisticated multi-condition rules concisely.

### **Using Booleans in Conditionals**

Booleans drive control flow - they determine which code executes:

```c
# Direct use (preferred)
if ( is_suspicious )
    print "Suspicious activity";

# Explicit comparison (unnecessary but valid)
if ( is_suspicious == T )
    print "Suspicious activity";

# Negation
if ( !is_encrypted )
    print "Unencrypted traffic";

# Complex boolean logic
if ( is_internal(src) && !is_internal(dst) && port == 443/tcp )
    print "Outbound HTTPS from internal to external";
```

**Style note:** In Zeek (and most languages), directly testing a boolean is preferred over explicitly comparing to `T`. Write `if ( is_suspicious )` rather than `if ( is_suspicious == T )` - it's cleaner and more idiomatic. The boolean value itself is the condition.

**Negation** with `!` is common for checking the opposite: "if not encrypted," "if not local," "if not already processed." This reads naturally and avoids the awkwardness of negative variable names.

**Complex conditions** combine multiple checks with `&&` (all must be true) and `||` (at least one must be true). The example above checks three conditions simultaneously: the source is internal, the destination is external, and the port is 443/tcp. Only when all three are true does the detection fire. This precision prevents false positives.

### **Operator Precedence and Clarity**

When combining boolean operators, understand their precedence:

- `!` (NOT) has highest precedence
- `&&` (AND) has medium precedence
- `||` (OR) has lowest precedence

This means `!a && b || c` is evaluated as `((!a) && b) || c`. While you can rely on precedence, explicit parentheses often make complex conditions clearer: `((!a) && b) || c` is unambiguous and easier to understand at a glance.

For security logic, **clarity is more important than brevity**. When you're building a detection that will run in production, potentially generating alerts that trigger incident response, you want your boolean logic to be crystal clear. Use parentheses liberally to make your intent obvious.

### **Real-World Example: Multi-Flag Threat Assessment**

Security detection often involves tracking multiple indicators and assessing overall threat level based on how many indicators are present. Booleans are perfect for this:

```c
# Multiple detection flags
type DetectionFlags: record {
    is_scanning: bool &default=F;
    is_brute_forcing: bool &default=F;
    is_beaconing: bool &default=F;
    high_volume: bool &default=F;
};

global threat_flags: table[addr] of DetectionFlags;

function assess_threat(ip: addr): string
{
    if ( ip !in threat_flags )
        return "clean";
    
    local flags = threat_flags[ip];
    
    # Count how many flags are set
    local flag_count = 0;
    if ( flags$is_scanning ) ++flag_count;
    if ( flags$is_brute_forcing ) ++flag_count;
    if ( flags$is_beaconing ) ++flag_count;
    if ( flags$high_volume ) ++flag_count;
    
    if ( flag_count == 0 )
        return "clean";
    else if ( flag_count == 1 )
        return "low";
    else if ( flag_count == 2 )
        return "medium";
    else
        return "high";
}
```

**Understanding this pattern:** We're tracking multiple behavioral indicators for each IP address using boolean flags. Each flag represents a different suspicious behavior detected by various parts of our analysis: scanning (connecting to many ports), brute-forcing (many failed authentications), beaconing (regular connection patterns), and high volume (transferring unusual amounts of data).

**Why this approach works:** Individual indicators can produce false positives. Maybe a host scans because it's a legitimate security scanner. Maybe high volume is normal for a file server. But when multiple indicators align - a host is both scanning AND beaconing AND generating high volume - the probability of malicious activity increases dramatically.

The `assess_threat()` function counts how many flags are set and returns a threat level. Zero flags means clean, one flag is low confidence (might be legitimate), two flags is medium confidence (investigate), and three or more flags is high confidence (likely malicious). This graduated response lets you prioritize alerts and resources appropriately.

NOTE: This example is admittedly somewhat contrived since we don't usually expect there to be a relationship between a host being scanned, and hosting a beacon per se. I'm just trying to illustrate the general principle here that using multiple proxies for malicious behaviours help us develop robust detections that minimize false positives.


### **Boolean Patterns in Security Detection**

Several common patterns use booleans effectively:

**Allow/deny list checking:**

```c
local is_allowed: bool = (ip in allow_list);
if ( !is_allowed )
    # Block or alert
```

**Multi-condition validation:**

```c
if ( is_internal(src) && !is_internal(dst) && is_sensitive_port(port) )
    # Data exfiltration concern
```

**State tracking:**

```c
global alerted: set[addr] = set();

if ( is_malicious && src !in alerted )
{
    alert(src);
    add alerted[src];  
    # Don't alert again
}
```

**Threshold crossing:**

```c
local exceeded: bool = (count > threshold);
if ( exceeded && !previously_exceeded )
    # First time crossing threshold
```

**Feature flags:**

```c
const enable_experimental_detection: bool = F;

if ( enable_experimental_detection )
    # Run new detection logic
```

### **Short-Circuit Evaluation**

Understanding how boolean operators evaluate is important for both correctness and performance:

**AND (`&&`) short-circuits:** If the left operand is false, the right operand isn't evaluated because the result is already false:

```c
if ( ip in local_networks && check_expensive_condition(ip) )
    # check_expensive_condition() only called if ip is local
```

Put cheaper or more likely-to-fail conditions first to avoid unnecessary computation.

**OR (`||`) short-circuits:** If the left operand is true, the right operand isn't evaluated because the result is already true:

```c
if ( ip in deny_list || is_known_malicious(ip) )
    # is_known_malicious() not called if already in deny_list
```

This isn't just an optimization - it can affect correctness. If the right-side function has side effects or could error on certain inputs, short-circuit evaluation protects you.

### **Why This Matters for Security**

The `bool` type might seem simple, but it's the foundation of all decision logic in security detection. Every sophisticated detection ultimately reduces to a series of boolean questions: "Does this match criterion A?" "Does it also match criterion B?" "Should we alert?"

Effective use of booleans makes your detections more readable, maintainable, and correct. When you return to code months later (or when a colleague reads it for the first time), clear boolean logic with well-named variables tells a story: `if ( is_suspicious && exceeds_threshold && !is_whitelisted )` immediately conveys the detection's logic without needing to decipher complex nested conditions.

Boolean flags let you implement stateful detection - tracking what you've seen from each host, correlating indicators across events, and building comprehensive threat profiles. This state tracking is what elevates simple signature matching to sophisticated behavioral analysis.

As you build more complex detections, you'll find yourself using booleans constantly: as function return values ("does this condition hold?"), as record fields (tracking multiple attributes), in tables (mapping entities to their states), and as the glue that combines simple checks into powerful compound detections. 



## Knowledge Check: bool Type

**Q1: In Zeek, what are the literal values for true and false, and how do they differ from most other programming languages?**

A: In Zeek, true is written as `T` and false as `F` (capital letters). This differs from most languages that use lowercase `true`and `false`. This is a Zeek-specific convention you must remember.

**Q2: What's the difference between writing "if ( is_suspicious )" versus "if ( is_suspicious == T )"? Which is preferred and why?**

A: Both work and produce the same result, but `if ( is_suspicious )` is preferred and more idiomatic. The boolean value itself is the condition - you don't need to explicitly compare it to T. This style is cleaner, more readable, and is the convention in Zeek (and most languages). The boolean variable directly expresses the condition you're testing.

**Q3: Explain what "short-circuit evaluation" means for the && and || operators, and provide an example where this behavior matters.**

A: Short-circuit evaluation means: for `&&`, if the left operand is false, the right isn't evaluated (result is already false); for `||`, if the left operand is true, the right isn't evaluated (result is already true). Example: `if ( ip in local_networks && expensive_check(ip) )` - the expensive_check() only runs if ip is local. This matters for performance (avoiding unnecessary computation) and correctness (if the right-side function could error on certain inputs, short-circuiting protects you).

**Q4: Why is combining multiple boolean indicators (flags) more effective than single-indicator detection? Use a threat assessment example to explain.**

A: Individual indicators often produce false positives - a host might scan legitimately (security scanner), or generate high volume normally (file server). But when multiple indicators align (scanning AND beaconing AND high volume), the probability of malicious activity increases dramatically. The threat assessment pattern counts how many flags are set and returns graduated threat levels: zero flags = clean, one = low confidence, two = medium, three+ = high. This reduces false positives and helps prioritize responses appropriately.

**Q5: List three common security detection patterns that rely heavily on boolean values.**

A: **(1) Allow/deny list checking** - testing if an entity is in an allowed set (bool result determines whether to block/alert), 

**(2) State tracking** - using booleans to remember what you've seen (e.g., "have we already alerted on this host?" stored as bool flag to prevent duplicates), 

**(3) Multi-condition validation** - combining several boolean checks with && and || to create precise detection rules (e.g., "is_internal(src) && !is_internal(dst) && is_sensitive_port(port)" to detect potential data exfiltration).






---
[|TOC|]({{< ref "../../../moc.md" >}})
[|PREV|]({{< ref "./string.md" >}})
[|NEXT|]({{< ref "./conclusion.md" >}})

