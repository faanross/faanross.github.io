/**
 * The homepage eye field: one hero eye plus a scattered cast, laid out around
 * the page text. Which designs appear is set by HOME_CAST (ids from designs.ts).
 */

import {
	EyeScene,
	PALETTES,
	clamp,
	makeDrift,
	mixedPalette,
	placeGroup,
	rand,
	type Group,
	type Palette,
	type PaletteName,
	type ProcDrawing,
	type Rect
} from './engine';
import { byId, spawn, type Design } from './designs';

/**
 * Cast entries are design ids from designs.ts. Append `:palette` to pin a colour
 * pairing (e.g. 'C02:aion'); otherwise each eye gets a random mix.
 */
export interface Cast {
	/** The large eye. Must be a procedural (group A) design; it is resized to fit. */
	hero: string;
	/** Medium eyes, each placed at most once. */
	mids: string[];
	/** Small eyes, sprinkled to fill the space (repeats allowed). */
	smalls: string[];
}

export const HOME_CAST: Cast = {
	hero: 'A01',
	mids: ['A03', 'A05:sun', 'B08', 'B02', 'B04', 'C03', 'C06:gold', 'C02:aion', 'C19'],
	smalls: ['D05', 'D08', 'E02', 'E05', 'E14', 'E18']
};

export interface EyeFieldOptions {
	avoidSelector?: string;
	cast?: Cast;
}

export interface EyeField {
	poke(x: number, y: number): void;
	/** Fade the eyes in: the big one first, the rest at short random intervals. */
	reveal(): void;
	destroy(): void;
}

interface Slot {
	x: number;
	y: number;
	rx: number;
	ry: number;
}

interface Entry {
	design: Design;
	pal?: Palette;
}

interface Kept {
	g: Group;
	nx: number;
	ny: number;
	gap: number;
}

function parseEntry(s: string): Entry | null {
	const [id, pal] = s.split(':');
	const design = byId(id);
	if (!design) return null;
	return { design, pal: pal && pal in PALETTES ? PALETTES[pal as PaletteName] : undefined };
}

const shuffle = <T>(list: T[]): T[] => list.sort(() => Math.random() - 0.5);

