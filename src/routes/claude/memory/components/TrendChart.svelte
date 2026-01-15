<script lang="ts">
	type TrendData = {
		date: string;
		count: number;
	};

	let { data = [], title = '' }: { data: TrendData[]; title?: string } = $props();

	// Chart dimensions
	const width = 700;
	const height = 200;
	const padding = { top: 20, right: 20, bottom: 30, left: 50 };

	const innerWidth = width - padding.left - padding.right;
	const innerHeight = height - padding.top - padding.bottom;

	// Scales
	const maxCount = $derived(Math.max(...data.map((d) => d.count), 1));

	function xScale(index: number): number {
		return padding.left + (index / (data.length - 1)) * innerWidth;
	}

	function yScale(value: number): number {
		return padding.top + innerHeight - (value / maxCount) * innerHeight;
	}

	// Generate path for area
	const areaPath = $derived(() => {
		if (data.length === 0) return '';
		const points = data.map((d, i) => `${xScale(i)},${yScale(d.count)}`);
		const baseline = `${xScale(data.length - 1)},${yScale(0)} ${xScale(0)},${yScale(0)}`;
		return `M${points.join(' L')} L${baseline} Z`;
	});

	// Generate path for line
	const linePath = $derived(() => {
		if (data.length === 0) return '';
		const points = data.map((d, i) => `${xScale(i)},${yScale(d.count)}`);
		return `M${points.join(' L')}`;
	});

	// Format date for display
	function formatDate(dateStr: string): string {
		// DuckDB WASM returns dates in various formats, handle them robustly
		if (!dateStr) return '';

		let date: Date;
		const dateVal = String(dateStr);

		// Check if it's a Unix timestamp (number)
		const numVal = Number(dateVal);
		if (!isNaN(numVal) && numVal > 1000000000 && numVal < 2000000000) {
			// Unix timestamp in seconds
			date = new Date(numVal * 1000);
		} else if (!isNaN(numVal) && numVal > 1000000000000) {
			// Unix timestamp in milliseconds
			date = new Date(numVal);
		} else {
			// Try parsing as-is first
			date = new Date(dateVal);

			// If invalid, try extracting date parts (handles "2026-01-13" format)
			if (isNaN(date.getTime())) {
				const parts = dateVal.match(/(\d{4})-(\d{2})-(\d{2})/);
				if (parts) {
					date = new Date(parseInt(parts[1]), parseInt(parts[2]) - 1, parseInt(parts[3]));
				}
			}
		}

		if (isNaN(date.getTime())) return dateVal.slice(0, 10);

		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	// Y-axis ticks
	const yTicks = $derived(() => {
		const ticks = [];
		const step = Math.ceil(maxCount / 4);
		for (let i = 0; i <= maxCount; i += step) {
			ticks.push(i);
		}
		return ticks;
	});
</script>

<div class="trend-chart">
	{#if title}
		<h3>{title}</h3>
	{/if}

	<div class="chart-wrapper">
		<svg viewBox="0 0 {width} {height}" preserveAspectRatio="xMidYMid meet">
			<!-- Y-axis grid lines and labels -->
			{#each yTicks() as tick}
				<line
					x1={padding.left}
					y1={yScale(tick)}
					x2={width - padding.right}
					y2={yScale(tick)}
					class="grid-line"
				/>
				<text x={padding.left - 8} y={yScale(tick) + 4} class="y-label">
					{tick >= 1000 ? `${(tick / 1000).toFixed(0)}k` : tick}
				</text>
			{/each}

			<!-- Area fill -->
			<path d={areaPath()} class="area" />

			<!-- Line -->
			<path d={linePath()} class="line" />

			<!-- Data points -->
			{#each data as d, i}
				<circle cx={xScale(i)} cy={yScale(d.count)} r="3" class="point" />
			{/each}

			<!-- X-axis labels (first, middle, last) -->
			{#if data.length > 0}
				<text x={xScale(0)} y={height - 8} class="x-label start">
					{formatDate(data[0].date)}
				</text>
				<text x={xScale(data.length - 1)} y={height - 8} class="x-label end">
					{formatDate(data[data.length - 1].date)}
				</text>
			{/if}
		</svg>
	</div>

	<!-- Summary stats -->
	<div class="stats-row">
		<div class="stat">
			<span class="stat-label">Total</span>
			<span class="stat-value">{data.reduce((sum, d) => sum + d.count, 0).toLocaleString()}</span>
		</div>
		<div class="stat">
			<span class="stat-label">Peak Day</span>
			<span class="stat-value">
				{Math.max(...data.map((d) => d.count)).toLocaleString()}
			</span>
		</div>
		<div class="stat">
			<span class="stat-label">Avg/Day</span>
			<span class="stat-value">
				{data.length > 0
					? Math.round(data.reduce((sum, d) => sum + d.count, 0) / data.length).toLocaleString()
					: 0}
			</span>
		</div>
	</div>
</div>

<style>
	.trend-chart {
		margin-bottom: 40px;
	}

	h3 {
		margin-bottom: 20px;
		font-size: 16px;
		font-weight: 600;
		color: var(--aion-purple);
	}

	.chart-wrapper {
		background: rgba(61, 61, 71, 0.2);
		border-radius: 8px;
		padding: 16px;
	}

	svg {
		width: 100%;
		height: auto;
	}

	.grid-line {
		stroke: rgba(255, 255, 255, 0.1);
		stroke-width: 1;
	}

	.area {
		fill: url(#areaGradient);
		fill: rgba(189, 147, 249, 0.2);
	}

	.line {
		fill: none;
		stroke: var(--aion-purple);
		stroke-width: 2;
		stroke-linecap: round;
		stroke-linejoin: round;
	}

	.point {
		fill: var(--aion-purple);
		stroke: rgba(0, 0, 0, 0.3);
		stroke-width: 1;
	}

	.y-label {
		font-size: 10px;
		fill: rgba(255, 255, 255, 0.5);
		text-anchor: end;
		font-family: 'JetBrains Mono', monospace;
	}

	.x-label {
		font-size: 10px;
		fill: rgba(255, 255, 255, 0.5);
		font-family: 'JetBrains Mono', monospace;
	}

	.x-label.start {
		text-anchor: start;
	}

	.x-label.end {
		text-anchor: end;
	}

	.stats-row {
		display: flex;
		justify-content: center;
		gap: 40px;
		margin-top: 20px;
	}

	.stat {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
	}

	.stat-label {
		font-size: 11px;
		color: rgba(255, 255, 255, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.stat-value {
		font-size: 18px;
		font-weight: 600;
		color: var(--aion-yellow);
	}

	@media (max-width: 600px) {
		.stats-row {
			gap: 20px;
		}

		.stat-value {
			font-size: 16px;
		}
	}
</style>
