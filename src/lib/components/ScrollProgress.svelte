<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fade } from 'svelte/transition';

	interface Props {
		contentSelector?: string;
	}

	let { contentSelector = '.article-content' }: Props = $props();

	let sections: { id: string; top: number }[] = $state([]);
	let activeIndex = $state(0);
	let visible = $state(false);

	function updateSections() {
		const headings = document.querySelectorAll(`${contentSelector} h2`);
		sections = Array.from(headings).map((h, i) => ({
			id: `section-${i}`,
			top: (h as HTMLElement).offsetTop
		}));
	}

	function handleScroll() {
		const scrollY = window.scrollY + window.innerHeight / 3;

		// Show dots only after scrolling past the hero
		visible = window.scrollY > 200;

		// Find active section
		for (let i = sections.length - 1; i >= 0; i--) {
			if (scrollY >= sections[i].top) {
				activeIndex = i;
				return;
			}
		}
		activeIndex = 0;
	}

	function scrollToSection(index: number) {
		const headings = document.querySelectorAll(`${contentSelector} h2`);
		if (headings[index]) {
			headings[index].scrollIntoView({ behavior: 'smooth', block: 'start' });
		}
	}

	onMount(() => {
		// Delay to ensure content is rendered
		setTimeout(() => {
			updateSections();
			handleScroll();
		}, 500);

		window.addEventListener('scroll', handleScroll, { passive: true });
		window.addEventListener('resize', updateSections, { passive: true });
	});

	onDestroy(() => {
		if (typeof window !== 'undefined') {
			window.removeEventListener('scroll', handleScroll);
			window.removeEventListener('resize', updateSections);
		}
	});
</script>

{#if visible && sections.length > 0}
	<nav class="scroll-progress" in:fade={{ duration: 300 }}>
		{#each sections as section, i}
			<button
				class="dot"
				class:active={i === activeIndex}
				onclick={() => scrollToSection(i)}
				aria-label={`Go to section ${i + 1}`}
			></button>
		{/each}
	</nav>
{/if}

<style>
	.scroll-progress {
		position: fixed;
		right: 24px;
		top: 50%;
		transform: translateY(-50%);
		display: flex;
		flex-direction: column;
		gap: 12px;
		z-index: 100;
	}

	.dot {
		width: 6px;
		height: 6px;
		border-radius: 1px;
		background: rgba(236, 235, 230, 0.2);
		border: none;
		cursor: pointer;
		padding: 0;
		transition:
			background 0.2s ease,
			transform 0.2s ease;
	}

	.dot:hover {
		background: var(--purple);
		transform: scale(1.3);
	}

	.dot.active {
		background: var(--yellow);
	}

	@media (max-width: 1100px) {
		.scroll-progress {
			right: 12px;
		}
	}

	@media (max-width: 900px) {
		.scroll-progress {
			display: none;
		}
	}
</style>
