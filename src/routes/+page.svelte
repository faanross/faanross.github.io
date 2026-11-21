<script lang="ts">
	import { onMount } from 'svelte';
	import EyeField from '$lib/components/EyeField.svelte';
	import { ICONS } from '$lib/icons';

	// Two lines from The Matrix. The first is typed in yellow and vanishes; the
	// second is typed in purple and shrinks into place as the tagline.
	const FIRST = "what's he doing?";
	const motto = "he's beginning to believe.";

	// Small brand icons at text size; the name stays as tooltip and screen-reader label.
	const links = [
		{ icon: ICONS.github, label: 'github', href: 'https://github.com/faanross' },
		{ icon: ICONS.youtube, label: 'youtube', href: 'https://www.youtube.com/@aionsec' },
		{ icon: ICONS.linkedin, label: 'linkedin', href: 'https://www.linkedin.com/in/faan-rossouw' },
		{ icon: ICONS.x, label: 'x', href: 'https://x.com/faanross' },
		{ icon: ICONS.email, label: 'email', href: 'mailto:moi@faanross.com' },
		{ icon: ICONS.discord, label: 'discord', href: 'https://discord.com/invite/Bzpk8yzZcf' }
	];

	let phase = $state<'intro' | 'done'>('intro');
	let typed = $state('');
	let lineLen = $state(FIRST.length);
	let lineColor = $state('var(--yellow)');
	let typing = $state(false);
	let eyesOn = $state(false);
	let hydrated = $state(false);

	let line = $state<HTMLElement>();
	let bigText = $state<HTMLElement>();
	let mottoText = $state<HTMLElement>();

	onMount(() => {
		hydrated = true;
		const timers: ReturnType<typeof setTimeout>[] = [];
		let over = false;
		const wait = (ms: number) => new Promise<void>((r) => timers.push(setTimeout(r, ms)));

		function finish() {
			if (over) return;
			over = true;
			timers.forEach(clearTimeout);
			typed = motto;
			phase = 'done';
			// Text first, line by line (about 1.9s), then the eyes open.
			timers.push(setTimeout(() => (eyesOn = true), 1900));
		}

		async function type(text: string) {
			lineLen = text.length;
			typing = true;
			for (let i = 1; i <= text.length && !over; i++) {
				typed = text.slice(0, i);
				await wait(text[i - 1] === ' ' ? 95 : 45 + Math.random() * 70);
			}
			typing = false;
		}

		async function run() {
			await wait(700);
			await type(FIRST);
			await wait(1200);
			if (over) return;
			// The yellow line vanishes, caret and all, and the purple line starts
			// typing in the same frame: no lone caret in between.
			lineColor = 'var(--purple)';
			await type(motto);
			await wait(1000);
			if (over || !line || !bigText || !mottoText) return finish();

			// Shrink the big line onto the tagline's exact spot, then hand over.
			const from = bigText.getBoundingClientRect();
			const to = mottoText.getBoundingClientRect();
			const s = to.width / from.width;
			const flight = line.animate(
				[
					{ transform: 'none', opacity: 1 },
					{ transform: `translate(${to.left - from.left}px, ${to.top - from.top}px) scale(${s})`, opacity: 0.7 }
				],
				{ duration: 1050, easing: 'cubic-bezier(0.65, 0, 0.25, 1)', fill: 'forwards' }
			);
			await flight.finished.catch(() => {});
			finish();
		}

		if (matchMedia('(prefers-reduced-motion: reduce)').matches) finish();
		else run();

		// Any click or key press skips straight to the end.
		const skip = () => finish();
		window.addEventListener('pointerdown', skip);
		window.addEventListener('keydown', skip);
		return () => {
			over = true;
			timers.forEach(clearTimeout);
			window.removeEventListener('pointerdown', skip);
			window.removeEventListener('keydown', skip);
		};
	});
</script>

<svelte:head>
	<title>Faan Rossouw</title>
	<meta
		name="description"
		content="Faan Rossouw, founder of aionsec.ai, where he teaches defenders to master agents."
	/>
	<meta property="og:title" content="Faan Rossouw" />
	<meta property="og:description" content="Founder of aionsec.ai, teaching defenders to master agents." />
	<meta property="og:url" content="https://www.faanross.com/" />
	<noscript>
		<style>
			.home .motto,
			.home .step {
				opacity: 1 !important;
				transform: none !important;
			}
			.stage {
				display: none !important;
			}
		</style>
	</noscript>
</svelte:head>

<EyeField visible={eyesOn} />

