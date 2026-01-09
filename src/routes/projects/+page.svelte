<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { projectCategories } from '$lib/data/projects';

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});
</script>

<svelte:head>
	<title>Projects | Faan Rossouw</title>
	<meta name="description" content="Open-source C2 tools and security research projects." />
</svelte:head>

<section class="projects-hero">
	<div class="container">
		{#if mounted}
			<h1 in:fly={{ y: 30, duration: 800, delay: 200 }}>Projects</h1>
			<p class="lead" in:fly={{ y: 20, duration: 600, delay: 400 }}>
				Open-source tools and research projects for the security community
			</p>
		{/if}
	</div>
</section>

{#each projectCategories as category, catIndex}
	<section class="projects-category">
		<div class="container">
			{#if mounted}
				<div class="category-header" in:fly={{ y: 30, duration: 600, delay: 500 + catIndex * 100 }}>
					<h2>{category.name}</h2>
					<p>{category.description}</p>
				</div>
				<div class="projects-grid">
					{#each category.projects as project, index}
						<a
							href={project.url}
							target="_blank"
							rel="noopener noreferrer"
							class="project-card glass-card"
							in:fly={{ y: 30, duration: 600, delay: 600 + catIndex * 100 + index * 80 }}
						>
							<div class="project-header">
								<svg class="github-icon" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
									<path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
								</svg>
								<h3>{project.name}</h3>
							</div>
							<p class="description">{project.description}</p>
							<div class="project-meta">
								<span class="language">
									<span class="language-dot"></span>
									{project.language}
								</span>
							</div>
							<div class="topics">
								{#each project.topics as topic}
									<span class="topic">{topic}</span>
								{/each}
							</div>
						</a>
					{/each}
				</div>
			{/if}
		</div>
	</section>
{/each}

<section class="github-cta">
	<div class="container">
		{#if mounted}
			<div class="cta-card glass-card" in:fly={{ y: 30, duration: 600, delay: 800 }}>
				<h2>More on GitHub</h2>
				<p>Check out all my repositories and contributions</p>
				<a href="https://github.com/faanross" target="_blank" rel="noopener noreferrer" class="btn-primary">
					View GitHub Profile
				</a>
			</div>
		{/if}
	</div>
</section>

<style>
	.projects-hero {
		padding: 80px 0 60px;
		text-align: center;
	}

	.projects-hero h1 {
		margin-bottom: 16px;
	}

	.lead {
		font-size: clamp(16px, 2vw, 20px);
		color: rgba(255, 255, 255, 0.7);
	}

	.projects-category {
		padding: 0 0 60px;
	}

	.category-header {
		margin-bottom: 32px;
	}

	.category-header h2 {
		font-size: clamp(24px, 3vw, 32px);
		margin-bottom: 8px;
		background: linear-gradient(135deg, var(--aion-purple), var(--aion-purple-light));
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.category-header p {
		font-size: 15px;
		color: rgba(255, 255, 255, 0.6);
	}

	.projects-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
		gap: 24px;
	}

	.project-card {
		display: flex;
		flex-direction: column;
		padding: 28px;
		text-decoration: none;
		transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.project-card:hover {
		transform: translateY(-4px);
		border-color: rgba(189, 147, 249, 0.5);
		box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4);
	}

	.project-header {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 16px;
	}

	.github-icon {
		color: var(--aion-purple);
		flex-shrink: 0;
	}

	.project-card:hover .github-icon {
		color: var(--aion-purple-light);
	}

	.project-header h3 {
		font-size: 20px;
		font-weight: 600;
		color: var(--white);
	}

	.description {
		font-size: 15px;
		color: rgba(255, 255, 255, 0.7);
		line-height: 1.6;
		flex-grow: 1;
		margin-bottom: 20px;
	}

	.project-meta {
		display: flex;
		align-items: center;
		gap: 20px;
		margin-bottom: 16px;
	}

	.language {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		color: rgba(255, 255, 255, 0.7);
	}

	.language-dot {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		background: #00ADD8; /* Go blue */
	}

	.topics {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.topic {
		font-size: 11px;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 4px 10px;
		background: rgba(189, 147, 249, 0.15);
		border: 1px solid rgba(189, 147, 249, 0.3);
		border-radius: 12px;
		color: var(--aion-purple-light);
	}

	.github-cta {
		padding: 0 0 100px;
	}

	.cta-card {
		max-width: 500px;
		margin: 0 auto;
		text-align: center;
		padding: 40px;
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
		.projects-grid {
			grid-template-columns: 1fr;
		}

		.project-card {
			padding: 24px;
		}

		.cta-card {
			padding: 32px 24px;
		}
	}
</style>
