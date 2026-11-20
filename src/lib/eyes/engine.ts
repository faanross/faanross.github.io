/**
 * ASCII eye engine.
 *
 * Used by the homepage and the 404 page: gaze and blink state, glyph sprite
 * atlases, and three renderers:
 *
 * - `proc`: procedural eyes sampled onto a character grid (the large AIONSEC
 *   eye and its relatives). These need room; below ~90px they turn to mush.
 * - `line`: single-line glyph eyes such as `(o)` whose pupil slides between
 *   the brackets.
 * - `art`: hand-drawn multi-line templates. `~` marks the inside of the eye;
 *   an iris sprite moves around in there and is clipped by the lids. Every
 *   character is placed by hand, which is what keeps mid-size eyes legible.
 *
 * `poke(x, y)` sends a wave out from a point; as it reaches each eye, the eye
 * snaps toward the point, its pupil widens, and eyes near the point flinch.
 */

export const PURPLE = '#bd93f9';
export const YELLOW = '#f1fa8c';
export const BONE = '#e6e1d6';
export const GREY = '#8d8c96';
export const FONT = '"Geist Mono", ui-monospace, SFMono-Regular, Menlo, monospace';

const WAVE_SPEED = 1250; // px per second
const FPS = 30;
const TAU = Math.PI * 2;
const DENSITY = '.:;=+*#%@';
const BANDS = ' .:-=+*#';
const SOFT = new Set(['.', "'", '`', ',', ':']);

export const clamp = (v: number, a = 0, b = 1) => Math.max(a, Math.min(b, v));
export const smooth = (v: number) => {
	v = clamp(v);
	return v * v * v * (v * (v * 6 - 15) + 10);
};
export const mix = (a: number, b: number, t: number) => a + (b - a) * t;
export const rand = (a: number, b: number) => a + Math.random() * (b - a);
export const pick = <T>(list: T[]): T => list[Math.floor(Math.random() * list.length)];
const hash = (a: number, b: number) => {
	const n = Math.sin(a * 12.9898 + b * 78.233) * 43758.5453;
	return n - Math.floor(n);
};

/* ------------------------------------------------------------------ */
/* Palettes                                                            */
/* ------------------------------------------------------------------ */

export interface Palette {
	lid: string;
	iris: string;
	sclera: string;
	rim: string;
}

// Lid / iris pairings of the four site colours: purple, yellow, bone, grey.
export const PALETTES = {
	aion: { lid: PURPLE, iris: YELLOW, sclera: PURPLE, rim: PURPLE },
	ghost: { lid: BONE, iris: PURPLE, sclera: BONE, rim: PURPLE },
	violet: { lid: PURPLE, iris: PURPLE, sclera: PURPLE, rim: YELLOW },
	ember: { lid: BONE, iris: YELLOW, sclera: BONE, rim: YELLOW },
	ash: { lid: GREY, iris: YELLOW, sclera: GREY, rim: GREY },
	slate: { lid: GREY, iris: PURPLE, sclera: GREY, rim: PURPLE },
	sun: { lid: YELLOW, iris: PURPLE, sclera: YELLOW, rim: YELLOW },
	gold: { lid: YELLOW, iris: YELLOW, sclera: YELLOW, rim: PURPLE },
	pearl: { lid: PURPLE, iris: BONE, sclera: PURPLE, rim: PURPLE },
	bone: { lid: BONE, iris: BONE, sclera: BONE, rim: BONE }
} satisfies Record<string, Palette>;

export type PaletteName = keyof typeof PALETTES;

/** How often each pairing appears on the homepage; purple/yellow stays the default. */
export const PALETTE_MIX: Array<[PaletteName, number]> = [
	['aion', 5],
	['ember', 2],
	['ghost', 2],
	['ash', 2],
	['slate', 2],
	['violet', 1],
	['sun', 1],
	['pearl', 1]
];

export function mixedPalette(): Palette {
	const total = PALETTE_MIX.reduce((n, [, w]) => n + w, 0);
	let r = Math.random() * total;
	for (const [name, w] of PALETTE_MIX) {
		r -= w;
		if (r < 0) return PALETTES[name];
	}
	return PALETTES.aion;
}

/* ------------------------------------------------------------------ */
/* Drawings (pure data)                                                */
/* ------------------------------------------------------------------ */

export type Pupil = 'round' | 'slit' | 'bar';
export type IrisStyle = 'fibers' | 'rings' | 'spokes' | 'flat';

export interface ProcSpec {
	half: number; // half-width of the opening, px
	fs: number; // glyph size, px
	topK: number; // upper lid height (fraction of half)
	botK: number; // lower lid depth
	arcPow: number; // lid curvature; lower = rounder
	tilt: number;
	ring: 0 | 1 | 2; // 0 none, 1 outline, 2 outline + dotted inner ring + dial ticks
	ringK: number;
	irisK: number;
	pupilK: number;
	pupil: Pupil;
	iris: IrisStyle;
	lashes: number;
	hood: number; // 0..1 resting droop of the upper lid
	crease: boolean;
	catchlight: boolean;
	sclera: number; // density of the dotted white of the eye, 0..1
	/**
	 * Mid-size mode: lids traced as a single-glyph line (`_.-'` curves), no
	 * crease or sclera, simpler iris. Detail that reads at 300px turns to
	 * mush at 200px, so medium eyes get less of it.
	 */
	simple?: boolean;
}

export interface ProcDrawing {
	kind: 'proc';
	spec: ProcSpec;
}

export interface LineDrawing {
	kind: 'line';
	open: string;
	close: string;
	inner: number;
	pupil: string;
	wide?: string;
	fill?: string; // character laid under the pupil track, e.g. '-' for (-o-)
	fs: number;
}

export interface ArtDrawing {
	kind: 'art';
	rows: string[]; // open eye; '~' marks the inside
	iris: string[]; // sprite; spaces are transparent
	wide?: string[]; // sprite while the pupil is dilated
	closed?: string[]; // blink frame; derived from `rows` when omitted
	range: [number, number]; // how far the iris travels, in cells
	fs: number;
	lh?: number; // line height as a multiple of fs
	pupils?: string; // sprite characters drawn at full strength
	fill?: string; // faint texture for the visible inside of the eye
}

export type Drawing = ProcDrawing | LineDrawing | ArtDrawing;

/* ------------------------------------------------------------------ */
/* Gaze / blink state                                                  */
/* ------------------------------------------------------------------ */

export interface State {
	ax: number;
	ay: number;
	size: number;
	gx: number;
	gy: number;
	fx: number;
	fy: number;
	tx: number;
	ty: number;
	moveStart: number;
	moveDur: number;
	nextLook: number;
	blinkStart: number;
	blinkDur: number;
	nextBlink: number;
	dilKick: number;
	attentionUntil: number;
	pending: { at: number; x: number; y: number; flinch: boolean } | null;
	lazy: number;
	closure: number;
	dil: number;
}

