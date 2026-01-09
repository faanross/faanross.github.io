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
				description: "Covert command and control channel using Network Time Protocol (NTP). Demonstrates how C2 traffic can be hidden within legitimate NTP communications.",
				url: "https://github.com/faanross/goMESA",
				language: "Go",
				topics: ["c2", "ntp", "covert-channel", "offensive-security", "threat-hunting"]
			},
			{
				name: "ICMP-GOSH",
				description: "Covert command and control channel using ICMP. Demonstrates how C2 traffic can be tunneled through ping requests and replies.",
				url: "https://github.com/faanross/ICMP-GOSH",
				language: "Go",
				topics: ["c2", "icmp", "covert-channel", "offensive-security", "threat-hunting"]
			}
		]
	}
];
