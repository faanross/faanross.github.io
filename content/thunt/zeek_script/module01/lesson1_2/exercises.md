---
showTableOfContents: true
title: "Part 4 - Practical Exercises"
type: "page"
---

## **PART 4: PRACTICAL EXERCISES**

Now let's apply this architectural knowledge to practical scenarios. These exercises will help you think through deployment decisions and performance considerations.

### **Exercise 1: Design Your Network Architecture**

For this exercise, you're going to design a Zeek deployment for a hypothetical organization. This will help you think through the architectural decisions we've discussed.

**Scenario:**

You're deploying Zeek for a mid-sized company with the following characteristics:

- **Network:** 2 Gbps internet connection, 10 Gbps internal backbone
- **Endpoints:** 500 workstations, 50 servers
- **Traffic pattern:** Heavy web browsing, moderate email, some file sharing
- **Monitoring goals:** Detect C2 traffic, ransomware propagation, data exfiltration
- **Budget:** Moderate (can afford multiple servers but not enterprise appliances)

**Your task:**

Design a Zeek deployment by answering these questions and creating a diagram:


1. **Sensor placement:** Where will you deploy Zeek sensors?

    - At the internet gateway? (captures all external traffic)
    - On the internal backbone? (captures server-to-server traffic)
    - Both? (comprehensive but more complex)

   Think about what visibility you need for your detection goals. C2 detection requires seeing outbound traffic. Ransomware propagation detection requires seeing internal traffic.

2. **Cluster vs. single instance:** Will you deploy a cluster or a single Zeek instance?

    - Calculate: 2 Gbps internet + potentially 10 Gbps internal
    - Consider: A single instance can handle ~1-2 Gbps, so you likely need a cluster
3. **Cluster design:** If using a cluster, how many nodes of each type?

    - Workers: How many do you need for 2-3 Gbps total?
    - Manager: How many? (hint: usually just one)
    - Proxies: Do you need them for this scale?
    - Logger: How many? Will it be dedicated or combined with manager?
4. **Hardware specifications:** What specs for each node?

    - Consider CPU, memory, storage based on the formulas we discussed
    - Think about growth (plan for 50% traffic increase over 2 years)
5. **Load balancing:** How will you distribute traffic across workers?

    - Hardware load balancer? (expensive but performant)
    - AF_PACKET fanout? (free but requires multiple workers on same host)
    - Other approach?

**Deliverable:**

Create a document that includes:

- A network diagram showing where Zeek sensors are placed
- A cluster architecture diagram showing node types and connections
- A table of hardware specifications for each node
- Written justification for your design decisions

Spend at least 30 minutes on this exercise. There's no single "correct" answer-the goal is to think through the trade-offs and design a solution that meets the requirements.



### **Exercise 2: Plan Sensor Placement for Maximum Visibility**

This exercise focuses specifically on sensor placement strategy, which is often overlooked but critically important.

**Scenario:**

You have the following network topology:

```
                    Internet
                       │
                       ▼
                 ┌─────────┐
                 │Firewall │
                 └────┬────┘
                      │
              ┌───────┴───────┐
              │               │
              ▼               ▼
        ┌─────────┐     ┌─────────┐
        │  DMZ    │     │Internal │
        │ Servers │     │ Network │
        └─────────┘     └────┬────┘
                             │
                    ┌────────┼────────┐
                    │        │        │
                    ▼        ▼        ▼
               ┌────────┐ ┌────────┐ ┌────────┐
               │Worksta-│ │Servers │ │ Guest  │
               │ tions  │ │        │ │ WiFi   │
               └────────┘ └────────┘ └────────┘
```

**Your task:**

Decide where to place Zeek sensors to achieve these objectives:

1. **Detect external threats:** C2 traffic, inbound attacks, data exfiltration
2. **Detect lateral movement:** Compromised workstations attacking servers
3. **Detect ransomware:** Propagating between workstations
4. **Monitor DMZ:** Public-facing servers under attack

For each potential sensor placement, consider:

**Placement Option A: Between Internet and Firewall**

- What visibility: All traffic entering/leaving organization
- What you'll detect: External C2, data exfiltration, inbound attacks
- What you'll miss: Internal lateral movement, workstation-to-workstation attacks
- Traffic volume: All internet traffic (could be high)
- Not great since it is pre-NAT, unable to distinguish between different internal target hosts

**Placement Option B: Behind Firewall on Internal Network**

- What visibility: All internal traffic
- What you'll detect: Lateral movement, internal reconnaissance, ransomware spread
- Traffic volume: Potentially very high (all internal communications)

**Placement Option C: DMZ segment**

- What visibility: Traffic to/from DMZ servers
- What you'll detect: Attacks against public services, compromised DMZ servers
- What you'll miss: Most internal activity, direct workstation threats
- Traffic volume: Moderate

**Placement Option D: Multiple sensors (combined approach)**

- What visibility: Comprehensive
- What you'll detect: Everything
- What you'll miss: Nothing, but complexity increases
- Traffic volume: Must handle multiple traffic streams

**Questions to answer:**

1. If you could only deploy ONE sensor, where would you place it and why?
2. What are the minimum number of sensors needed to meet all four objectives?
3. For each sensor you propose, estimate the traffic volume it will see
4. How would your placement strategy change if:
    - Your primary concern was ransomware detection?
    - Your primary concern was APT/espionage detection?
    - You had unlimited budget vs. tight budget?

**Deliverable:**

Write a 1-2 page document explaining:

- Your sensor placement strategy with justifications
- A diagram showing sensor locations
- Analysis of what each sensor will and won't see
- Traffic volume estimates for each sensor
- How your strategy addresses each of the four objectives

### **Exercise 3: Calculate Resource Requirements**

This exercise helps you practice sizing Zeek deployments based on traffic characteristics.

**Scenario 1: Small Office**

- **Internet bandwidth:** 100 Mbps
- **Employees:** 50
- **Peak concurrent connections:** ~2,000
- **Traffic pattern:** 80% web browsing, 15% email, 5% other

Calculate:

1. How much RAM does Zeek need? (Use the formula from earlier)
2. How much disk space for logs per day?
3. How many CPU cores are recommended?
4. Can this run on a single instance or need a cluster?

**Scenario 2: Medium Enterprise**

- **Internet bandwidth:** 5 Gbps
- **Employees:** 1,000
- **Peak concurrent connections:** ~50,000
- **Traffic pattern:** 60% web, 20% cloud services, 10% email, 10% other

Calculate:

1. How many worker nodes needed?
2. Total RAM across all workers?
3. Logger disk space for logs (per day, per week, per month)?
4. Do you need proxy nodes?

**Scenario 3: Large Data Center**

- **Network bandwidth:** 40 Gbps aggregate
- **Connections:** ~500,000 concurrent
- **Traffic pattern:** Mixed (web services, APIs, databases, file storage)

Calculate:

1. How many workers needed?
2. Is this a candidate for PF_RING?
3. Total cluster memory requirements?
4. Logger capacity requirements (assuming 30-day retention)?

**Deliverable:**

Create a spreadsheet or table with your calculations for all three scenarios. Show your work so you understand how you arrived at each number. Include:

- Worker count
- CPU cores per worker
- RAM per worker (and total)
- Storage requirements
- Any special hardware (load balancers, PF_RING, etc.)

---



---
[|TOC|]({{< ref "../../moc.md" >}})
[|PREV|]({{< ref "./optimize.md" >}})
[|NEXT|]({{< ref "./validation.md" >}})

