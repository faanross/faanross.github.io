<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { talkSeries } from '$lib/data/talks';

	let mounted = $state(false);
	let activeVideo = $state<string | null>(null);

	onMount(() => {
		mounted = true;
	});

	function openVideo(videoId: string) {
		activeVideo = videoId;
	}

	function closeVideo() {
		activeVideo = null;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && activeVideo) {
			closeVideo();
		}
	}
</script>

<svelte:head>
	<title>Talks | Faan Rossouw</title>
	<meta name="description" content="Webcasts and podcast appearances on C2, threat detection, and security research." />
</svelte:head>

<svelte:window onkeydown={handleKeydown} />

<section class="talks-hero">
	<div class="container">
		{#if mounted}
			<h1 in:fly={{ y: 30, duration: 800, delay: 200 }}>Talks</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				Webcasts and podcast appearances on C2, threat detection, and security research
			</p>
		{/if}
	</div>
</section>

{#each talkSeries as series, seriesIndex}
	<section class="talk-series">
		<div class="container">
			{#if mounted}
				<div class="series-header" in:fly={{ y: 30, duration: 600, delay: 500 + seriesIndex * 100 }}>
					<h2>{series.name}</h2>
					<p>{series.description}</p>
				</div>
				<div class="videos-grid">
					{#each series.talks as talk, talkIndex}
						<button
							class="video-card glass-card"
							onclick={() => openVideo(talk.videoId)}
							in:fly={{ y: 30, duration: 500, delay: 600 + seriesIndex * 100 + talkIndex * 80 }}
						>
							<div
								class="thumbnail"
								style="background-image: url('https://img.youtube.com/vi/{talk.videoId}/maxresdefault.jpg')"
							>
								<div class="play-button">
									<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="currentColor">
										<polygon points="5 3 19 12 5 21 5 3"></polygon>
									</svg>
								</div>
							</div>
							<h3>{talk.title}</h3>
							{#if talk.description}
								<p>{talk.description}</p>
							{/if}
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</section>
{/each}

<section class="aionsec-cta">
	<div class="container">
		{#if mounted}
			<div class="cta-card glass-card" in:fly={{ y: 30, duration: 600, delay: 800 }}>
				<h2>Go Deeper</h2>
				<p>Enjoyed the talks? AionSec offers hands-on courses that take these concepts from presentation to production</p>
				<a href="https://aionsec.ai/courses" target="_blank" rel="noopener noreferrer" class="btn-primary">
					Explore Courses
				</a>
			</div>
		{/if}
	</div>
</section>

{#if activeVideo}
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div
		class="video-modal"
		onclick={closeVideo}
		onkeydown={(e) => e.key === 'Escape' && closeVideo()}
		in:fade={{ duration: 200 }}
		out:fade={{ duration: 200 }}
		role="dialog"
		aria-modal="true"
		tabindex="-1"
	>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div class="modal-content" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()}>
			<button class="close-btn" onclick={closeVideo} aria-label="Close video">
				<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<line x1="18" y1="6" x2="6" y2="18"></line>
					<line x1="6" y1="6" x2="18" y2="18"></line>
				</svg>
			</button>
			<iframe
				src="https://www.youtube.com/embed/{activeVideo}?autoplay=1"
				title="Video player"
				frameborder="0"
				allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
				allowfullscreen
			></iframe>
		</div>
	</div>
{/if}

<style>
	.talks-hero {
		padding: 80px 0 60px;
		text-align: center;
	}

	.talks-hero h1 {
		margin-bottom: 16px;
	}

	.lead {
		font-size: clamp(16px, 2vw, 20px);
		color: rgba(255, 255, 255, 0.7);
	}

	.talk-series {
		padding: 40px 0 60px;
	}

	.series-header {
		margin-bottom: 32px;
	}

	.series-header h2 {
		font-size: clamp(24px, 3vw, 32px);
		margin-bottom: 12px;
		background: linear-gradient(135deg, var(--aion-purple), var(--aion-purple-light));
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.series-header p {
		font-size: 15px;
		color: rgba(255, 255, 255, 0.6);
	}

	.videos-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 24px;
	}

	.video-card {
		background: rgba(61, 61, 71, 0.4);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border: 1px solid rgba(189, 147, 249, 0.2);
		border-radius: 14px;
		padding: 0;
		overflow: hidden;
		cursor: pointer;
		text-align: left;
		transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.video-card:hover {
		transform: translateY(-4px);
		border-color: rgba(189, 147, 249, 0.5);
		box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4);
	}

	.thumbnail {
		aspect-ratio: 16/9;
		background-size: cover;
		background-position: center;
		background-color: rgba(0, 0, 0, 0.3);
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.play-button {
		width: 64px;
		height: 64px;
		background: rgba(245, 230, 99, 0.95);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--black);
		transition: transform 0.3s ease, background 0.3s ease;
	}

	.play-button svg {
		margin-left: 4px;
	}

	.video-card:hover .play-button {
		transform: scale(1.1);
		background: var(--aion-yellow-light);
	}

	.video-card h3 {
		font-size: 16px;
		font-weight: 600;
		color: var(--white);
		padding: 16px 20px 8px;
	}

	.video-card p {
		font-size: 13px;
		color: rgba(255, 255, 255, 0.6);
		padding: 0 20px 16px;
		line-height: 1.5;
	}

	.video-modal {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.9);
		z-index: 1000;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 24px;
	}

	.modal-content {
		position: relative;
		width: 100%;
		max-width: 1000px;
		aspect-ratio: 16/9;
	}

	.modal-content iframe {
		width: 100%;
		height: 100%;
		border-radius: 8px;
	}

	.close-btn {
		position: absolute;
		top: -48px;
		right: 0;
		background: transparent;
		border: none;
		color: var(--white);
		cursor: pointer;
		padding: 8px;
		transition: color 0.3s ease;
	}

	.close-btn:hover {
		color: var(--aion-yellow);
	}

	.aionsec-cta {
		padding: 20px 0 100px;
	}

	.cta-card {
		max-width: 500px;
		margin: 0 auto;
		text-align: center;
		padding: 40px;
		animation: ctaGlow 5s ease-in-out infinite;
	}

	@keyframes ctaGlow {
		0%, 100% {
			box-shadow: 0 0 15px rgba(245, 230, 99, 0.1), 0 0 30px rgba(245, 230, 99, 0.05);
		}
		50% {
			box-shadow: 0 0 20px rgba(245, 230, 99, 0.18), 0 0 40px rgba(245, 230, 99, 0.08);
		}
	}

	.cta-card h2 {
		font-size: 24px;
		margin-bottom: 12px;
	}

	.cta-card p {
		font-size: 15px;
		color: rgba(255, 255, 255, 0.7);
		margin-bottom: 24px;
	}

	@media (max-width: 768px) {
		.videos-grid {
			grid-template-columns: 1fr;
		}

		.cta-card {
			padding: 32px 24px;
		}

		.modal-content {
			max-width: 100%;
		}

		.close-btn {
			top: -44px;
		}
	}
</style>