export function createEyeField(canvas: HTMLCanvasElement, options: EyeFieldOptions = {}): EyeField {
	const scene = new EyeScene(canvas);
	const avoidSelector = options.avoidSelector ?? '[data-eye-avoid]';
	const cast = options.cast ?? HOME_CAST;
	/** The scattered eyes and where they sit, as fractions of the window. */
	let others: Kept[] = [];
	let hero: Group | null = null;
	let revealed = false;

	function textRects(): Rect[] {
		return Array.from(document.querySelectorAll(avoidSelector)).map((el) => {
			const r = el.getBoundingClientRect();
			return { left: r.left, top: r.top, right: r.right, bottom: r.bottom };
		});
	}

	/** The big eye keeps clear of the text; the rest may drift behind it, dimmed. */
	function avoidRects(): Rect[] {
		const pad = 28;
		return textRects().map((r) => ({
			left: r.left - pad,
			top: r.top - pad,
			right: r.right + pad,
			bottom: r.bottom + pad
		}));
	}

	/**
	 * The big eye always sits in the same spot for a given window size (right of
	 * the text on desktop, top centre on phones). It shrinks to clear the text
	 * but never moves at random.
	 */
	function placeHero(avoid: Rect[], placed: Slot[], out: Group[]): Group | null {
		const entry = parseEntry(cast.hero);
		if (!entry || entry.design.draw.kind !== 'proc') return null;
		const W = scene.width;
		const H = scene.height;
		const mobile = W < 640;
		const base = entry.design.draw.spec;
		const ideal = mobile ? clamp(W * 0.34, 96, 170) : clamp(Math.min(W, H) * 0.27, 140, 290);
		const margin = 24;
		let chosen: { g: Group; slot: Slot } | null = null;
		for (const shrink of [1, 0.9, 0.8, 0.7, 0.6, 0.5]) {
			const half = ideal * shrink;
			const drawing: ProcDrawing = {
				kind: 'proc',
				spec: { ...base, half, fs: mobile || half < 150 ? 9 : half > 210 ? 11 : 10 }
			};
			const g = spawn(entry.design, scene.time, { drawing, pal: entry.pal, drift: makeDrift(4) });
			const x = mobile ? W / 2 : clamp(W * 0.7, g.rx + margin, W - g.rx - margin);
			const y = mobile ? g.ry + Math.max(margin, H * 0.05) : clamp(H * 0.42, g.ry + margin, H - g.ry - margin);
			chosen = { g, slot: { x, y, rx: g.rx, ry: g.ry } };
			if (fits(x, y, g.rx, g.ry, [], avoid, 0)) break;
		}
		if (!chosen) return null;
		placeGroup(chosen.g, chosen.slot.x, chosen.slot.y);
		placed.push(chosen.slot);
		out.push(chosen.g);
		return chosen.g;
	}

	function fits(x: number, y: number, rx: number, ry: number, placed: Slot[], avoid: Rect[], gap: number) {
		const W = scene.width;
		const H = scene.height;
		if (x - rx < 8 || x + rx > W - 8 || y - ry < 8 || y + ry > H - 8) return false;
		for (const p of placed) {
			if (Math.max(Math.abs(x - p.x) - (rx + p.rx), Math.abs(y - p.y) - (ry + p.ry)) < gap) return false;
		}
		for (const r of avoid) {
			if (Math.max(r.left - (x + rx), x - rx - r.right, r.top - (y + ry), y - ry - r.bottom) < 0) return false;
		}
		return true;
	}

	/**
	 * Stratified scatter: the page is cut into a coarse grid and each eye goes to
	 * a random spot inside one of the least-occupied cells. Eyes spread over the
	 * whole free area without lining up along edges or bunching in one gap.
	 */
	function scatter(items: Group[], placed: Slot[], avoid: Rect[], gap: number, out: Group[]) {
		const W = scene.width;
		const H = scene.height;
		const cols = Math.max(2, Math.round(W / 230));
		const rows = Math.max(2, Math.round(H / 190));
		const cw = W / cols;
		const ch = H / rows;
		const load = new Float32Array(cols * rows);
		const bump = (x0: number, y0: number, x1: number, y1: number) => {
			for (let r = Math.max(0, Math.floor(y0 / ch)); r <= Math.min(rows - 1, Math.floor(y1 / ch)); r++)
				for (let c = Math.max(0, Math.floor(x0 / cw)); c <= Math.min(cols - 1, Math.floor(x1 / cw)); c++)
					load[r * cols + c] += 1;
		};
		for (const r of avoid) bump(r.left, r.top, r.right, r.bottom);
		for (const p of placed) bump(p.x - p.rx, p.y - p.ry, p.x + p.rx, p.y + p.ry);

		for (const g of items) {
			const order = Array.from(load, (l, i) => [i, l + Math.random() * 0.9] as const).sort((a, b) => a[1] - b[1]);
			let slot: Slot | null = null;
			// Prefer generous spacing, but tighten it rather than drop an eye.
			for (const spacing of [gap, gap * 0.5, 4]) {
				for (const [i] of order) {
					const c = i % cols;
					const r = Math.floor(i / cols);
					for (let k = 0; k < 14 && !slot; k++) {
						const x = (c + Math.random()) * cw;
						const y = (r + Math.random()) * ch;
						if (fits(x, y, g.rx, g.ry, placed, avoid, spacing)) slot = { x, y, rx: g.rx, ry: g.ry };
					}
					if (slot) break;
				}
				if (slot) break;
			}
			if (!slot) continue;
			placeGroup(g, slot.x, slot.y);
			placed.push(slot);
			out.push(g);
			bump(slot.x - slot.rx, slot.y - slot.ry, slot.x + slot.rx, slot.y + slot.ry);
		}
	}

	function layout() {
		const W = scene.width;
		const H = scene.height;
		const groups: Group[] = [];
		const placed: Slot[] = [];
		const avoid = avoidRects();
		const t = scene.time;
		scene.dimRects = textRects();

		hero = placeHero(avoid, placed, groups);

		// Every medium eye in the cast, exactly once. Large procedural ones are
		// capped so they stay "medium" on smaller screens. Biggest go first;
		// smaller ones fill the gaps.
		const capHalf = Math.max(70, Math.min(W, H) * 0.12);
		const mids = shuffle(cast.mids.map(parseEntry).filter((e): e is Entry => !!e))
			.map(({ design, pal }) => {
				const drawing: ProcDrawing | undefined =
					design.draw.kind === 'proc' && design.draw.spec.half > capHalf
						? { kind: 'proc', spec: { ...design.draw.spec, half: capHalf } }
						: undefined;
				return spawn(design, t, { drawing, pal: pal ?? mixedPalette() });
			})
			.sort((a, b) => b.rx * b.ry - a.rx * a.ry);
		// No text avoidance from here on: these may sit behind the text (dimmed).
		scatter(mids, placed, [], 36, groups);

		// Every small eye in the cast, exactly once (no repeats).
		const smallGroups = shuffle(cast.smalls.map(parseEntry).filter((e): e is Entry => !!e)).map(
			({ design, pal }) => spawn(design, t, { pal: pal ?? mixedPalette() })
		);
		scatter(smallGroups, placed, [], 30, groups);

		// Remember where everyone landed, relative to the window, for reflow().
		others = groups
			.filter((g) => g !== hero)
			.map((g) => ({ g, nx: g.x / W, ny: g.y / H, gap: mids.includes(g) ? 36 : 30 }));
		// Hidden until reveal() is called.
		if (!revealed) for (const g of groups) g.fadeStart = Infinity;
		scene.groups = groups;
		if (import.meta.env.DEV) canvas.dataset.eyes = `${groups.length}/${1 + mids.length + smallGroups.length}`;
	}

	function reveal() {
		if (revealed) return;
		if (!laid) {
			// Asked before the first layout (fonts still loading): do it right after.
			pendingReveal = true;
			return;
		}
		revealed = true;
		if (scene.reduced) {
			for (const g of scene.groups) g.fadeStart = -Infinity;
			scene.render();
			return;
		}
		// Big eye first, then the rest at short, uneven intervals.
		let t = scene.time;
		if (hero) hero.fadeStart = t;
		t += 0.55;
		for (const o of shuffle([...others])) {
			o.g.fadeStart = t;
			t += rand(0.07, 0.28);
		}
	}

	/**
	 * Window resized: keep the arrangement. Each eye keeps its relative position;
	 * only the ones that would now hit the text, the big eye or another eye are
	 * moved (or hidden until there is room again).
	 */
	function reflow() {
		const W = scene.width;
		const H = scene.height;
		const groups: Group[] = [];
		const placed: Slot[] = [];
		const avoid = avoidRects();
		scene.dimRects = textRects();
		hero = placeHero(avoid, placed, groups);
		if (hero && !revealed) hero.fadeStart = Infinity;
		const misfits: Kept[] = [];
		for (const o of others) {
			const x = o.nx * W;
			const y = o.ny * H;
			if (fits(x, y, o.g.rx, o.g.ry, placed, [], o.gap)) {
				placeGroup(o.g, x, y);
				placed.push({ x, y, rx: o.g.rx, ry: o.g.ry });
				groups.push(o.g);
			} else {
				misfits.push(o);
			}
		}
		scatter(
			misfits.map((o) => o.g),
			placed,
			[],
			30,
			groups
		);
		for (const o of misfits) {
			if (!groups.includes(o.g)) continue;
			o.nx = o.g.x / W;
			o.ny = o.g.y / H;
		}
		scene.groups = groups;
	}

	// First size: a fresh random arrangement. Every later resize (including a
	// phone toolbar sliding away) only reflows it.
	let laid = false;
	let pendingReveal = false;
	let lastW = 0;
	let lastH = 0;
	scene.onResize = () => {
		if (!laid) {
			laid = true;
			layout();
			if (pendingReveal) reveal();
		} else if (scene.width !== lastW || scene.height !== lastH) {
			reflow();
		}
		lastW = scene.width;
		lastH = scene.height;
	};

	scene.start();

	return {
		poke: (x, y) => scene.poke(x, y),
		reveal,
		destroy: () => scene.destroy()
	};
}