{#if phase === 'intro'}
	<div class="stage" aria-hidden="true">
		<div class="line" bind:this={line} style:color={lineColor} style="width: calc({lineLen}ch + 0.9em)">
			<span class="txt" bind:this={bigText}>{typed}</span><span class="caret" class:solid={typing}></span>
		</div>
	</div>
{/if}

<main class="home" class:revealed={phase === 'done'}>
	<div class="intro" data-eye-avoid inert={hydrated && phase === 'intro'}>
		<p class="motto">
			<span class="txt" bind:this={mottoText}>{motto}</span><span class="caret" aria-hidden="true"></span>
		</p>

		<div class="bio">
			<h1 class="step" style="--i: 0">Hi, I'm Faan.</h1>
			<p class="step" style="--i: 1">
				Founder of <a href="https://aionsec.ai" class="aion">aionsec.ai</a>, where I teach defenders to master
				agents.
			</p>

			<p class="links step" style="--i: 2">
				<span class="frame" aria-hidden="true">{'[ '}</span>{#each links as l, i}{#if i > 0}<span
							class="frame"
							aria-hidden="true">{' | '}</span
						>{/if}<a
						href={l.href}
						target={l.href.startsWith('mailto:') ? undefined : '_blank'}
						rel={l.href.startsWith('mailto:') ? undefined : 'noopener noreferrer'}
						aria-label={l.label}
						title={l.label}
						class="icon-link"
						><svg viewBox="0 0 24 24" class="icon" class:stroke={l.icon.stroke} aria-hidden="true"
							>{#each l.icon.rects ?? [] as r}<rect {...r} />{/each}{#each l.icon.paths as d}<path
									{d}
								/>{/each}</svg
						></a
					>{/each}<span class="frame" aria-hidden="true">{' ]'}</span>
			</p>
		</div>

		<p class="note step" style="--i: 3">
			Taking my C2 in Go course from AntiSyphon? The <a href="/courses/course01">course archive</a> is here.
		</p>
	</div>
</main>

<style>
	/* ---------------- intro overlay ---------------- */

	.stage {
		position: fixed;
		inset: 0;
		z-index: 2;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0 24px;
		pointer-events: none;
	}

	/* Fixed width per line, text left-aligned: it types rightward and ends centred. */
	.line {
		display: flex;
		align-items: center;
		font-size: clamp(15px, 1.7vw, 22px);
		line-height: 1.2;
		white-space: nowrap;
		transform-origin: 0 0;
	}

	.caret {
		display: inline-block;
		flex: none;
		width: 0.6em;
		height: 1.05em;
		margin-left: 0.25em;
		background: currentColor;
		animation: blink 1.05s steps(1) infinite;
	}

	.caret.solid {
		animation: none;
	}

	/* ---------------- page ---------------- */

	.home {
		position: relative;
		z-index: 1;
		min-height: 100vh;
		min-height: 100dvh;
		display: flex;
		flex-direction: column;
		justify-content: flex-end;
		padding: 7rem clamp(24px, 11vw, 220px) clamp(48px, 12vh, 128px);
		pointer-events: none;
	}

	.intro {
		/* Wide enough for the bio sentence (64 characters) on one line. */
		max-width: 38rem;
		pointer-events: auto;
	}

	.motto {
		display: flex;
		align-items: center;
		margin-bottom: 2.4rem;
		font-size: 15px; /* same as the bio */
		line-height: 1.2;
		color: var(--purple);
		white-space: nowrap;
	}

	/*
	 * Hidden until the intro lands, then revealed line by line (--i is the
	 * line's position). Opacity only: a link can't escape its parent's opacity,
	 * whereas with `visibility` some browsers showed "course archive" early.
	 * The block is `inert` during the intro, so nothing hidden is clickable.
	 */
	.motto {
		opacity: 0;
	}

	.revealed .motto {
		opacity: 0.7;
	}

	.step {
		opacity: 0;
		transform: translateY(6px);
	}

	.revealed .step {
		opacity: 1;
		transform: none;
		transition:
			opacity 0.7s ease,
			transform 0.7s ease;
		transition-delay: calc(0.15s + var(--i) * 0.35s);
	}

	.bio {
		font-size: 15px;
		line-height: max(24px, 2.9vh);
	}

	.bio h1 {
		font-size: inherit;
		font-weight: 400;
		line-height: inherit;
		color: var(--fg);
		/* One blank line before the bio. */
		margin-bottom: 1.6em;
		margin-bottom: 1lh;
	}

	.bio p {
		color: var(--fg-dim);
	}

	.aion {
		color: var(--yellow);
		text-decoration-color: rgba(241, 250, 140, 0.4);
	}

	.links {
		margin-top: 1.4rem;
	}

	.links a {
		color: var(--fg-dim);
		text-decoration-color: var(--line);
	}

	/* Same yellow as the aionsec.ai link. */
	.links a:hover,
	.links a:focus-visible {
		color: var(--yellow);
	}

	.frame {
		color: var(--fg-faint);
		white-space: pre-wrap;
	}

	/* Icons sit in the line like characters, in the text colour. */
	.icon-link {
		display: inline-block;
		padding: 0 0.15em;
		text-decoration: none;
		transition: color 0.15s ease;
	}

	.icon {
		width: 1em;
		height: 1em;
		vertical-align: -0.14em;
		fill: currentColor;
	}

	/* Outline icons (the envelope): sharp joins to match the lo-fi set. */
	.icon.stroke {
		fill: none;
		stroke: currentColor;
		stroke-width: 2;
		stroke-linecap: butt;
		stroke-linejoin: miter;
	}

	.note {
		margin-top: 2.6rem;
		font-size: 12px;
		line-height: 1.6;
		color: var(--fg-faint);
	}

	.note a {
		color: var(--fg-dim);
	}

	.note a:hover {
		color: var(--fg);
	}

	@keyframes blink {
		50% {
			opacity: 0;
		}
	}

	@media (max-width: 640px) {
		.home {
			padding: 5rem 24px 40px;
		}

		.motto {
			margin-bottom: 1.6rem;
			font-size: 14px;
		}

		.bio {
			font-size: 14px;
			line-height: 1.7;
		}

		.note {
			margin-top: 2rem;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.caret {
			animation: none;
		}
	}
</style>
