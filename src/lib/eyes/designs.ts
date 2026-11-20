/**
 * The eye catalogue. Every design has a stable id (A01, C07, ...) so any
 * design can be referenced from the homepage cast (HOME_CAST in field.ts).
 */

import {
	PALETTES,
	makeDrift,
	makeEye,
	makeState,
	mirrorArt,
	rand,
	type ArtDrawing,
	type Drawing,
	type Drift,
	type Group,
	type LineDrawing,
	type Palette,
	type PaletteName,
	type ProcDrawing,
	type ProcSpec
} from './engine';

export type GroupId = 'A' | 'B' | 'C' | 'D' | 'E';

export interface Design {
	id: string;
	name: string;
	group: GroupId;
	note?: string;
	draw: Drawing;
	pal: PaletteName;
	alpha?: number;
	lazy?: number;
	pair?: { gap: number; mirror?: boolean };
}

export const GROUPS: Record<GroupId, { title: string; blurb: string }> = {
	A: { title: 'signature', blurb: 'Large procedural eyes. The AIONSEC eye and its relatives.' },
	B: {
		title: 'procedural, medium',
		blurb: 'Reworked: one-glyph lid line, no crease or dotted white, calmer iris. B06 unchanged.'
	},
	C: { title: 'drawn, medium', blurb: 'Hand-drawn templates. Every character placed, so they read at mid size.' },
	D: { title: 'drawn, small', blurb: 'Two- and three-line eyes, mostly in pairs.' },
	E: { title: 'single line', blurb: 'The ultra simple ones.' }
};

/** Multi-line template literal -> rows (drops the first/last blank line). */
function t(s: string): string[] {
	const lines = s.split('\n').map((l) => l.replace(/^\t+/, ''));
	if (lines[0].trim() === '') lines.shift();
	if (lines.length && lines[lines.length - 1].trim() === '') lines.pop();
	return lines;
}

const SIGIL: ProcSpec = {
	half: 150,
	fs: 10,
	topK: 0.383,
	botK: 0.302,
	arcPow: 0.76,
	tilt: -0.025,
	ring: 2,
	ringK: 1,
	irisK: 0.325,
	pupilK: 0.365,
	pupil: 'round',
	iris: 'fibers',
	lashes: 0,
	hood: 0,
	crease: true,
	catchlight: true,
	sclera: 1
};

const OWL: Partial<ProcSpec> = {
	topK: 0.5,
	botK: 0.46,
	arcPow: 0.5,
	irisK: 0.42,
	pupilK: 0.42,
	iris: 'spokes',
	crease: false
};

const proc = (over: Partial<ProcSpec>): ProcDrawing => ({ kind: 'proc', spec: { ...SIGIL, ...over } });

/** Medium procedural eye on the simple renderer. */
const mid = (over: Partial<ProcSpec>): ProcDrawing =>
	proc({
		half: 100,
		fs: 11,
		ring: 0,
		iris: 'flat',
		crease: false,
		sclera: 0,
		simple: true,
		...over
	});

const art = (a: Omit<ArtDrawing, 'kind' | 'fs'> & { fs?: number }): ArtDrawing => ({
	kind: 'art',
	fs: 12,
	...a
});

const line = (a: Omit<LineDrawing, 'kind' | 'fs'> & { fs?: number }): LineDrawing => ({
	kind: 'line',
	fs: 13,
	...a
});

const IRIS3 = [' .-. ', '( o )', " '-' "];
const IRIS3W = [' .-. ', '( @ )', " '-' "];

