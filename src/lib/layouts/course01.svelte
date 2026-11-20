<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import ScrollProgress from '$lib/components/ScrollProgress.svelte';

	interface Props {
		title?: string;
		children?: Snippet;
	}

	let { title = '', children }: Props = $props();

	// "Lesson 5: DNS Server" -> kicker "lesson 05", heading "DNS Server".
	const parts = $derived.by(() => {
		const m = title.match(/^Lesson (\d+):\s*(.*)$/);
		return m
			? { kicker: `lesson ${m[1].padStart(2, '0')}`, heading: m[2] }
			: { kicker: '', heading: title };
	});

	const slug = $derived(page.url.pathname.replace(/\/$/, '').split('/').pop() ?? '');
</script>

<ScrollProgress contentSelector=".lesson-body" />

<svelte:head>
	<title>{title} · Building a C2 Framework in Go</title>
</svelte:head>

<main class="lesson">
	<nav class="crumbs" aria-label="Breadcrumb">
		<a href="/">faanross</a><span class="slash">/</span><a href="/courses/course01">c2-framework-go</a><span
			class="slash">/</span
		><span class="here">{slug}</span>
	</nav>

	<header class="head">
		<p class="kicker">{parts.kicker || 'course archive'}</p>
		<h1>{parts.heading}</h1>
	</header>

	<article class="lesson-body">
		{@render children?.()}
	</article>

	<footer class="foot">
		<a href="/courses/course01">↑ course archive</a>
		<span class="dot">·</span>
		<a href="/">faanross.com</a>
	</footer>
</main>

