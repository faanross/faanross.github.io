export interface Project {
	name: string;
	description: string;
	url: string;
	language: string;
	topics: string[];
}

export const projects: Project[] = [
	{
		name: "goMESA",
		description: "Covert command and control channel using Network Time Protocol (NTP). Demonstrates how C2 traffic can be hidden within legitimate NTP communications.",
		url: "https://github.com/faanross/goMESA",
		language: "Go",
		topics: ["c2", "ntp", "covert-channel", "offensive-security", "threat-hunting"]
	}
];
