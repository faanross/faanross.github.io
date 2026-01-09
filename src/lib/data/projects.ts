export interface Project {
	name: string;
	description: string;
	url: string;
	language: string;
	topics: string[];
}

export interface ProjectCategory {
	name: string;
	description: string;
	projects: Project[];
}

export const projectCategories: ProjectCategory[] = [
	{
		name: "Offensive Tools",
		description: "C2 frameworks and covert channel implementations",
		projects: [
			{
				name: "goMESA",
				description: "Covert C2 channel using NTP. Server functions as legitimate time server while managing agents.",
				url: "https://github.com/faanross/goMESA",
				language: "Go",
				topics: ["c2", "ntp", "covert-channel"]
			},
			{
				name: "ICMP_GOSH",
				description: "Covert C2 channel using ICMP Type 3. Tunnels commands through error messages.",
				url: "https://github.com/faanross/ICMP_GOSH",
				language: "Go",
				topics: ["c2", "icmp", "covert-channel"]
			},
			{
				name: "spinnekop",
				description: "DNS+HTTPS hybrid C2 inspired by Sunburst. DNS for check-ins, HTTPS for data transfer.",
				url: "https://github.com/faanross/spinnekop",
				language: "Go",
				topics: ["c2", "dns", "https", "sunburst"]
			},
			{
				name: "joker_screenmate",
				description: "DNS tunnel C2 using TXT records. Z-value command dispatch for covert signaling.",
				url: "https://github.com/faanross/joker_screenmate",
				language: "Go",
				topics: ["c2", "dns", "txt-records", "covert-channel"]
			},
			{
				name: "Go_TelegramAPI_C2",
				description: "C2 framework using Telegram Bot API. Commands via chat, blends with legitimate traffic.",
				url: "https://github.com/faanross/Go_TelegramAPI_C2",
				language: "Go",
				topics: ["c2", "telegram", "bot-api"]
			},
			{
				name: "IPv6_rotationalC2",
				description: "C2 using IPv6 address aliasing. Rotational staggering across multiple addresses.",
				url: "https://github.com/faanross/IPv6_rotationalC2",
				language: "Go",
				topics: ["c2", "ipv6", "address-aliasing"]
			}
		]
	},
	{
		name: "Detection Tools",
		description: "Threat hunting and packet analysis utilities",
		projects: [
			{
				name: "dns-threat-toolkit",
				description: "Consolidated DNS analysis suite. Packet inspection, threat hunting, Z-flag detection, visualization.",
				url: "https://github.com/faanross/dns-threat-toolkit",
				language: "Go",
				topics: ["dns", "threat-hunting", "packet-analysis"]
			},
			{
				name: "go-packet-peeker",
				description: "Analyze packet size variations to detect covert payloads. Histogram visualization + payload extraction.",
				url: "https://github.com/faanross/go-packet-peeker",
				language: "Go",
				topics: ["pcap", "covert-channel", "detection"]
			}
		]
	}
];
