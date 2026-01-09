export interface Article {
	title: string;
	url: string;
}

export interface ArticleSeries {
	name: string;
	description: string;
	articles: Article[];
}

export const articleSeries: ArticleSeries[] = [
	{
		name: "Malware of the Day",
		description: "Monthly intrusion simulations with pcap + Zeek logs for hands-on threat hunting practice",
		articles: [
			{
				title: "TXT Record Abuse in DNS C2 (Joker Screenmate)",
				url: "https://www.activecountermeasures.com/malware-of-the-day-txt-record-abuse-in-dns-c2-joker-screenmate/"
			},
			{
				title: "Command and Control via Google Workspace APIs",
				url: "https://www.activecountermeasures.com/malware-of-the-day-command-and-control-via-google-workspace-apis/"
			},
			{
				title: "Agent-to-Agent Communication via SMB (AdaptixC2)",
				url: "https://www.activecountermeasures.com/malware-of-the-day-agent-to-agent-communication-via-smb-adaptixc2/"
			},
			{
				title: "Velociraptor as C2",
				url: "https://www.activecountermeasures.com/malware-of-the-day-velociraptor-as-c2"
			},
			{
				title: "ZetaSwitch - DNS/HTTP Multi-Modal C2",
				url: "https://www.activecountermeasures.com/malware-of-the-day-zetaswitch-dns-http-multi-modal-c2/"
			},
			{
				title: "Multi-Modal C2 Communication (Numinon C2)",
				url: "https://www.activecountermeasures.com/malware-of-the-day-multi-modal-c2-communication-numinon-c2/"
			},
			{
				title: "C2 over ICMP (ICMP-GOSH)",
				url: "https://www.activecountermeasures.com/malware-of-the-day-c2-over-icmp-icmp-gosh/"
			},
			{
				title: "C2 over NTP (goMESA)",
				url: "https://www.activecountermeasures.com/malware-of-the-day-c2-over-ntp-gomesa/"
			},
			{
				title: "IPv6 Address Aliasing",
				url: "https://www.activecountermeasures.com/malware-of-the-day-ipv6-address-aliasing/"
			},
			{
				title: "Merlin C2 Data Jitter",
				url: "https://www.activecountermeasures.com/malware-of-the-day-merlin-c2-data-jitter/"
			},
			{
				title: "Tunneling RDP with Microsoft Dev Tunnels",
				url: "https://www.activecountermeasures.com/malware-of-the-day-tunneling-rdp-with-microsoft-dev-tunnels/"
			},
			{
				title: "Tunneling Havoc C2 with Microsoft Dev Tunnels",
				url: "https://www.activecountermeasures.com/malware-of-the-day-tunneling-havoc-c2-with-microsoft-dev-tunnels/"
			},
			{
				title: "Specula",
				url: "https://www.activecountermeasures.com/malware-of-the-day-specula/"
			},
			{
				title: "IcedID Loader to ALPHV Ransomware",
				url: "https://www.activecountermeasures.com/malware-of-the-day-icedid-loader-to-alphv-ransomware-campaign/"
			},
			{
				title: "XenoRAT",
				url: "https://www.activecountermeasures.com/malware-of-the-day-xenorat/"
			},
			{
				title: "AsyncRAT",
				url: "https://www.activecountermeasures.com/malware-of-the-day-asyncrat/"
			},
			{
				title: "Tunneled C2 Beaconing (Ligolo-ng)",
				url: "https://www.activecountermeasures.com/malware-of-the-day-tunneled-c2-beaconing/"
			}
		]
	},
	{
		name: "Network Threat Hunting",
		description: "Deep dives into C2 detection, DNS analysis, and threat hunting methodology",
		articles: [
			{
				title: "Hunt What Hurts: The Pyramid of Pain",
				url: "https://www.activecountermeasures.com/hunt-what-hurts-the-pyramid-of-pain"
			},
			{
				title: "Threat Hunting and the Philosophy of Assumed Breach",
				url: "https://www.activecountermeasures.com/threat-hunting-and-the-philosophy-of-assumed-breach/"
			},
			{
				title: "A Network Threat Hunter's Guide to DNS Records",
				url: "https://www.activecountermeasures.com/a-network-threat-hunters-guide-to-dns-records/"
			},
			{
				title: "DNS Packet Inspection for Network Threat Hunters",
				url: "https://www.activecountermeasures.com/dns-packet-inspection-for-network-threat-hunters/"
			},
			{
				title: "Threat Hunting C2 over HTTPS Using TLS Certificates",
				url: "https://www.activecountermeasures.com/threat-hunting-c2-over-https-connections-using-the-tls-certificate/"
			},
			{
				title: "Beginner's Guide to C2 Part 1 - How C2 Frameworks Operate",
				url: "https://www.activecountermeasures.com/the-beginners-guide-to-command-and-control-part-1-how-c2-frameworks-operate/"
			},
			{
				title: "Beginner's Guide to C2 Part 2 - The Role of C2 in Modern Threats",
				url: "https://www.activecountermeasures.com/the-beginners-guide-to-command-and-control-part-2-the-role-of-c2-in-modern-threat-campaigns/"
			},
			{
				title: "Threat Hunting a Telegram C2 Channel",
				url: "https://www.activecountermeasures.com/threat-hunting-a-telegram-c2-channel/"
			},
			{
				title: "Measuring Data Jitter Using RCR",
				url: "https://www.activecountermeasures.com/measuring-data-jitter-using-rcr/"
			},
			{
				title: "A Network Threat Hunter's Guide to C2 over QUIC",
				url: "https://www.activecountermeasures.com/a-network-threat-hunters-guide-to-c2-over-quic/"
			},
			{
				title: "Understanding C2 Beacons - Part 1",
				url: "https://www.activecountermeasures.com/malware-of-the-day-understanding-c2-beacons-part-1-of-2/"
			},
			{
				title: "Understanding C2 Beacons - Part 2",
				url: "https://www.activecountermeasures.com/malware-of-the-day-understanding-c2-beacons-part-2-of-2/"
			}
		]
	}
];