export function makeState(time: number, lazy = 1): State {
	return {
		ax: 0,
		ay: 0,
		size: 1,
		gx: 0,
		gy: 0,
		fx: 0,
		fy: 0,
		tx: 0,
		ty: 0,
		moveStart: time,
		moveDur: 0.4,
		nextLook: time + rand(0.6, 2.6),
		blinkStart: -10,
		blinkDur: 0.3,
		nextBlink: time + rand(1.2, 6),
		dilKick: -10,
		attentionUntil: 0,
		pending: null,
		lazy,
		closure: 0,
		dil: 1
	};
}

/* ------------------------------------------------------------------ */
/* Eye instances and groups                                            */
/* ------------------------------------------------------------------ */

interface Base {
	st: State;
	x: number;
	y: number;
	pal: Palette;
	alpha: number;
	phase: number;
	rx: number;
	ry: number;
}

export interface ProcEye extends Base {
	type: 'proc';
	spec: ProcSpec;
	cellW: number;
	cellH: number;
	ox: number;
	oy: number;
	rangeX: number;
	rangeY: number;
}

export interface LineEye extends Base {
	type: 'line';
	d: LineDrawing;
}

export interface ArtEye extends Base {
	type: 'art';
	p: PreparedArt;
}

export type EyeInst = ProcEye | LineEye | ArtEye;

/** Slow wandering offset so eyes float a little in space. */
export interface Drift {
	ax: number;
	ay: number;
	f: number[];
	p: number[];
}

export function makeDrift(amp = rand(6, 12)): Drift {
	return {
		ax: amp,
		ay: amp * rand(0.6, 1),
		// Two incommensurate sines per axis: never quite repeats, never jerks.
		f: [rand(0.13, 0.25), rand(0.31, 0.47), rand(0.11, 0.22), rand(0.27, 0.43)],
		p: [rand(0, TAU), rand(0, TAU), rand(0, TAU), rand(0, TAU)]
	};
}

/** One or more eyes that share a gaze (a pair of eyes moves together). */
export interface Group {
	st: State;
	eyes: EyeInst[];
	offsets: number[];
	rx: number;
	ry: number;
	hidden: boolean;
	/** Resting position; the drift is added on top when drawing. */
	x: number;
	y: number;
	drift: Drift | null;
	/** Scene time this group starts fading in (-Infinity: visible, Infinity: hidden). */
	fadeStart: number;
}

export interface Rect {
	left: number;
	top: number;
	right: number;
	bottom: number;
}

function positionGroup(g: Group, dx: number, dy: number) {
	const x = g.x + dx;
	const y = g.y + dy;
	g.st.ax = x;
	g.st.ay = y;
	g.eyes.forEach((e, i) => {
		e.x = x + g.offsets[i];
		e.y = y;
		if (e.type === 'proc') {
			e.ox = e.x;
			e.oy = e.y;
		}
	});
}

export function placeGroup(g: Group, x: number, y: number) {
	g.x = x;
	g.y = y;
	positionGroup(g, 0, 0);
}

/* ------------------------------------------------------------------ */
/* Art templates                                                       */
/* ------------------------------------------------------------------ */

export interface PreparedArt {
	open: string[];
	closed: string[];
	comp: Int16Array[]; // component id of each inside cell, -1 elsewhere
	centers: Array<[number, number]>;
	cols: number;
	rows: number;
	iris: string[];
	wide: string[];
	range: [number, number];
	pupils: string;
	fill: string;
	fs: number;
	lh: number;
}

const MIRROR: Record<string, string> = {
	'/': '\\',
	'\\': '/',
	'(': ')',
	')': '(',
	'<': '>',
	'>': '<',
	'[': ']',
	']': '[',
	'{': '}',
	'}': '{'
};

function mirrorRows(rows: string[]): string[] {
	const w = Math.max(...rows.map((r) => r.length));
	return rows.map((r) =>
		r
			.padEnd(w)
			.split('')
			.reverse()
			.map((c) => MIRROR[c] ?? c)
			.join('')
	);
}

const mirrored = new WeakMap<ArtDrawing, ArtDrawing>();
export function mirrorArt(a: ArtDrawing): ArtDrawing {
	let m = mirrored.get(a);
	if (!m) {
		m = {
			...a,
			rows: mirrorRows(a.rows),
			iris: mirrorRows(a.iris),
			wide: a.wide && mirrorRows(a.wide),
			closed: a.closed && mirrorRows(a.closed)
		};
		mirrored.set(a, m);
	}
	return m;
}

const prepared = new WeakMap<ArtDrawing, PreparedArt>();
export function prepareArt(a: ArtDrawing): PreparedArt {
	const cached = prepared.get(a);
	if (cached) return cached;

	const rowCount = Math.max(a.rows.length, a.closed?.length ?? 0);
	const cols = Math.max(...a.rows.map((r) => r.length), ...(a.closed ?? []).map((r) => r.length));
	const grid = Array.from({ length: rowCount }, (_, r) => (a.rows[r] ?? '').padEnd(cols));

	// Label each connected region of '~' so templates can hold several eyes.
	const comp = grid.map(() => new Int16Array(cols).fill(-1));
	const boxes: Array<[number, number, number, number]> = [];
	for (let r = 0; r < rowCount; r++) {
		for (let c = 0; c < cols; c++) {
			if (grid[r][c] !== '~' || comp[r][c] !== -1) continue;
			const id = boxes.length;
			const box: [number, number, number, number] = [c, r, c, r];
			const stack: Array<[number, number]> = [[r, c]];
			comp[r][c] = id;
			while (stack.length) {
				const [rr, cc] = stack.pop()!;
				box[0] = Math.min(box[0], cc);
				box[1] = Math.min(box[1], rr);
				box[2] = Math.max(box[2], cc);
				box[3] = Math.max(box[3], rr);
				for (const [dr, dc] of [
					[1, 0],
					[-1, 0],
					[0, 1],
					[0, -1]
				]) {
					const nr = rr + dr;
					const nc = cc + dc;
					if (grid[nr]?.[nc] === '~' && comp[nr][nc] === -1) {
						comp[nr][nc] = id;
						stack.push([nr, nc]);
					}
				}
			}
			boxes.push(box);
		}
	}
	const centers = boxes.map(
		([c0, r0, c1, r1]) => [Math.round((c0 + c1) / 2), Math.round((r0 + r1) / 2)] as [number, number]
	);

	const open = grid.map((r) => r.replace(/~/g, ' '));
	let closed: string[];
	if (a.closed) {
		closed = Array.from({ length: rowCount }, (_, r) => (a.closed![r] ?? '').padEnd(cols));
	} else {
		// Default blink: the inside empties and a lid line runs through the middle.
		closed = open.map((row, r) =>
			row
				.split('')
				.map((ch, c) => {
					const id = comp[r][c];
					return id >= 0 && centers[id][1] === r ? '-' : ch;
				})
				.join('')
		);
	}

	const p: PreparedArt = {
		open,
		closed,
		comp,
		centers,
		cols,
		rows: rowCount,
		iris: a.iris,
		wide: a.wide ?? a.iris,
		range: a.range,
		pupils: a.pupils ?? 'o@O0*|=#',
		fill: a.fill ?? '',
		fs: a.fs,
		lh: a.lh ?? 1.15
	};
	prepared.set(a, p);
	return p;
}

