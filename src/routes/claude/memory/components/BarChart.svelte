<script lang="ts">
	type BarData = {
		label: string;
		value: number;
	};

	let { data = [], title = '' }: { data: BarData[]; title?: string } = $props();

	const maxValue = $derived(Math.max(...data.map((d) => d.value), 1));

	function getWidth(value: number): number {
		return (value / maxValue) * 100;
	}

	function formatLabel(label: string): string {
		// Clean up MCP tool names
		if (label.startsWith('mcp__')) {
			const parts = label.split('__');
			return parts.slice(1).join(' → ');
		}
		return label;
	}
</script>

<div class="bar-chart">
	{#if title}
		<h3>{title}</h3>
	{/if}

	<div class="bars">
		{#each data as item, i}
			<div class="bar-row">
				<span class="bar-label">{formatLabel(item.label)}</span>
				<div class="bar-container">
					<div
						class="bar"
						style="width: {getWidth(item.value)}%"
						class:top-3={i < 3}
					></div>
					<span class="bar-value">{item.value.toLocaleString()}</span>
				</div>
			</div>
		{/each}
	</div>
</div>

<style>
	.bar-chart {
		margin-bottom: 40px;
	}

	h3 {
		margin-bottom: 20px;
		font-size: 16px;
		font-weight: 600;
		color: var(--aion-purple);
	}

	.bars {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.bar-row {
		display: grid;
		grid-template-columns: 140px 1fr;
		align-items: center;
		gap: 12px;
	}

	.bar-label {
		font-size: 12px;
		color: rgba(255, 255, 255, 0.7);
		text-align: right;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.bar-container {
		display: flex;
		align-items: center;
		gap: 10px;
		height: 24px;
	}

	.bar {
		height: 100%;
		background: linear-gradient(90deg, rgba(189, 147, 249, 0.6), rgba(189, 147, 249, 0.3));
		border-radius: 4px;
		min-width: 4px;
		transition: width 0.5s ease-out;
	}

	.bar.top-3 {
		background: linear-gradient(90deg, var(--aion-yellow), rgba(245, 230, 99, 0.5));
	}

	.bar-value {
		font-size: 12px;
		font-weight: 600;
		color: rgba(255, 255, 255, 0.6);
		min-width: 50px;
	}

	@media (max-width: 600px) {
		.bar-row {
			grid-template-columns: 100px 1fr;
		}

		.bar-label {
			font-size: 11px;
		}
	}
</style>
