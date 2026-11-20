<script lang="ts">
	import '$lib/styles/global.css';
	import { afterNavigate } from '$app/navigation';
	import { addCopyButtons } from '$lib/utils/addCopyButtons';

	let { children } = $props();

	// Runs on first mount and after every navigation; lesson code blocks get copy buttons.
	afterNavigate(() => {
		setTimeout(addCopyButtons, 100);
	});
</script>

<svelte:head>
	<link rel="preload" href="/fonts/GeistMono.woff2" as="font" type="font/woff2" crossorigin="anonymous" />
	<link rel="icon" href="/favicon.ico" sizes="48x48" />
	<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png" />
	<link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png" />
	<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
	<meta name="theme-color" content="#0e0e10" />
</svelte:head>

{@render children()}
