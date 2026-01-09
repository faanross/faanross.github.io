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
		description: "Podcast discussions on threat hunting and security research",
		talks: [
			{
				videoId: "G2QYJFalj38",
				title: "AntiCast Episode 1",
				description: "Threat hunting fundamentals and methodology"
			},
			{
				videoId: "6Nl3uKlIReI",
				title: "AntiCast Episode 2",
				description: "Network-centric threat detection"
			}
		]
	},
	{
		name: "C2 Webcast Series",
		description: "Deep dives into command and control infrastructure and detection",
		talks: [
			{
				videoId: "aD8w0Q_IZJc",
				title: "C2 Webcast Part 1",
				description: "Introduction to command and control"
			},
			{
				videoId: "0xHfMzIEh-U",
				title: "C2 Webcast Part 2",
				description: "C2 communication patterns"
			},
			{
				videoId: "xN7DG6pxFZk",
				title: "C2 Webcast Part 3",
				description: "Protocol analysis and detection"
			},
			{
				videoId: "U3gIx1Ojo_U",
				title: "C2 Webcast Part 4",
				description: "Advanced C2 techniques"
			},
			{
				videoId: "wlZfypMkOGc",
				title: "C2 Webcast Part 5",
				description: "Hunting C2 in practice"
			}
		]
	}
];