/* ------------------------------------------------------------------ */
/* Geometry                                                            */
/* ------------------------------------------------------------------ */

export function procExtent(spec: ProcSpec): [number, number] {
	if (spec.ring) {
		const r = spec.half * spec.ringK + spec.fs * (spec.ring === 2 ? 2.2 : 1);
		return [Math.max(r, spec.half), r];
	}
	return [spec.half, spec.half * Math.max(spec.topK, spec.botK) + spec.fs * (spec.lashes ? 2.6 : 1.2)];
}

/** Half extents of a single eye, in px. */
export function drawingExtent(d: Drawing): [number, number] {
	if (d.kind === 'proc') return procExtent(d.spec);
	if (d.kind === 'line') {
		const n = d.open.length + d.inner + d.close.length;
		return [(n * d.fs * 0.6) / 2, d.fs * 0.6];
	}
	const p = prepareArt(d);
	return [(p.cols * p.fs * 0.6) / 2, (p.rows * p.fs * p.lh) / 2];
}

export function makeEye(d: Drawing, st: State, pal: Palette, alpha: number): EyeInst {
	const [rx, ry] = drawingExtent(d);
	const base = { st, x: 0, y: 0, pal, alpha, phase: rand(0, 100), rx, ry };
	if (d.kind === 'proc') {
		const s = d.spec;
		return {
			...base,
			type: 'proc',
			spec: s,
			cellW: s.fs * 0.6,
			cellH: s.fs * 0.98,
			ox: 0,
			oy: 0,
			rangeX: (1 - s.irisK) * 0.31,
			rangeY: Math.max(0.04, Math.min(s.topK * (1 - s.hood * 0.6), s.botK) * 0.26)
		};
	}
	if (d.kind === 'line') return { ...base, type: 'line', d };
	return { ...base, type: 'art', p: prepareArt(d) };
}

/* ------------------------------------------------------------------ */
/* Scene                                                               */
/* ------------------------------------------------------------------ */

interface Ping {
	x: number;
	y: number;
	t: number;
}

interface Atlas {
	canvas: HTMLCanvasElement;
	w: number;
	h: number;
	index: Map<string, number>;
}

const CHARSET = ' .:;=+*#%@-_|/\\`\'oO0()<>[]{},^~"';

export class EyeScene {
	readonly canvas: HTMLCanvasElement;
	readonly reduced: boolean;
	time = 0;
	width = 0;
	height = 0;
	groups: Group[] = [];
	/** Called after every resize (including the first). */
	onResize: (() => void) | null = null;
	/** Called before every render, e.g. to follow elements as the page scrolls. */
	beforeRender: (() => void) | null = null;
	/**
	 * Areas (usually page text) where eyes fade back: glyphs inside the rect,
	 * plus a small margin, are drawn at `dimAlpha`, easing to full strength
	 * across `dimFeather` px. Feels like an invisible frame around the text.
	 */
	dimRects: Rect[] = [];
	dimAlpha = 0.3;
	dimPad = 14;
	dimFeather = 40;
	/** Seconds each group takes to fade in. */
	fadeDur = 0.9;

	private ctx: CanvasRenderingContext2D;
	private pings: Ping[] = [];
	private pointer: { x: number; y: number; t: number } | null = null;
	private dpr = 1;
	private fs = 11;
	private alphaMul = 1;
	private atlases = new Map<string, Atlas>();
	private raf = 0;
	private lastStamp = 0;
	private lastDraw = -Infinity;
	private dirty = false;
	private destroyed = false;
	private cleanups: Array<() => void> = [];

	constructor(canvas: HTMLCanvasElement) {
		this.canvas = canvas;
		this.ctx = canvas.getContext('2d')!;
		this.reduced = matchMedia('(prefers-reduced-motion: reduce)').matches;
	}

	start() {
		let timer: ReturnType<typeof setTimeout> | undefined;
		const onResize = () => {
			clearTimeout(timer);
			timer = setTimeout(() => this.resize(), 120);
		};
		const onMove = (ev: PointerEvent) => {
			this.pointer = { x: ev.clientX, y: ev.clientY, t: this.time };
		};
		// If the web font lands after we started, rebuild sprites with it.
		const onFonts = () => {
			this.atlases.clear();
			this.invalidate();
		};
		window.addEventListener('resize', onResize);
		window.addEventListener('pointermove', onMove, { passive: true });
		document.fonts?.addEventListener('loadingdone', onFonts);
		this.cleanups.push(() => {
			clearTimeout(timer);
			window.removeEventListener('resize', onResize);
			window.removeEventListener('pointermove', onMove);
			document.fonts?.removeEventListener('loadingdone', onFonts);
		});

		const go = () => {
			if (this.destroyed) return;
			this.resize();
			if (!this.reduced) this.raf = requestAnimationFrame(this.frame);
		};
		Promise.race([
			document.fonts?.load(`400 11px ${FONT}`) ?? Promise.resolve(),
			new Promise((r) => setTimeout(r, 600))
		]).then(go, go);
	}

	destroy() {
		this.destroyed = true;
		cancelAnimationFrame(this.raf);
		for (const fn of this.cleanups) fn();
	}

	/** Draw on the next animation frame even if the frame budget says wait. */
	invalidate() {
		this.dirty = true;
		if (this.reduced) requestAnimationFrame(() => this.render());
	}

	resize() {
		const rect = this.canvas.getBoundingClientRect();
		this.width = rect.width;
		this.height = rect.height;
		this.dpr = Math.min(window.devicePixelRatio || 1, 2);
		this.canvas.width = Math.round(this.width * this.dpr);
		this.canvas.height = Math.round(this.height * this.dpr);
		this.ctx.setTransform(this.dpr, 0, 0, this.dpr, 0, 0);
		this.onResize?.();
		for (const g of this.groups) this.update(g.st);
		this.render();
	}

	poke(x: number, y: number) {
		if (this.reduced) {
			for (const g of this.groups) {
				const [tx, ty] = this.aim(g.st, x, y);
				g.st.fx = g.st.gx = g.st.tx = tx;
				g.st.fy = g.st.gy = g.st.ty = ty;
			}
			this.render();
			return;
		}
		this.pings.push({ x, y, t: this.time });
		for (const g of this.groups) {
			const st = g.st;
			const d = Math.hypot(x - st.ax, y - st.ay);
			st.pending = {
				at: this.time + d / WAVE_SPEED + rand(0, 0.06),
				x,
				y,
				flinch: d < st.size * 1.05
			};
		}
	}

