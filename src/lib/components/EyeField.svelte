<script lang="ts">
	import { onMount } from 'svelte';
	import { createEyeField, type EyeField } from '$lib/eyes/field';

	interface Props {
		/** Text blocks: the big eye keeps clear of them, other eyes dim behind them. */
		avoid?: string;
		/** Opacity of the whole eye layer, so the page text stays on top. */
		opacity?: number;
		/** When this turns true the eyes fade in, big one first, the rest staggered. */
		visible?: boolean;
	}

	let { avoid = '[data-eye-avoid]', opacity = 0.65, visible = true }: Props = $props();

	let canvas: HTMLCanvasElement;
	let field: EyeField | null = null;

	$effect(() => {
		if (visible) field?.reveal();
	});

	onMount(() => {
		const f = createEyeField(canvas, { avoidSelector: avoid });
		field = f;
		if (visible) f.reveal();

		// Any press that isn't on a link or button draws every eye's attention.
		function onPointerDown(ev: PointerEvent) {
			if (ev.button !== 0) return;
			const target = ev.target as Element | null;
			if (target?.closest('a, button, input, textarea, select, label')) return;
			f.poke(ev.clientX, ev.clientY);
		}

		window.addEventListener('pointerdown', onPointerDown);
		return () => {
			window.removeEventListener('pointerdown', onPointerDown);
			f.destroy();
			field = null;
		};
	});
</script>

<canvas bind:this={canvas} class="eye-field" style:opacity aria-hidden="true"></canvas>

<style>
	.eye-field {
		position: fixed;
		inset: 0;
		width: 100%;
		height: 100%;
		pointer-events: none;
		z-index: 0;
	}
</style>