export const CATALOG: Design[] = [
	/* ---------------- A · signature ---------------- */
	{ id: 'A01', name: 'sigil', group: 'A', note: 'the main one', pal: 'aion', draw: proc({}) },
	{
		id: 'A02',
		name: 'sigil mini',
		group: 'A',
		note: 'the smaller ringed one',
		pal: 'aion',
		draw: proc({ ...OWL, half: 95, ring: 1, ringK: 0.82, tilt: 0.02 })
	},
	{ id: 'A03', name: 'sigil small', group: 'A', note: 'main eye, scaled down', pal: 'aion', draw: proc({ half: 78, fs: 8 }) },
	{ id: 'A04', name: 'ghost sigil', group: 'A', pal: 'ghost', draw: proc({ half: 120 }) },
	{
		id: 'A05',
		name: 'serpent sigil',
		group: 'A',
		pal: 'ember',
		draw: proc({ half: 120, ring: 1, pupil: 'slit', iris: 'rings', catchlight: false })
	},
	{
		id: 'A06',
		name: 'orbit',
		group: 'A',
		pal: 'violet',
		draw: proc({ half: 120, pupil: 'bar', iris: 'spokes', catchlight: false })
	},
	{
		id: 'A07',
		name: 'owl ring',
		group: 'A',
		pal: 'violet',
		draw: proc({ ...OWL, half: 95, ring: 1, ringK: 1, iris: 'rings' })
	},

	/* ---------------- B · procedural, medium ---------------- */
	// All but B06 use the simple renderer: one-glyph lid line, no crease or
	// sclera, calmer iris. B06 is kept exactly as first shown.
	{ id: 'B01', name: 'almond', group: 'B', pal: 'aion', draw: mid({ tilt: 0 }) },
	{
		id: 'B02',
		name: 'owl',
		group: 'B',
		pal: 'aion',
		draw: mid({ topK: 0.48, botK: 0.44, arcPow: 0.5, irisK: 0.4, pupilK: 0.4, iris: 'rings', catchlight: false, half: 90 })
	},
	{
		id: 'B03',
		name: 'serpent',
		group: 'B',
		pal: 'ember',
		draw: mid({ topK: 0.34, botK: 0.3, arcPow: 1.1, irisK: 0.36, pupilK: 0.32, pupil: 'slit', catchlight: false })
	},
	{ id: 'B04', name: 'hooded', group: 'B', pal: 'aion', lazy: 1.7, draw: mid({ hood: 0.4, catchlight: false }) },
	{
		id: 'B05',
		name: 'goat',
		group: 'B',
		pal: 'violet',
		draw: mid({ topK: 0.36, botK: 0.33, arcPow: 0.9, irisK: 0.38, pupil: 'bar', iris: 'rings', catchlight: false })
	},
	{
		id: 'B06',
		name: 'bands',
		group: 'B',
		pal: 'aion',
		note: 'kept as first shown',
		draw: proc({ half: 100, ring: 0, iris: 'rings', sclera: 0, crease: false })
	},
	{ id: 'B07', name: 'lashes', group: 'B', pal: 'aion', draw: mid({ lashes: 5 }) },
	{ id: 'B08', name: 'bands, clean', group: 'B', pal: 'aion', note: 'B06 with the thin lid line', draw: mid({ iris: 'rings', catchlight: false }) },
	{
		id: 'B09',
		name: 'narrow',
		group: 'B',
		pal: 'ember',
		draw: mid({ topK: 0.27, botK: 0.22, arcPow: 0.9, irisK: 0.3, pupilK: 0.38, iris: 'rings', catchlight: false, half: 105 })
	},
	{
		id: 'B10',
		name: 'round',
		group: 'B',
		pal: 'violet',
		draw: mid({ topK: 0.46, botK: 0.42, arcPow: 0.45, irisK: 0.36, iris: 'spokes', half: 85 })
	},

	/* ---------------- C · drawn, medium ---------------- */
	{
		id: 'C01',
		name: 'pill',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
 .---------.
(~~~~~~~~~~~)
 '---------'
`),
			iris: ['(o)'],
			wide: ['(O)'],
			range: [4, 0]
		})
	},
	{
		id: 'C02',
		name: 'lens',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
   .-'''''-.
  /~~~~~~~~~\
 |~~~~~~~~~~~|
  \~~~~~~~~~/
   '-.....-'
`),
			iris: IRIS3,
			wide: IRIS3W,
			range: [3, 1]
		})
	},
	{
		id: 'C03',
		name: 'almond',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
      _.-----._
   .-'~~~~~~~~~'-.
  <~~~~~~~~~~~~~~~>
   '-._~~~~~~~_.-'
       ''---''
`),
			iris: IRIS3W,
			wide: [' .-. ', '(@@@)', " '-' "],
			range: [4, 1]
		})
	},
	{
		id: 'C04',
		name: 'orb',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
     .-'''''''-.
   .'~~~~~~~~~~~'.
  /~~~~~~~~~~~~~~~\
 |~~~~~~~~~~~~~~~~~|
  \~~~~~~~~~~~~~~~/
   '.~~~~~~~~~~~.'
     '-.......-'
`),
			iris: t(String.raw`
 .---.
/     \
| (@) |
\     /
 '---'
`),
			wide: t(String.raw`
 .---.
/     \
| @@@ |
\     /
 '---'
`),
			range: [4, 1],
			fill: '.'
		})
	},
	{
		id: 'C05',
		name: 'sleepy',
		group: 'C',
		pal: 'aion',
		lazy: 2.4,
		draw: art({
			rows: t(String.raw`
   _____________
  /=============\
 (~~~~~~~~~~~~~~~)
  '-------------'
`),
			iris: ['(o)'],
			wide: ['(O)'],
			range: [5, 0]
		})
	},
	{
		id: 'C06',
		name: 'lashes',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
     \ \ | / /
    .-'''''''-.
   (~~~~~~~~~~~)
    '-._____.-'
`),
			closed: ['', '', "    '-._____.-'", '     / / | \\ \\'],
			iris: ['(@)'],
			wide: ['(O)'],
			range: [4, 0]
		})
	},
	{
		id: 'C07',
		name: 'cat',
		group: 'C',
		pal: 'ember',
		draw: art({
			rows: t(String.raw`
     .--'''--.
   .'~~~~~~~~~'.
  <~~~~~~~~~~~~~>
   '.~~~~~~~~~.'
     '--...--'
`),
			iris: [' .|. ', '( | )', " '|' "],
			wide: [' .-. ', '( 0 )', " '-' "],
			pupils: '|0',
			range: [4, 1]
		})
	},
	{
		id: 'C08',
		name: 'goat',
		group: 'C',
		pal: 'violet',
		draw: art({
			rows: t(String.raw`
    .---------.
  .'~~~~~~~~~~~'.
 (~~~~~~~~~~~~~~~)
  '.~~~~~~~~~~~.'
    '---------'
`),
			iris: [' .---. ', '(=====)', " '---' "],
			pupils: '=',
			range: [3, 1]
		})
	},
	{
		id: 'C09',
		name: 'providence',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
          .
         / \
        /   \
       /     \
      /.-'''-.\
     /(~~~~~~~)\
    /  '-...-'  \
   /_____________\
`),
			iris: ['(@)'],
			wide: ['(O)'],
			range: [2, 0]
		})
	},
	{
		id: 'C10',
		name: 'seal',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
       .-'''-.
    .-'   |   '-.
   /    .---.    \
  |   .'~~~~~'.   |
--+--(~~~~~~~~~)--+--
  |   '.~~~~~.'   |
   \    '---'    /
    '-.   |   .-'
       '-...-'
`),
			iris: ['(@)'],
			wide: ['(O)'],
			range: [3, 1]
		})
	},
	{
		id: 'C11',
		name: 'sensor',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
.-----------------------.
|~~~~~~~~~~~~~~~~~~~~~~~|
'-----------------------'
`),
			iris: ['-=O=-'],
			wide: ['=<@>='],
			pupils: 'O@',
			range: [9, 0]
		})
	},
	{
		id: 'C12',
		name: 'camera',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
    .-------.
  .'  .---.  '.
 /  .'~~~~~'.  \
|  (~~~~~~~~~)  |
 \  '.~~~~~.'  /
  '.  '---'  .'
    '-------'
`),
			iris: ['(O)'],
			wide: ['(@)'],
			range: [3, 1]
		})
	},
	{
		id: 'C13',
		name: 'screen',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
+-----------------+
|~~~~~~~~~~~~~~~~~|
|~~~~~~~~~~~~~~~~~|
|~~~~~~~~~~~~~~~~~|
+-----------------+
`),
			iris: ['.-.', '|o|', "'-'"],
			wide: ['.-.', '|@|', "'-'"],
			pupils: 'o@',
			range: [6, 1],
			fill: '.'
		})
	},
	{
		id: 'C14',
		name: 'double lid',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
       _____
   .-''_____''-.
  /.-'~~~~~~~'-.\
 (~~~~~~~~~~~~~~~)
  \'-.~~~~~~~.-'/
   '-._______.-'
`),
			iris: IRIS3W,
			wide: [' .-. ', '(@@@)', " '-' "],
			range: [4, 1]
		})
	},
	{
		id: 'C15',
		name: 'squint',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
 _______________
<~~~~~~~~~~~~~~~>
`),
			iris: ['(o)'],
			wide: ['(@)'],
			range: [6, 0]
		})
	},
	{
		id: 'C16',
		name: 'weeping',
		group: 'C',
		pal: 'ghost',
		draw: art({
			rows: t(String.raw`
   .-'''-.
  (~~~~~~~)
   '-._.-'
        :
        '
`),
			iris: ['(o)'],
			wide: ['(O)'],
			range: [2, 0]
		})
	},
	{
		id: 'C17',
		name: 'diamond',
		group: 'C',
		pal: 'violet',
		draw: art({
			rows: t(String.raw`
      .
    .'~'.
  .'~~~~~'.
 <~~~~~~~~~>
  '.~~~~~.'
    '.~.'
      '
`),
			iris: [' . ', '(@)', " ' "],
			pupils: '@',
			range: [2, 1]
		})
	},
	{
		id: 'C18',
		name: 'brooding',
		group: 'C',
		pal: 'aion',
		lazy: 1.6,
		draw: art({
			rows: t(String.raw`
   _________
  /=========\
 |~~~~~~~~~~~|
  \~~~~~~~~~/
   '-.....-'
`),
			iris: ['( o )'],
			wide: ['( @ )'],
			range: [3, 1]
		})
	},
	{
		id: 'C19',
		name: 'button',
		group: 'C',
		pal: 'aion',
		draw: art({
			rows: t(String.raw`
  .---.
 /~~~~~\
|~~~~~~~|
 \~~~~~/
  '---'
`),
			iris: ['(o)'],
			wide: ['(@)'],
			range: [2, 1]
		})
	},

	/* ---------------- D · drawn, small ---------------- */
	{
		id: 'D01',
		name: 'brows',
		group: 'D',
		pal: 'aion',
		pair: { gap: 1 },
		draw: art({ fs: 13, rows: ['  _  ', '(~~~)'], iris: ['o'], wide: ['O'], range: [1, 0] })
	},
	{
		id: 'D02',
		name: 'angry',
		group: 'D',
		pal: 'aion',
		pair: { gap: 1, mirror: true },
		draw: art({ fs: 13, rows: ['   \\ ', '(~~~)'], iris: ['o'], wide: ['O'], range: [1, 0] })
	},
	{
		id: 'D03',
		name: 'worried',
		group: 'D',
		pal: 'aion',
		pair: { gap: 1, mirror: true },
		draw: art({ fs: 13, rows: ['   / ', '(~~~)'], iris: ['o'], wide: ['O'], range: [1, 0] })
	},
	{
		id: 'D04',
		name: 'tiny lens',
		group: 'D',
		pal: 'aion',
		pair: { gap: 1 },
		draw: art({ fs: 13, rows: [' .-. ', '(~~~)', " '-' "], iris: ['o'], wide: ['O'], range: [1, 0] })
	},
	{
		id: 'D05',
		name: 'tiny almond',
		group: 'D',
		pal: 'aion',
		pair: { gap: 2 },
		draw: art({ fs: 13, rows: [' .-. ', '<~~~>', " '-' "], iris: ['@'], wide: ['O'], range: [1, 0] })
	},
	{
		id: 'D06',
		name: 'spectacles',
		group: 'D',
		pal: 'aion',
		draw: art({ fs: 13, rows: ['(~~~)-(~~~)'], iris: ['o'], wide: ['O'], range: [1, 0] })
	},
	{
		id: 'D07',
		name: 'goggles',
		group: 'D',
		pal: 'aion',
		draw: art({ fs: 13, rows: [' ___   ___ ', '[~~~]=[~~~]'], iris: ['@'], wide: ['O'], range: [1, 0] })
	},
	{
		id: 'D08',
		name: 'heavy lids',
		group: 'D',
		pal: 'aion',
		lazy: 2.5,
		pair: { gap: 1 },
		draw: art({ fs: 13, rows: ['_____', '(~~~)'], iris: ['o'], wide: ['O'], range: [1, 0] })
	},

	/* ---------------- E · single line ---------------- */
	{ id: 'E01', name: '(o)', group: 'E', pal: 'aion', pair: { gap: 1 }, draw: line({ open: '(', close: ')', inner: 1, pupil: 'o', wide: 'O' }) },
	{ id: 'E02', name: '( o )', group: 'E', pal: 'aion', pair: { gap: 2 }, draw: line({ open: '(', close: ')', inner: 3, pupil: 'o', wide: '@' }) },
	{ id: 'E03', name: '<o>', group: 'E', pal: 'aion', pair: { gap: 2 }, draw: line({ open: '<', close: '>', inner: 1, pupil: 'o', wide: 'O' }) },
	{ id: 'E04', name: '[@]', group: 'E', pal: 'aion', pair: { gap: 2 }, draw: line({ open: '[', close: ']', inner: 1, pupil: '@', wide: '#' }) },
	{ id: 'E05', name: '{o}', group: 'E', pal: 'aion', pair: { gap: 1 }, draw: line({ open: '{', close: '}', inner: 2, pupil: 'o', wide: 'O' }) },
	{ id: 'E06', name: '((o))', group: 'E', pal: 'aion', pair: { gap: 2 }, draw: line({ open: '((', close: '))', inner: 1, pupil: 'o', wide: 'O' }) },
	{ id: 'E07', name: '<(o)>', group: 'E', pal: 'aion', pair: { gap: 2 }, draw: line({ open: '<(', close: ')>', inner: 1, pupil: 'o', wide: 'O' }) },
	{ id: 'E08', name: '-(o)-', group: 'E', pal: 'aion', pair: { gap: 1 }, draw: line({ open: '-(', close: ')-', inner: 1, pupil: 'o', wide: 'O' }) },
	{ id: 'E09', name: '|o|', group: 'E', pal: 'aion', pair: { gap: 2 }, draw: line({ open: '|', close: '|', inner: 1, pupil: 'o', wide: 'O' }) },
	{ id: 'E10', name: '=( o )=', group: 'E', pal: 'aion', pair: { gap: 1 }, draw: line({ open: '=(', close: ')=', inner: 3, pupil: 'o', wide: '@' }) },
	{ id: 'E11', name: '[ o ]', group: 'E', pal: 'aion', pair: { gap: 2 }, draw: line({ open: '[', close: ']', inner: 3, pupil: 'o', wide: '@' }) },
	{ id: 'E12', name: '(   o   )', group: 'E', pal: 'aion', draw: line({ open: '(', close: ')', inner: 7, pupil: 'o', wide: '@' }) },
	{ id: 'E13', name: '(-o-)', group: 'E', pal: 'aion', pair: { gap: 1 }, draw: line({ open: '(', close: ')', inner: 3, pupil: 'o', wide: 'O', fill: '-' }) },
	{ id: 'E14', name: '<==o==>', group: 'E', pal: 'aion', draw: line({ open: '<', close: '>', inner: 5, pupil: 'o', wide: 'O', fill: '=' }) },
	{ id: 'E15', name: '(0)', group: 'E', pal: 'aion', pair: { gap: 1 }, draw: line({ open: '(', close: ')', inner: 1, pupil: '0', wide: '@' }) },
	{ id: 'E16', name: '~(o)~', group: 'E', pal: 'aion', pair: { gap: 1 }, draw: line({ open: '~(', close: ')~', inner: 1, pupil: 'o', wide: 'O' }) },
	{ id: 'E17', name: 'o  o', group: 'E', pal: 'aion', alpha: 0.8, pair: { gap: 2 }, draw: line({ open: '', close: '', inner: 1, pupil: 'o', wide: 'O' }) },
	{ id: 'E18', name: '*   *', group: 'E', pal: 'aion', alpha: 0.8, pair: { gap: 3 }, draw: line({ open: '', close: '', inner: 1, pupil: '*', wide: '#' }) },
	{ id: 'E19', name: '@  @', group: 'E', pal: 'aion', alpha: 0.8, pair: { gap: 2 }, draw: line({ open: '', close: '', inner: 1, pupil: '@', wide: '#' }) },
	{ id: 'E20', name: '(o)(o)', group: 'E', pal: 'aion', pair: { gap: 0 }, draw: line({ open: '(', close: ')', inner: 1, pupil: 'o', wide: 'O' }) }
];

const BY_ID = new Map(CATALOG.map((d) => [d.id, d]));
export const byId = (id: string) => BY_ID.get(id);

export interface SpawnOptions {
	pal?: Palette;
	alpha?: number;
	/** Override the drawing, e.g. a resized procedural spec. */
	drawing?: Drawing;
	/** Floating motion; defaults to a gentle random drift, null to hold still. */
	drift?: Drift | null;
}

/** Create the eye(s) for a design, sharing one gaze. Position with placeGroup(). */
export function spawn(d: Design, time: number, opts: SpawnOptions = {}): Group {
	const draw = opts.drawing ?? d.draw;
	const pal = opts.pal ?? PALETTES[d.pal];
	const alpha = opts.alpha ?? d.alpha ?? 0.92;
	const st = makeState(time, d.lazy ?? rand(0.85, 1.2));
	const first = makeEye(draw, st, pal, alpha);

	const base = {
		st,
		hidden: false,
		x: 0,
		y: 0,
		drift: opts.drift === undefined ? makeDrift() : opts.drift,
		fadeStart: -Infinity
	};

	if (!d.pair || draw.kind === 'proc') {
		st.size = draw.kind === 'proc' ? draw.spec.half : Math.max(40, first.rx * 1.6);
		return { ...base, eyes: [first], offsets: [0], rx: first.rx, ry: first.ry };
	}

	const second = makeEye(d.pair.mirror && draw.kind === 'art' ? mirrorArt(draw) : draw, st, pal, alpha);
	const gap = d.pair.gap * draw.fs * 0.6;
	const off = first.rx + gap / 2;
	st.size = Math.max(40, (off + first.rx) * 1.2);
	return { ...base, eyes: [first, second], offsets: [-off, off], rx: off + first.rx, ry: first.ry };
}