	private frame = (stamp: number) => {
		if (this.destroyed) return;
		this.raf = requestAnimationFrame(this.frame);
		if (!this.dirty && stamp - this.lastDraw < 1000 / FPS - 2) return;
		this.dirty = false;
		const dt = this.lastStamp ? Math.min((stamp - this.lastStamp) / 1000, 0.1) : 0;
		this.lastStamp = stamp;
		this.lastDraw = stamp;
		this.time += dt;
		for (const g of this.groups) this.update(g.st);
		this.pings = this.pings.filter((p) => this.time - p.t < 1.5);
		this.render();
	};

	/* ---------------- motion ---------------- */

	private lookTo(st: State, tx: number, ty: number, dur: number) {
		st.fx = st.gx;
		st.fy = st.gy;
		st.tx = clamp(tx, -1, 1);
		st.ty = clamp(ty, -1, 1);
		st.moveStart = this.time;
		st.moveDur = dur;
	}

	/** Normalised gaze from a group's anchor toward a point. */
	private aim(st: State, px: number, py: number): [number, number] {
		const dx = px - st.ax;
		const dy = py - st.ay;
		const d = Math.hypot(dx, dy) || 1;
		const f = smooth((d - st.size * 0.12) / (st.size * 1.4));
		return [(dx / d) * f, (dy / d) * f];
	}

	private beginBlink(st: State, forced: boolean) {
		st.blinkStart = this.time;
		st.blinkDur = rand(0.24, 0.34);
		const double = !forced && Math.random() < 0.14;
		st.nextBlink = this.time + (double ? st.blinkDur + 0.14 : rand(2.8, 8.5) * st.lazy);
	}

	private update(st: State) {
		const time = this.time;
		if (st.pending && time >= st.pending.at) {
			const p = st.pending;
			st.pending = null;
			const [tx, ty] = this.aim(st, p.x, p.y);
			this.lookTo(st, tx, ty, rand(0.12, 0.2));
			st.attentionUntil = time + rand(3.2, 5.6);
			st.nextLook = st.attentionUntil;
			st.dilKick = time;
			if (p.flinch) this.beginBlink(st, true);
		}
		if (time >= st.nextLook && time >= st.attentionUntil) {
			const r = Math.random();
			let tx: number, ty: number;
			if (this.pointer && time - this.pointer.t < 6 && r < 0.2) {
				[tx, ty] = this.aim(st, this.pointer.x, this.pointer.y);
			} else if (r < 0.45) {
				tx = rand(-0.25, 0.25);
				ty = rand(-0.3, 0.3);
			} else {
				tx = rand(-1, 1);
				ty = rand(-0.9, 0.9);
			}
			this.lookTo(st, tx, ty, rand(0.28, 0.6));
			st.nextLook = time + st.moveDur + rand(1.3, 3.8) * st.lazy;
		}
		const s = smooth((time - st.moveStart) / st.moveDur);
		st.gx = mix(st.fx, st.tx, s);
		st.gy = mix(st.fy, st.ty, s);
		if (!this.reduced && time >= st.nextBlink) this.beginBlink(st, false);
		if (this.reduced) {
			st.closure = 0;
		} else {
			const t = (time - st.blinkStart) / st.blinkDur;
			st.closure =
				t < 0 || t >= 1 ? 0 : t < 0.29 ? smooth(t / 0.29) : t < 0.4 ? 1 : 1 - smooth((t - 0.4) / 0.6);
		}
		const k = time - st.dilKick;
		st.dil = 1 + 0.42 * (k < 0 ? 0 : k < 0.18 ? smooth(k / 0.18) : Math.exp(-(k - 0.18) * 1.3));
	}

	/* ---------------- drawing ---------------- */

	// Thousands of glyphs per frame: blitting pre-rendered sprites is several
	// times cheaper than fillText, so each (size, colour) gets a sprite strip.
	private atlasFor(size: number, color: string): Atlas {
		const key = `${size}|${color}|${this.dpr}`;
		let atlas = this.atlases.get(key);
		if (atlas) return atlas;
		const w = Math.ceil(size * 1.2 * this.dpr);
		const h = Math.ceil(size * 1.6 * this.dpr);
		const canvas = document.createElement('canvas');
		canvas.width = w * CHARSET.length;
		canvas.height = h;
		const g = canvas.getContext('2d')!;
		g.font = `400 ${size * this.dpr}px ${FONT}`;
		g.textAlign = 'center';
		g.textBaseline = 'middle';
		g.fillStyle = color;
		const index = new Map<string, number>();
		for (let i = 0; i < CHARSET.length; i++) {
			g.fillText(CHARSET[i], i * w + w / 2, h / 2);
			index.set(CHARSET[i], i);
		}
		atlas = { canvas, w, h, index };
		this.atlases.set(key, atlas);
		return atlas;
	}

	private dimAt(x: number, y: number) {
		let f = 1;
		for (const r of this.dimRects) {
			const dx = Math.max(r.left - x, 0, x - r.right);
			const dy = Math.max(r.top - y, 0, y - r.bottom);
			const d = dx === 0 && dy === 0 ? 0 : Math.hypot(dx, dy);
			f = Math.min(f, mix(this.dimAlpha, 1, smooth((d - this.dimPad) / this.dimFeather)));
		}
		return f;
	}

	/** 0..1 visibility of a group as it fades in. */
	private fadeOf(g: Group) {
		if (g.fadeStart === -Infinity) return 1;
		if (this.reduced) return this.time >= g.fadeStart ? 1 : 0;
		return smooth((this.time - g.fadeStart) / this.fadeDur);
	}

	private glyph(ch: string, x: number, y: number, color: string, alpha: number) {
		alpha *= this.alphaMul;
		if (this.dimRects.length) alpha *= this.dimAt(x, y);
		if (alpha < 0.01) return;
		const ctx = this.ctx;
		ctx.globalAlpha = alpha > 1 ? 1 : alpha;
		const atlas = this.atlasFor(this.fs, color);
		const i = atlas.index.get(ch);
		if (i === undefined) {
			ctx.font = `400 ${this.fs}px ${FONT}`;
			ctx.fillStyle = color;
			ctx.fillText(ch, x, y);
			return;
		}
		const dpr = this.dpr;
		const dw = atlas.w / dpr;
		const dh = atlas.h / dpr;
		const dx = Math.round((x - dw / 2) * dpr) / dpr;
		const dy = Math.round((y - dh / 2) * dpr) / dpr;
		ctx.drawImage(atlas.canvas, i * atlas.w, 0, atlas.w, atlas.h, dx, dy, dw, dh);
	}

