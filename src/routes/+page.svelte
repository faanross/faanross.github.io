<script lang="ts">
	import { onMount } from 'svelte';
	import EyeField from '$lib/components/EyeField.svelte';

	// Two lines from The Matrix. The first is typed in yellow and vanishes; the
	// second is typed in purple and shrinks into place as the tagline.
	const FIRST = "what's happening?";
	const motto = "he's beginning to believe.";

	const links = [
		{ label: 'github', href: 'https://github.com/faanross' },
		{ label: 'youtube', href: 'https://www.youtube.com/@aionsec' },
		{ label: 'linkedin', href: 'https://www.linkedin.com/in/faan-rossouw' },
		{ label: 'x', href: 'https://x.com/faanross' },
		{ label: 'email', href: 'mailto:moi@faanross.com' },
		{ label: 'discord', href: 'https://discord.com/invite/Bzpk8yzZcf' }
	];

	let phase = $state<'intro' | 'done'>('intro');
	let typed = $state('');
	let lineLen = $state(FIRST.length);
	let lineColor = $state('var(--yellow)');
	let typing = $state(false);
	let eyesOn = $state(false);

	let line = $state<HTMLElement>();
	let bigText = $state<HTMLElement>();
	let mottoText = $state<HTMLElement>();

	onMount(() => {
		const timers: ReturnType<typeof setTimeout>[] = [];
		let over = false;
		const wait = (ms: number) => new Promise<void>((r) => timers.push(setTimeout(r, ms)));

		function finish() {
			if (over) return;
			over = true;
			timers.forEach(clearTimeout);
			typed = motto;
			phase = 'done';
			// Text first, then the eyes open.
			timers.push(setTimeout(() => (eyesOn = true), 1100));
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
		content="Faan Rossouw, founder of aionsec.ai, where he teaches defenders to build with agents."
	/>
	<meta property="og:title" content="Faan Rossouw" />
	<meta property="og:description" content="Founder of aionsec.ai, teaching defenders to build with agents." />
	<meta property="og:url" content="https://www.faanross.com/" />
	<noscript>
		<style>
			.home .motto,
			.home .bio,
			.home .note {
				opacity: 1 !important;
				visibility: visible !important;
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
	<div class="intro" data-eye-avoid>
		<p class="motto">
			<span class="txt" bind:this={mottoText}>{motto}</span><span class="caret" aria-hidden="true"></span>
		</p>

		<div class="bio">
			<h1>Hi, I'm Faan.</h1>
			<p>
				Founder of <a href="https://aionsec.ai" class="aion">aionsec.ai</a>, where I teach defenders to build
				with agents.
			</p>

			<p class="links">
				<span class="frame" aria-hidden="true">{'[ '}</span>{#each links as l, i}{#if i > 0}<span
							class="frame"
							aria-hidden="true">{' | '}</span
						>{/if}<a
						href={l.href}
						target={l.href.startsWith('mailto:') ? undefined : '_blank'}
						rel={l.href.startsWith('mailto:') ? undefined : 'noopener noreferrer'}>{l.label}</a
					>{/each}<span class="frame" aria-hidden="true">{' ]'}</span>
			</p>
		</div>

		<p class="note">
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
		max-width: 35rem;
		pointer-events: auto;
	}

	.motto {
		display: flex;
		align-items: center;
		margin-bottom: 2.4rem;
		font-size: 13px;
		line-height: 1.2;
		color: var(--purple);
		white-space: nowrap;
	}

	/* Hidden (and unclickable) until the intro lands. */
	.motto,
	.bio,
	.note {
		opacity: 0;
		visibility: hidden;
	}

	.bio,
	.note {
		transform: translateY(6px);
	}

	.revealed .motto {
		opacity: 0.7;
		visibility: visible;
	}

	.revealed .bio,
	.revealed .note {
		opacity: 1;
		visibility: visible;
		transform: none;
		transition:
			opacity 0.8s ease 0.15s,
			transform 0.8s ease 0.15s,
			visibility 0s linear 0.15s;
	}

	.revealed .note {
		transition-delay: 0.5s;
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

	.links a:hover,
	.links a:focus-visible {
		color: var(--fg);
		text-decoration-color: var(--purple);
	}

	.frame {
		color: var(--fg-faint);
		white-space: pre-wrap;
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