<style>
	.lesson {
		max-width: 48rem;
		margin: 0 auto;
		padding: 56px 24px 96px;
	}

	.crumbs {
		font-size: 12px;
		color: var(--fg-faint);
		margin-bottom: 64px;
		overflow-wrap: anywhere;
	}

	.crumbs a,
	.foot a {
		color: var(--fg-dim);
		text-decoration-color: var(--line);
	}

	.crumbs a:hover,
	.foot a:hover {
		color: var(--fg);
		text-decoration-color: var(--purple);
	}

	.slash {
		margin: 0 0.6ch;
		opacity: 0.7;
	}

	.here {
		color: var(--fg-dim);
	}

	.kicker {
		font-size: 12px;
		color: var(--purple);
		opacity: 0.85;
		margin-bottom: 12px;
	}

	.head h1 {
		font-size: clamp(22px, 3.2vw, 28px);
		font-weight: 400;
		line-height: 1.3;
		letter-spacing: -0.01em;
		margin-bottom: 48px;
	}

	.foot {
		margin-top: 72px;
		padding-top: 24px;
		border-top: 1px dashed var(--line);
		font-size: 12px;
	}

	.dot {
		margin: 0 0.8ch;
		color: var(--fg-faint);
	}

	/* ---------------- lesson prose ---------------- */

	.lesson-body {
		font-size: 14.5px;
		line-height: 1.85;
		color: rgba(236, 235, 230, 0.8);
	}

	.lesson-body :global(h2),
	.lesson-body :global(h3),
	.lesson-body :global(h4) {
		font-weight: 400;
		color: var(--fg);
		line-height: 1.4;
		scroll-margin-top: 24px;
	}

	.lesson-body :global(h2) {
		font-size: 18px;
		margin: 56px 0 16px;
	}

	.lesson-body :global(h3) {
		font-size: 15.5px;
		margin: 40px 0 12px;
	}

	.lesson-body :global(h4) {
		font-size: 14.5px;
		margin: 32px 0 10px;
		color: var(--fg-dim);
	}

	.lesson-body :global(h2)::before {
		content: '## ';
		color: var(--purple);
	}

	.lesson-body :global(h3)::before {
		content: '### ';
		color: var(--purple);
		opacity: 0.6;
	}

	.lesson-body :global(h4)::before {
		content: '#### ';
		color: var(--purple);
		opacity: 0.4;
	}

	.lesson-body :global(p) {
		margin-bottom: 18px;
	}

	.lesson-body :global(strong) {
		color: var(--fg);
		font-weight: 400;
		background: rgba(189, 147, 249, 0.12);
		padding: 0 3px;
		border-radius: 2px;
	}

	.lesson-body :global(em) {
		color: var(--fg);
	}

	.lesson-body :global(ul),
	.lesson-body :global(ol) {
		margin: 0 0 18px;
		padding-left: 2.2ch;
	}

	.lesson-body :global(li) {
		margin-bottom: 6px;
		padding-left: 0.6ch;
	}

	.lesson-body :global(li::marker) {
		color: var(--purple);
	}

	.lesson-body :global(a) {
		color: var(--fg);
		text-decoration-color: var(--purple-soft);
	}

	.lesson-body :global(a:hover) {
		color: var(--yellow);
		text-decoration-color: currentColor;
	}

	.lesson-body :global(:not(pre) > code) {
		font-family: var(--mono);
		font-size: 0.92em;
		color: var(--yellow);
		background: rgba(236, 235, 230, 0.05);
		padding: 1px 5px;
		border-radius: 3px;
		overflow-wrap: anywhere;
	}

	.lesson-body :global(pre) {
		background: var(--bg-raised);
		border: 1px solid var(--line);
		border-radius: 6px;
		padding: 18px 20px;
		margin: 22px 0;
		overflow-x: auto;
		font-size: 13px;
		line-height: 1.65;
		tab-size: 4;
	}

	.lesson-body :global(pre code) {
		font-family: var(--mono);
		color: rgba(236, 235, 230, 0.86);
	}

	/* Prism tokens from mdsvex, in the same purple/yellow as the eyes. */
	.lesson-body :global(.token.comment),
	.lesson-body :global(.token.prolog),
	.lesson-body :global(.token.doctype) {
		color: rgba(236, 235, 230, 0.38);
		font-style: italic;
	}

	.lesson-body :global(.token.keyword),
	.lesson-body :global(.token.builtin),
	.lesson-body :global(.token.important),
	.lesson-body :global(.token.boolean) {
		color: var(--purple);
	}

	.lesson-body :global(.token.string),
	.lesson-body :global(.token.char),
	.lesson-body :global(.token.number),
	.lesson-body :global(.token.regex) {
		color: var(--yellow);
	}

	.lesson-body :global(.token.function),
	.lesson-body :global(.token.class-name) {
		color: #f2efe8;
	}

	.lesson-body :global(.token.operator),
	.lesson-body :global(.token.punctuation) {
		color: rgba(236, 235, 230, 0.5);
	}

	.lesson-body :global(blockquote) {
		margin: 22px 0;
		padding: 2px 0 2px 18px;
		border-left: 2px solid var(--purple);
		color: var(--fg-dim);
	}

	.lesson-body :global(blockquote p:last-child) {
		margin-bottom: 0;
	}

	.lesson-body :global(img) {
		display: block;
		max-width: 100%;
		height: auto;
		margin: 24px 0;
		border: 1px solid var(--line);
		border-radius: 4px;
	}

	.lesson-body :global(table) {
		width: 100%;
		border-collapse: collapse;
		margin: 22px 0;
		font-size: 13px;
		display: block;
		overflow-x: auto;
	}

	.lesson-body :global(th),
	.lesson-body :global(td) {
		padding: 9px 12px;
		border-bottom: 1px solid var(--line);
		text-align: left;
		vertical-align: top;
	}

	.lesson-body :global(th) {
		font-weight: 400;
		color: var(--fg-faint);
		border-bottom-color: rgba(236, 235, 230, 0.18);
	}

	.lesson-body :global(hr) {
		border: none;
		border-top: 1px dashed var(--line);
		margin: 48px 0 0;
	}

	@media (max-width: 640px) {
		.lesson {
			padding: 32px 20px 72px;
		}

		.crumbs {
			margin-bottom: 44px;
		}

		.lesson-body {
			font-size: 14px;
		}

		.lesson-body :global(pre) {
			padding: 16px;
			margin-left: -4px;
			margin-right: -4px;
		}
	}
</style>