	render() {
		this.beforeRender?.();
		if (!this.reduced) {
			const t = this.time;
			for (const g of this.groups) {
				const d = g.drift;
				if (!d) continue;
				const dx = d.ax * (0.6 * Math.sin(t * d.f[0] + d.p[0]) + 0.4 * Math.sin(t * d.f[1] + d.p[1]));
				const dy = d.ay * (0.6 * Math.sin(t * d.f[2] + d.p[2]) + 0.4 * Math.sin(t * d.f[3] + d.p[3]));
				positionGroup(g, dx, dy);
			}
		}
		const ctx = this.ctx;
		ctx.clearRect(0, 0, this.width, this.height);
		ctx.textAlign = 'center';
		ctx.textBaseline = 'middle';
		const W = this.width;
		const H = this.height;
		for (const g of this.groups) {
			if (g.hidden) continue;
			const fade = this.fadeOf(g);
			if (fade <= 0) continue;
			this.alphaMul = fade;
			for (const e of g.eyes) {
				if (e.x + e.rx < -20 || e.x - e.rx > W + 20 || e.y + e.ry < -40 || e.y - e.ry > H + 40) continue;
				if (e.type === 'proc') this.drawProc(e);
				else if (e.type === 'line') this.drawLine(e);
				else this.drawArt(e);
			}
		}
		this.alphaMul = 1;
		for (const p of this.pings) this.drawPing(p);
		ctx.globalAlpha = 1;
	}

	/** Medium procedural eye: thin traced lids and a two-tone iris. */
	private drawSimple(e: ProcEye) {
		const s = e.spec;
		const st = e.st;
		const { half } = s;
		const cellW = e.cellW;
		const cellH = e.cellH;
		const A = e.alpha;
		const pal = e.pal;
		const closure = st.closure;
		const gazeX = st.gx * e.rangeX;
		const gazeY = st.gy * e.rangeY;
		const ph = this.time + e.phase;
		const cx = e.x + half * (gazeX * 0.1 + Math.sin(ph * 0.57) * 0.012);
		const cy = e.y + half * (gazeY * 0.14 + Math.sin(ph * 0.41) * 0.01);
		const tiltSlope = s.tilt + Math.sin(ph * 0.46) * 0.009 + gazeX * 0.025;
		const irisR = half * s.irisK;
		const ix = gazeX * half;
		const iy = gazeY * half;
		const pupilR = Math.min(irisR * 0.75, irisR * s.pupilK * st.dil);
		const seamK = s.botK * 0.27;
		const ox = e.ox;
		const oy = e.oy;
		this.fs = s.fs;

		const lidY = (x: number, top: boolean) => {
			const u = x / half;
			const arc = Math.pow(Math.max(0, 1 - u * u), s.arcPow);
			const tilt = x * tiltSlope;
			const seam = half * seamK * arc + tilt;
			if (top) return mix(mix(-half * s.topK * arc + tilt + iy * 0.18 * arc, seam, s.hood), seam, closure);
			return mix(half * s.botK * arc + tilt + iy * 0.1 * arc, seam, closure);
		};
		const colOf = (px: number) => Math.round((px - ox) / cellW);
		const rowOf = (py: number) => Math.round((py - oy) / cellH);
		const key = (c: number, r: number) => (r + 2048) * 4096 + (c + 2048);

		// 1. Trace each lid one glyph thick. Where the curve is shallow the glyph
		// follows its height within the cell (_ . - '), which is what makes ASCII
		// curves look smooth; where it is steep, slashes keep it connected.
		const outline = new Map<number, [string, number]>();
		const set = (c: number, r: number, g: string, a: number, force = false) => {
			const k = key(c, r);
			if (force || !outline.has(k)) outline.set(k, [g, a]);
		};
		const cL = colOf(cx - half);
		const cR = colOf(cx + half);
		const round = s.arcPow < 0.6;
		const fillGlyph = (left: boolean, dir: number) => (round ? (left ? '(' : ')') : dir < 0 ? '/' : '\\');

		const trace = (top: boolean, alpha: number) => {
			let prev = rowOf(cy + lidY(-half, top));
			for (let c = cL + 1; c < cR; c++) {
				const x = ox + c * cellW - cx;
				const rowF = (cy + lidY(x, top) - oy) / cellH;
				const r = Math.round(rowF);
				const frac = rowF - r;
				const slope = (lidY(x + cellW * 0.5, top) - lidY(x - cellW * 0.5, top)) / cellH;
				let g: string;
				let gr = r;
				if (Math.abs(slope) < 0.9) {
					const flat = Math.abs(slope) < 0.3;
					if (top && flat && s.hood > 0.2 && closure < 0.5 && Math.abs(x) < half * 0.45) g = '=';
					else if (top && flat && frac < -0.18) {
						// A flat run near the top of the cell reads better as ____ on
						// the row above than as a dotted run of apostrophes.
						g = '_';
						gr = r - 1;
					} else if (!top && frac < -0.18) g = '-';
					else g = frac > 0.28 ? '_' : frac > 0.08 ? '.' : frac > -0.18 ? '-' : "'";
				} else {
					g = slope < 0 ? '/' : '\\';
				}
				if (Math.abs(r - prev) > 1) {
					const dir = Math.sign(r - prev);
					const fc = x < 0 ? c - 1 : c;
					for (let rr = prev + dir; rr !== r; rr += dir) set(fc, rr, fillGlyph(x < 0, dir), alpha);
				}
				set(c, gr, g, alpha);
				prev = r;
			}
			const end = rowOf(cy + lidY(half, top));
			if (Math.abs(end - prev) > 1) {
				const dir = Math.sign(end - prev);
				for (let rr = prev + dir; rr !== end; rr += dir) set(cR, rr, fillGlyph(false, dir), alpha);
			}
		};
		trace(true, 1);
		if (closure < 0.85) trace(false, 0.8);
		set(cL, rowOf(cy + lidY(-half, true)), round ? '(' : '<', 0.95, true);
		set(cR, rowOf(cy + lidY(half, true)), round ? ')' : '>', 0.95, true);

		// 2. Iris and pupil, kept clear of the lid lines.
		if (closure < 0.9) {
			const rTop = rowOf(cy - half * s.topK) - 1;
			const rBot = rowOf(cy + half * s.botK) + 1;
			for (let r = rTop; r <= rBot; r++) {
				const py = oy + r * cellH;
				const y = py - cy;
				for (let c = cL + 1; c < cR; c++) {
					if (outline.has(key(c, r))) continue;
					const px = ox + c * cellW;
					const x = px - cx;
					if (y <= lidY(x, true) + cellH * 0.35 || y >= lidY(x, false) - cellH * 0.35) continue;
					const dx = (x - ix) / (1 - Math.abs(gazeX) * 0.2);
					const dy = y - iy;
					const rr = Math.hypot(dx, dy);
					const rn = rr / irisR;
					if (rn >= 1.02) continue;

					let inPupil: boolean;
					let border: boolean;
					if (s.pupil === 'round') {
						inPupil = rr < pupilR;
						border = !inPupil && rr - pupilR < cellW * 0.95;
					} else if (s.pupil === 'slit') {
						const w = Math.max(cellW * 0.5, pupilR * 0.4);
						const h = irisR * 0.85;
						const sx = w * Math.sqrt(Math.max(0, 1 - (dy / h) ** 2));
						inPupil = Math.abs(dx) < sx && Math.abs(dy) < h;
						border = !inPupil && Math.abs(dx) < sx + cellW && Math.abs(dy) < h;
					} else {
						const w = irisR * 0.78;
						const h = Math.max(cellH * 0.45, pupilR * 0.42);
						const sy = h * Math.sqrt(Math.max(0, 1 - (dx / w) ** 2));
						inPupil = Math.abs(dy) < sy && Math.abs(dx) < w;
						border = !inPupil && Math.abs(dy) < sy + cellH * 0.8 && Math.abs(dx) < w;
					}
					if (inPupil) continue;

					let ch: string;
					let color = pal.iris;
					let alpha: number;
					if (rn > 0.9) {
						ch = 'oO@'[Math.min(2, Math.floor((1.02 - rn) * 19))];
						color = pal.rim;
						alpha = 0.8;
					} else if (border) {
						ch = s.pupil === 'round' ? '*' : s.pupil === 'slit' ? '|' : '=';
						alpha = 0.95;
					} else if (s.iris === 'rings') {
						const light = clamp(0.5 + 0.5 * Math.sin(rn * 21 - ph * 0.9));
						ch = BANDS[1 + Math.floor(light * (BANDS.length - 2))];
						alpha = 0.55 + light * 0.45;
					} else if (s.iris === 'spokes') {
						const light = clamp(0.4 + 0.45 * Math.sin(Math.atan2(dy, dx) * 8 + ph * 0.25));
						ch = light > 0.55 ? '+' : ':';
						alpha = 0.55 + light * 0.4;
					} else {
						// flat (and the fallback): two calm bands
						const inner = (rn - pupilR / irisR) / Math.max(0.1, 0.9 - pupilR / irisR) < 0.5;
						ch = inner ? '+' : ':';
						alpha = inner ? 0.9 : 0.62;
					}
					if (s.catchlight) {
						const spot = ((dx + irisR * 0.32) / (irisR * 0.16)) ** 2 + ((dy + irisR * 0.3) / (irisR * 0.16)) ** 2;
						if (spot < 1) {
							ch = '#';
							color = pal.iris;
							alpha = 1;
						}
					}
					this.glyph(ch, px, py, color, alpha * A);
				}
			}
		}

		// 3. Lid line on top of everything, then lashes riding on it.
		for (const [k, [g, a]] of outline) {
			const c = (k % 4096) - 2048;
			const r = Math.floor(k / 4096) - 2048;
			this.glyph(g, ox + c * cellW, oy + r * cellH, pal.lid, a * A);
		}
		if (s.lashes) {
			for (let i = 0; i < s.lashes; i++) {
				const u = mix(-0.55, 0.55, s.lashes === 1 ? 0.5 : i / (s.lashes - 1));
				const x = u * half;
				const c = colOf(cx + x);
				// One row above whatever the lid line actually drew in this column.
				let r = rowOf(cy + lidY(x, true));
				while (outline.has(key(c, r - 1))) r--;
				r -= 1;
				const g = u < -0.2 ? '\\' : u > 0.2 ? '/' : '|';
				this.glyph(g, ox + c * cellW, oy + r * cellH, pal.lid, 0.6 * A);
			}
		}
	}

