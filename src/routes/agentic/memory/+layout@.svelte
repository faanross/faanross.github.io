<script lang="ts">
	import '$lib/styles/global.css';
	import Background from '$lib/components/Background.svelte';

	let { children } = $props();
</script>

<Background />

<main class="dashboard-main">
	{@render children()}
</main>

<style>
	.dashboard-main {
		position: relative;
		z-index: 10;
		min-height: 100vh;
		padding: 20px;
	}
</style>
