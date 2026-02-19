<script lang="ts">
	type HeatmapData = {
		day_of_week: number;
		hour: number;
		count: number;
	};

	let { data = [] }: { data: HeatmapData[] } = $props();

	const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
	const hours = Array.from({ length: 24 }, (_, i) => i);

	// Create a lookup map for quick access
	const dataMap = $derived(() => {
		const map = new Map<string, number>();
		for (const d of data) {
			map.set(`${d.day_of_week}-${d.hour}`, d.count);
		}
		return map;
	});

	// Calculate max for color scaling
	const maxCount = $derived(Math.max(...data.map((d) => d.count), 1));

	// Get count for a cell
	function getCount(day: number, hour: number): number {
		return dataMap().get(`${day}-${hour}`) ?? 0;
	}

	// Get color intensity (purple gradient)
	function getColor(count: number): string {
		if (count === 0) return 'rgba(61, 61, 71, 0.3)';
		const intensity = count / maxCount;
		// Gradient from dark purple to bright purple/yellow
		if (intensity < 0.25) return 'rgba(189, 147, 249, 0.2)';
		if (intensity < 0.5) return 'rgba(189, 147, 249, 0.4)';
		if (intensity < 0.75) return 'rgba(189, 147, 249, 0.7)';
		return 'rgba(245, 230, 99, 0.9)'; // Yellow for hot spots
	}

	// Tooltip state
	let tooltip = $state<{ x: number; y: number; day: string; hour: number; count: number } | null>(
		null
	);

	function showTooltip(event: MouseEvent, day: number, hour: number) {
		const count = getCount(day, hour);
		const rect = (event.target as SVGElement).getBoundingClientRect();
		tooltip = {
			x: rect.left + rect.width / 2,
			y: rect.top - 10,
			day: days[day],
			hour,
			count
		};
	}

	function hideTooltip() {
		tooltip = null;
	}

	// Cell dimensions - transposed: hours on X (24 cols), days on Y (7 rows)
	const cellWidth = 28;
	const cellHeight = 28;
	const labelWidth = 40;
	const labelHeight = 25;
</script>

<div class="heatmap-container">
	<h3>Activity by Day & Hour</h3>

	<div class="heatmap-wrapper">
		<svg
			width={labelWidth + hours.length * cellWidth + 10}
			height={labelHeight + days.length * cellHeight + 10}
		>
			<!-- Hour labels (top) -->
			{#each hours as hour}
				{#if hour % 3 === 0}
					<text
						x={labelWidth + hour * cellWidth + cellWidth / 2}
						y={labelHeight - 8}
						class="label hour-label-top"
					>
						{hour.toString().padStart(2, '0')}:00
					</text>
				{/if}
			{/each}

			<!-- Day labels (left side) -->
			{#each days as day, i}
				<text
					x={labelWidth - 8}
					y={labelHeight + i * cellHeight + cellHeight / 2 + 4}
					class="label day-label-left"
				>
					{day}
				</text>
			{/each}

			<!-- Heatmap cells: X = hour, Y = day -->
			{#each days as _, dayIndex}
				{#each hours as hour}
					<rect
						x={labelWidth + hour * cellWidth}
						y={labelHeight + dayIndex * cellHeight}
						width={cellWidth - 2}
						height={cellHeight - 2}
						fill={getColor(getCount(dayIndex, hour))}
						rx="3"
						class="cell"
						onmouseenter={(e) => showTooltip(e, dayIndex, hour)}
						onmouseleave={hideTooltip}
					/>
				{/each}
			{/each}
		</svg>
	</div>

	<!-- Tooltip -->
	{#if tooltip}
		<div class="tooltip" style="left: {tooltip.x}px; top: {tooltip.y}px;">
			<strong>{tooltip.day} {tooltip.hour.toString().padStart(2, '0')}:00</strong>
			<span>{tooltip.count.toLocaleString()} messages</span>
		</div>
	{/if}

	<!-- Legend -->
	<div class="legend">
		<span class="legend-label">Less</span>
		<div class="legend-scale">
			<div class="legend-cell" style="background: rgba(61, 61, 71, 0.3)"></div>
			<div class="legend-cell" style="background: rgba(189, 147, 249, 0.2)"></div>
			<div class="legend-cell" style="background: rgba(189, 147, 249, 0.4)"></div>
			<div class="legend-cell" style="background: rgba(189, 147, 249, 0.7)"></div>
			<div class="legend-cell" style="background: rgba(245, 230, 99, 0.9)"></div>
		</div>
		<span class="legend-label">More</span>
	</div>
</div>

<style>
	.heatmap-container {
		margin-bottom: 40px;
	}

	h3 {
		margin-bottom: 16px;
		font-size: 16px;
		font-weight: 600;
		color: var(--aion-purple);
	}

	.heatmap-wrapper {
		overflow-x: auto;
		padding-bottom: 8px;
	}

	.label {
		font-size: 10px;
		fill: rgba(255, 255, 255, 0.6);
		font-family: 'JetBrains Mono', monospace;
	}

	.day-label-left {
		text-anchor: end;
	}

	.hour-label-top {
		text-anchor: middle;
	}

	.cell {
		cursor: pointer;
		transition: opacity 0.15s;
	}

	.cell:hover {
		opacity: 0.8;
		stroke: var(--aion-yellow);
		stroke-width: 1;
	}

	.tooltip {
		position: fixed;
		transform: translateX(-50%) translateY(-100%);
		background: rgba(30, 30, 38, 0.95);
		border: 1px solid var(--aion-purple);
		border-radius: 6px;
		padding: 8px 12px;
		font-size: 12px;
		pointer-events: none;
		z-index: 100;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.tooltip strong {
		color: var(--aion-yellow);
	}

	.tooltip span {
		color: rgba(255, 255, 255, 0.8);
	}

	.legend {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-top: 12px;
		justify-content: center;
	}

	.legend-label {
		font-size: 10px;
		color: rgba(255, 255, 255, 0.5);
	}

	.legend-scale {
		display: flex;
		gap: 2px;
	}

	.legend-cell {
		width: 16px;
		height: 16px;
		border-radius: 2px;
	}
</style>