	private drawProc(e: ProcEye) {
		const s = e.spec;
		if (s.simple) {
			this.drawSimple(e);
			return;
		}
		const st = e.st;
		const { half } = s;
		const cellW = e.cellW;
		const cellH = e.cellH;
		const A = e.alpha;
		const pal = e.pal;
		const closure = st.closure;
		const gazeX = st.gx * e.rangeX;
		const gazeY = st.gy * e.rangeY;
		const ph = this.time + e.phase;

		// The whole eye leans a little toward where it is looking.
		const swayX = half * (gazeX * 0.1 + Math.sin(ph * 0.57) * 0.012);
		const swayY = half * (gazeY * 0.14 + Math.sin(ph * 0.41) * 0.01);
		const cx = e.x + swayX;
		const cy = e.y + swayY;
		const tiltSlope = s.tilt + Math.sin(ph * 0.46) * 0.009 + gazeX * 0.025;
		const ringStretch = 1 + Math.sin(ph * 0.81) * 0.025 - closure * 0.014;
		const outerR = half * s.ringK;
		const irisR = half * s.irisK;
		const ix = gazeX * half + Math.sin(ph * 1.19) * half * 0.0018;
		const iy = gazeY * half + Math.sin(ph * 0.93 + 0.8) * half * 0.0012;
		const pupilR = Math.min(irisR * 0.8, irisR * s.pupilK * (1 + Math.sin(ph * 0.38) * 0.025) * st.dil);
		const seamK = s.botK * 0.27;

		const extX = (s.ring ? Math.max(outerR, half) : half) * 1.03 + cellW;
		const extY = s.ring
			? outerR * 1.05 + cellH * 2
			: half * Math.max(s.topK, s.botK) * 1.15 + cellH * (s.lashes ? 3.5 : 2);
		const c0 = Math.floor((cx - extX - e.ox) / cellW);
		const c1 = Math.ceil((cx + extX - e.ox) / cellW);
		const r0 = Math.floor((cy - extY - e.oy) / cellH);
		const r1 = Math.ceil((cy + extY - e.oy) / cellH);

		this.fs = s.fs;

		for (let row = r0; row <= r1; row++) {
			const py = e.oy + row * cellH;
			const y = py - cy;
			for (let col = c0; col <= c1; col++) {
				const px = e.ox + col * cellW;
				const x = px - cx;
				const u = x / half;
				const arc = Math.pow(Math.max(0, 1 - u * u), s.arcPow);
				const tilt = x * tiltSlope;
				const seam = half * seamK * arc + tilt;
				const restTop = mix(-half * s.topK * arc + tilt + iy * 0.18 * arc, seam, s.hood);
				const restBottom = half * s.botK * arc + tilt + iy * 0.1 * arc;
				const top = mix(restTop, seam, closure);
				const bottom = mix(restBottom, seam, closure);
				const edge = Math.min(Math.abs(y - top), Math.abs(y - bottom));
				const inside = Math.abs(u) < 1 && y > top && y < bottom && closure < 0.97;
				const onLid = Math.abs(u) <= 1 && edge < cellH * 0.53;

				// Ring sits behind the eye; the opening and lids mask it.
				if (s.ring && !inside && !onLid) {
					const ringY = (y - tilt) / ringStretch;
					const radius = Math.hypot(x, ringY);
					const nx = x / (radius || 1);
					const ny = ringY / (radius || 1);
					const contour = Math.hypot(cellW * nx, cellH * ny);
					const outerEdge = Math.abs(radius - outerR);
					const innerEdge = Math.abs(radius - (outerR - cellW * 1.8));
					if (outerEdge < contour * 0.52) {
						let ch: string;
						if (Math.abs(ny) > 0.93) ch = ringY < 0 ? '_' : '-';
						else if (Math.abs(nx) > 0.94) ch = '|';
						else ch = x * ringY > 0 ? '/' : '\\';
						this.glyph(ch, px, py, pal.lid, 0.86 * A);
					} else if (s.ring === 2 && innerEdge < contour * 0.36) {
						this.glyph('.', px, py, pal.lid, 0.23 * A);
					}
				}

				if (Math.abs(u) > 1.025) continue;

				let ch = '';
				let color = pal.lid;
				let alpha = 1;

				if (inside) {
					const dx = (x - ix) / (1 - Math.abs(gazeX) * 0.2);
					const dy = y - iy;
					const r = Math.hypot(dx, dy);
					const rn = r / irisR;
					if (rn < 1.02) {
						let inPupil = false;
						let border = false;
						if (s.pupil === 'round') {
							inPupil = r < pupilR;
							border = !inPupil && r - pupilR < cellW * 0.95;
						} else if (s.pupil === 'slit') {
							const w = Math.max(cellW * 0.55, pupilR * 0.42);
							const h = irisR * 0.9;
							const sx = w * Math.sqrt(Math.max(0, 1 - (dy / h) ** 2));
							inPupil = Math.abs(dx) < sx && Math.abs(dy) < h;
							border = !inPupil && Math.abs(dx) < sx + cellW && Math.abs(dy) < h + cellH * 0.4;
						} else {
							const w = irisR * 0.82;
							const h = Math.max(cellH * 0.5, pupilR * 0.45);
							const sy = h * Math.sqrt(Math.max(0, 1 - (dx / w) ** 2));
							inPupil = Math.abs(dy) < sy && Math.abs(dx) < w;
							border = !inPupil && Math.abs(dy) < sy + cellH * 0.85 && Math.abs(dx) < w + cellW * 0.6;
						}

						if (inPupil) {
							ch = '';
						} else if (rn > 0.91) {
							ch = 'oO@'[Math.min(2, Math.floor((1.02 - rn) * 19))];
							color = pal.rim;
							alpha = 0.78;
						} else {
							const angle = Math.atan2(dy, dx);
							let light: number;
							if (s.iris === 'fibers') {
								const fibers =
									0.5 +
									0.23 * Math.sin(angle * 41 + rn * 8) +
									0.17 * Math.sin(angle * 79 - rn * 14) +
									0.1 * Math.sin(angle * 121 + rn * 24);
								const innerRing = Math.exp(-Math.pow((rn - 0.45) / 0.065, 2));
								light = clamp(0.2 + fibers * 0.62 + innerRing * 0.22 + ((-dx - dy) / irisR) * 0.1);
								ch = DENSITY[Math.floor(light * (DENSITY.length - 1))];
							} else if (s.iris === 'rings') {
								light = clamp(0.5 + 0.5 * Math.sin(rn * 21 - ph * 0.9));
								ch = BANDS[1 + Math.floor(light * (BANDS.length - 2))];
							} else {
								const spokes = Math.sin(angle * 12 + ph * 0.25);
								light = clamp(0.35 + 0.38 * spokes * (1.1 - rn * 0.5) + 0.2 * (1 - rn));
								ch = DENSITY[Math.floor(light * (DENSITY.length - 1))];
							}
							color = pal.iris;
							alpha = 0.55 + light * 0.45;
							if (border) {
								ch = s.pupil === 'round' ? '*' : ':';
								alpha = 0.95;
							}
						}
						if (s.catchlight && !inPupil) {
							const spot =
								((dx + irisR * 0.3) / (irisR * 0.12)) ** 2 + ((dy + irisR * 0.31) / (irisR * 0.11)) ** 2;
							if (spot < 1) {
								ch = '#';
								color = pal.iris;
								alpha = 1;
							}
						}
					} else if (s.sclera > 0) {
						const nearLid = Math.min(y - top, bottom - y) / half;
						const value = clamp(0.15 + u * u * 0.5 + (1 - clamp(nearLid / 0.13)) * 0.31);
						if (hash(col, row) < (0.34 + value * 0.4) * s.sclera) {
							ch = value > 0.57 ? ':' : '.';
							color = pal.sclera;
							alpha = 0.19 + value * 0.32;
						}
					}
				}

				if (onLid) {
					const isTop = Math.abs(y - top) < Math.abs(y - bottom);
					if (Math.abs(u) < 0.27) ch = isTop ? '_' : '-';
					else if (Math.abs(u) < 0.65) ch = isTop ? '.' : u < 0 ? '`' : "'";
					else ch = isTop ? (u < 0 ? '/' : '\\') : u < 0 ? '\\' : '/';
					if (closure > 0.8) ch = Math.abs(u) < 0.72 ? '_' : '-';
					color = pal.lid;
					alpha = isTop ? 1 : 0.8;
				}

				// Quiet crease above and rim below give the lids depth.
				if (!ch && s.crease && closure < 0.55 && Math.abs(u) < 0.9) {
					const crease = restTop - half * 0.071 * arc - cellH * 0.45;
					if (Math.abs(y - crease) < cellH * 0.37) {
						ch = Math.abs(u) < 0.35 ? '_' : u < 0 ? '/' : '\\';
						alpha = 0.24 * (1 - closure);
					}
				}
				if (
					!ch &&
					s.crease &&
					closure < 0.45 &&
					Math.abs(u) < 0.85 &&
					Math.abs(y - restBottom - half * 0.045 * arc - cellH * 0.4) < cellH * 0.32
				) {
					ch = '.';
					alpha = 0.22;
				}

				if (ch) this.glyph(ch, px, py, color, alpha * A);
			}
		}

		// Lashes ride on the upper lid, so they follow every blink.
		if (s.lashes) {
			for (let i = 0; i < s.lashes; i++) {
				const u = mix(-0.7, 0.7, s.lashes === 1 ? 0.5 : i / (s.lashes - 1));
				const x = u * half;
				const arc = Math.pow(Math.max(0, 1 - u * u), s.arcPow);
				const tilt = x * tiltSlope;
				const seam = half * seamK * arc + tilt;
				const restTop = mix(-half * s.topK * arc + tilt + iy * 0.18 * arc, seam, s.hood);
				const top = mix(restTop, seam, closure);
				const ch = u < -0.25 ? '\\' : u > 0.25 ? '/' : '|';
				for (let k = 1; k <= 2; k++) {
					const lx = cx + x + u * cellW * 0.9 * k;
					const ly = cy + top - cellH * k;
					const col = Math.round((lx - e.ox) / cellW);
					const row = Math.round((ly - e.oy) / cellH);
					this.glyph(ch, e.ox + col * cellW, e.oy + row * cellH, pal.lid, (k === 1 ? 0.6 : 0.35) * A);
				}
			}
		}

		// Slow dial ticks around the sigil's ring.
		if (s.ring === 2) {
			const ticks = 48;
			const tr = outerR + cellW * 2.4;
			for (let i = 0; i < ticks; i++) {
				const a = (i / ticks) * TAU + this.time * 0.015;
				const tx = cx + Math.cos(a) * tr;
				const ty = cy + Math.sin(a) * tr * ringStretch + Math.cos(a) * tr * tiltSlope;
				const major = i % 6 === 0;
				const col = Math.round((tx - e.ox) / cellW);
				const row = Math.round((ty - e.oy) / cellH);
				this.glyph(
					major ? '+' : '.',
					e.ox + col * cellW,
					e.oy + row * cellH,
					pal.lid,
					(major ? 0.32 : 0.16) * A
				);
			}
		}
	}

