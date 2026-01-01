---
showTableOfContents: true
title: "Using MITRE ATT&CK for Threat Hunting"
type: "page"
---






## Lateral Movement: Expanding the Foothold

Lateral movement represents adversaries traversing your network, moving from system to system to reach objectives or expand control. They use remote services, exploit trust relationships, and leverage stolen credentials.

**Threat Hunting Reality**: Excellent hunting opportunities using both endpoint and network data. Lateral movement generates rich telemetry from Windows Security Event logs for authentication (Event IDs 4624, 4625, 4648), Sysmon Event ID 1 for process creation on remote systems, and Sysmon Event ID 3 for network connections. Network-side detection uses firewall logs, NetFlow/IPFIX data, Zeek logs for protocol analysis, and packet captures.

Windows Management Instrumentation (WMI) is commonly abused for lateral movement. Hunt for `wmic.exe` or PowerShell executing with remote parameters in command lines - patterns like `wmic /node:[target] process call create` or `Invoke-WmiMethod -ComputerName` are clear indicators (Sysmon Event ID 1).

On destination systems, processes spawned by `wmiprvse.exe` (WMI Provider Host), especially `powershell.exe` or `cmd.exe`, indicate remote execution. Network-side, WMI uses RPC over TCP port 135 for initial connection, then dynamic high ports (49152-65535). Zeek's `dce_rpc.log` captures this activity - hunt for connections to port 135 followed by high-port traffic between workstations or workstation-to-server, especially in rapid succession.

Remote Desktop Protocol (RDP) lateral movement shows through Event ID 4624 with logon type 10 (RemoteInteractive). Network telemetry reveals RDP through TCP port 3389 connections. Hunt in firewall logs or NetFlow for unusual RDP traffic patterns - workstation-to-workstation RDP, or single sources connecting to multiple destinations sequentially. Zeek's `rdp.log` provides deeper visibility into RDP session characteristics.

SMB-based lateral movement appears in both host and network telemetry. Event IDs 5140 and 5145 show share access on endpoints, while network analysis reveals SMB traffic (TCP ports 445, 139). Zeek's `smb_files.log` and `smb_mapping.log` track file transfers and share mappings. Hunt for workstations accessing administrative shares (`C$`, `ADMIN$`) on multiple systems, or unusual SMB file write patterns across your network.

Authentication patterns reveal lateral movement campaigns. NetFlow can show the same source IP establishing connections to multiple destinations on authentication ports (88 for Kerberos, 389/636 for LDAP, 445 for SMB) in rapid succession - indicating an adversary pivoting through your network. Correlate this with Event ID 4624 logon events to confirm credential usage.

**However, the challenge is legitimate administration**. IT teams use WMI, RDP, and administrative shares daily. Effective hunting requires understanding normal administrative traffic flows, typical management system IPs, and expected network patterns to distinguish attacks from operations.





---
[|TOC|]({{< ref "../../../thrunt/_index.md" >}})
[|PREV|]({{< ref "./10_discovery.md" >}})
[|NEXT|]({{< ref "./12_collect.md" >}})

