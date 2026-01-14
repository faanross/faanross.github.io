export interface Talk {
	videoId: string;
	title: string;
	description?: string;
}

export interface TalkSeries {
	name: string;
	description: string;
	talks: Talk[];
}

export const talkSeries: TalkSeries[] = [
	{
		name: "AntiCasts",
		description: "Discussions on C2, threat hunting, and security research",
		talks: [
			{
				videoId: "G2QYJFalj38",
				title: "AntiCast Episode 1",
				description: "Let's Build A Covert C2 Channel"
			},
			{
				videoId: "6Nl3uKlIReI",
				title: "AntiCast Episode 2",
				description: "Developing a C2 Framework"
			},
			{
				videoId: "4u5X2R-FQsU",
				title: "AntiCast Episode 3",
				description: "Threat Hunting Malware Communication over DNS"
			}
		]
	},
	{
		name: "C2 Webcast Series",
		description: "Deep dives into command and control infrastructure and detection",
		talks: [
			{
				videoId: "aD8w0Q_IZJc",
				title: "C2 Webcast Episode 1",
				description: "Fiesta"
			},
			{
				videoId: "0xHfMzIEh-U",
				title: "C2 Webcast Episode 2",
				description: "Merlin and Data Jitter"
			},
			{
				videoId: "xN7DG6pxFZk",
				title: "C2 Webcast Episode 3",
				description: "DNS Tunneling (dnscat2)"
			},
			{
				videoId: "U3gIx1Ojo_U",
				title: "C2 Webcast Episode 4",
				description: "Building Your Own Threat Hunting Home Lab"
			},
			{
				videoId: "wlZfypMkOGc",
				title: "C2 Webcast Episode 5",
				description: "Tunneled C2 Communication with Ligolo-ng"
			},
			{
				videoId: "x_X1o22yXRA",
				title: "C2 Webcast Episode 6",
				description: "Velociraptor as C2"
			}
		]
	}
];