	private drawLine(e: LineEye) {
		const d = e.d;
		const st = e.st;
		this.fs = d.fs;
		const cw = d.fs * 0.6;
		const n = d.inner;
		const lead = d.open.length;
		const total = lead + n + d.close.length;
		const left = e.x - ((total - 1) * cw) / 2;
		const A = e.alpha;
		const pal = e.pal;
		for (let i = 0; i < lead; i++) this.glyph(d.open[i], left + i * cw, e.y, pal.lid, 0.85 * A);
		for (let i = 0; i < d.close.length; i++)
			this.glyph(d.close[i], left + (lead + n + i) * cw, e.y, pal.lid, 0.85 * A);

		if (st.closure > 0.5) {
			for (let i = 0; i < n; i++) this.glyph('-', left + (lead + i) * cw, e.y, pal.lid, 0.8 * A);
			return;
		}
		const travel = n === 1 ? cw * 0.32 : ((n - 1) * cw) / 2;
		const px = left + cw * (lead + (n - 1) / 2) + st.gx * travel;
		const py = e.y + st.gy * d.fs * 0.16;
		if (d.fill) {
			const at = Math.round((px - left) / cw) - lead;
			for (let i = 0; i < n; i++)
				if (i !== at) this.glyph(d.fill, left + (lead + i) * cw, e.y, pal.lid, 0.5 * A);
		}
		this.glyph(st.dil > 1.2 && d.wide ? d.wide : d.pupil, px, py, pal.iris, A);
	}

