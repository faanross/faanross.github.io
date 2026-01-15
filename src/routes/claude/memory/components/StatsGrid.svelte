<script lang="ts">
	type StatItem = {
		label: string;
		value: string | number;
		subtext?: string;
	};

	let { stats = [], title = '' }: { stats: StatItem[]; title?: string } = $props();
</script>

<div class="stats-grid">
	{#if title}
		<h3>{title}</h3>
	{/if}

	<div class="grid">
		{#each stats as stat}
			<div class="stat-card">
				<span class="stat-value">{stat.value}</span>
				<span class="stat-label">{stat.label}</span>
				{#if stat.subtext}
					<span class="stat-subtext">{stat.subtext}</span>
				{/if}
			</div>
		{/each}
	</div>
</div>

<style>
	.stats-grid {
		margin-bottom: 40px;
	}

	h3 {
		margin-bottom: 20px;
		font-size: 16px;
		font-weight: 600;
		color: var(--aion-purple);
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
		gap: 16px;
	}

	.stat-card {
		background: rgba(61, 61, 71, 0.2);
		border-radius: 8px;
		padding: 20px;
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.stat-value {
		font-size: 28px;
		font-weight: 700;
		color: var(--aion-yellow);
		font-family: 'JetBrains Mono', monospace;
		line-height: 1;
		margin-bottom: 8px;
	}

	.stat-label {
		font-size: 11px;
		color: rgba(255, 255, 255, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.stat-subtext {
		font-size: 10px;
		color: rgba(255, 255, 255, 0.3);
		margin-top: 4px;
	}
</style>