	private drawArt(e: ArtEye) {
		const p = e.p;
		const st = e.st;
		this.fs = p.fs;
		const cw = p.fs * 0.6;
		const ch = p.fs * p.lh;
		const left = e.x - ((p.cols - 1) * cw) / 2;
		const top = e.y - ((p.rows - 1) * ch) / 2;
		const A = e.alpha;
		const pal = e.pal;
		const closed = st.closure > 0.5;
		const frame = closed ? p.closed : p.open;

		for (let r = 0; r < p.rows; r++) {
			const row = frame[r];
			for (let c = 0; c < p.cols; c++) {
				const g = row[c];
				if (g === ' ') continue;
				this.glyph(g, left + c * cw, top + r * ch, pal.lid, (SOFT.has(g) ? 0.72 : 0.95) * A);
			}
		}
		if (closed) return;

		const sprite = st.dil > 1.2 ? p.wide : p.iris;
		const sh = sprite.length;
		const sw = Math.max(...sprite.map((s) => s.length));
		const dc = Math.round(st.gx * p.range[0]);
		const dr = Math.round(st.gy * p.range[1]);
		const covered = p.fill ? new Set<number>() : null;

		p.centers.forEach(([cc, cr], k) => {
			const c0 = cc + dc - Math.floor(sw / 2);
			const r0 = cr + dr - Math.floor(sh / 2);
			for (let sr = 0; sr < sh; sr++) {
				for (let sc = 0; sc < sw; sc++) {
					const g = sprite[sr][sc];
					if (!g || g === ' ') continue;
					const R = r0 + sr;
					const C = c0 + sc;
					if (p.comp[R]?.[C] !== k) continue;
					covered?.add(R * 1000 + C);
					const strong = p.pupils.includes(g);
					this.glyph(g, left + C * cw, top + R * ch, pal.iris, (strong ? 1 : 0.55) * A);
				}
			}
		});

		if (p.fill && covered) {
			for (let r = 0; r < p.rows; r++) {
				for (let c = 0; c < p.cols; c++) {
					if (p.comp[r][c] < 0 || covered.has(r * 1000 + c) || hash(c, r) > 0.55) continue;
					this.glyph(p.fill, left + c * cw, top + r * ch, pal.sclera, 0.22 * A);
				}
			}
		}
	}

	private drawPing(p: Ping) {
		const ctx = this.ctx;
		const age = this.time - p.t;
		this.fs = 11;
		const cw = 11 * 0.6;
		const chh = 11 * 0.98;

		// The visible front of the wave that the eyes respond to.
		if (age < 0.6) {
			const R = age * WAVE_SPEED;
			const fade = (1 - age / 0.6) ** 2;
			const steps = Math.min(420, Math.max(12, Math.ceil((TAU * R) / (cw * 1.7))));
			for (let i = 0; i < steps; i++) {
				const a = (i / steps) * TAU;
				const c = Math.cos(a);
				const s = Math.sin(a);
				const gx = Math.round((p.x + c * R) / cw) * cw;
				const gy = Math.round((p.y + s * R) / chh) * chh;
				const g = Math.abs(s) > 0.86 ? '-' : Math.abs(c) > 0.9 ? '|' : s * c > 0 ? '/' : '\\';
				this.glyph(g, gx, gy, PURPLE, fade * (0.3 + 0.35 * hash(i, p.t)));
			}
		}

		// A small reticle that tightens on the point.
		const life = age < 0.9 ? 1 : 1 - smooth((age - 0.9) / 0.5);
		if (life > 0) {
			const d = mix(34, 13, smooth(age / 0.22));
			this.glyph('[', p.x - d, p.y, PURPLE, 0.9 * life);
			this.glyph(']', p.x + d, p.y, PURPLE, 0.9 * life);
			this.glyph('+', p.x, p.y, YELLOW, life);
			ctx.font = `400 10px ${FONT}`;
			ctx.textAlign = 'left';
			ctx.fillStyle = BONE;
			ctx.globalAlpha = 0.38 * life;
			ctx.fillText(`${Math.round(p.x)},${Math.round(p.y)}`, p.x + d + 8, p.y + 13);
			ctx.textAlign = 'center';
		}
	}
}
