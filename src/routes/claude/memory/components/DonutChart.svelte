<script lang="ts">
	type DataItem = {
		label: string;
		value: number;
	};

	let { data = [], title = '' }: { data: DataItem[]; title?: string } = $props();

	// Chart dimensions
	const size = 200;
	const cx = size / 2;
	const cy = size / 2;
	const outerRadius = 80;
	const innerRadius = 50;

	// Colors for segments
	const colors = ['var(--aion-purple)', 'var(--aion-yellow)', '#4ade80', '#f472b6', '#60a5fa'];

	const total = $derived(data.reduce((sum, d) => sum + d.value, 0));

	// Generate arc paths
	function polarToCartesian(centerX: number, centerY: number, radius: number, angleInDegrees: number) {
		const angleInRadians = ((angleInDegrees - 90) * Math.PI) / 180;
		return {
			x: centerX + radius * Math.cos(angleInRadians),
			y: centerY + radius * Math.sin(angleInRadians)
		};
	}

	function describeArc(
		x: number,
		y: number,
		innerR: number,
		outerR: number,
		startAngle: number,
		endAngle: number
	) {
		const start1 = polarToCartesian(x, y, outerR, endAngle);
		const end1 = polarToCartesian(x, y, outerR, startAngle);
		const start2 = polarToCartesian(x, y, innerR, endAngle);
		const end2 = polarToCartesian(x, y, innerR, startAngle);

		const largeArcFlag = endAngle - startAngle <= 180 ? 0 : 1;

		return [
			'M',
			start1.x,
			start1.y,
			'A',
			outerR,
			outerR,
			0,
			largeArcFlag,
			0,
			end1.x,
			end1.y,
			'L',
			end2.x,
			end2.y,
			'A',
			innerR,
			innerR,
			0,
			largeArcFlag,
			1,
			start2.x,
			start2.y,
			'Z'
		].join(' ');
	}

	const arcs = $derived(() => {
		if (total === 0) return [];
		let currentAngle = 0;
		return data.map((d, i) => {
			const angle = (d.value / total) * 360;
			const startAngle = currentAngle;
			const endAngle = currentAngle + angle;
			currentAngle = endAngle;
			return {
				path: describeArc(cx, cy, innerRadius, outerRadius, startAngle, endAngle - 0.5),
				color: colors[i % colors.length],
				label: d.label,
				value: d.value,
				percentage: ((d.value / total) * 100).toFixed(1)
			};
		});
	});
</script>

<div class="donut-chart">
	{#if title}
		<h3>{title}</h3>
	{/if}

	<div class="chart-container">
		<svg viewBox="0 0 {size} {size}" preserveAspectRatio="xMidYMid meet">
			{#each arcs() as arc}
				<path d={arc.path} fill={arc.color} class="segment">
					<title>{arc.label}: {arc.value.toLocaleString()} ({arc.percentage}%)</title>
				</path>
			{/each}
			<!-- Center text -->
			<text x={cx} y={cy - 8} class="center-value">{total.toLocaleString()}</text>
			<text x={cx} y={cy + 12} class="center-label">total</text>
		</svg>

		<div class="legend">
			{#each arcs() as arc, i}
				<div class="legend-item">
					<span class="legend-color" style="background: {arc.color}"></span>
					<span class="legend-label">{arc.label}</span>
					<span class="legend-value">{arc.percentage}%</span>
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.donut-chart {
		margin-bottom: 40px;
	}

	h3 {
		margin-bottom: 20px;
		font-size: 16px;
		font-weight: 600;
		color: var(--aion-purple);
	}

	.chart-container {
		display: flex;
		align-items: center;
		gap: 40px;
		background: rgba(61, 61, 71, 0.2);
		border-radius: 8px;
		padding: 24px;
	}

	svg {
		width: 200px;
		height: 200px;
		flex-shrink: 0;
	}

	.segment {
		transition: opacity 0.2s;
	}

	.segment:hover {
		opacity: 0.8;
	}

	.center-value {
		font-size: 24px;
		font-weight: 700;
		fill: white;
		text-anchor: middle;
		font-family: 'JetBrains Mono', monospace;
	}

	.center-label {
		font-size: 11px;
		fill: rgba(255, 255, 255, 0.5);
		text-anchor: middle;
		text-transform: uppercase;
		letter-spacing: 0.1em;
	}

	.legend {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.legend-color {
		width: 12px;
		height: 12px;
		border-radius: 3px;
		flex-shrink: 0;
	}

	.legend-label {
		font-size: 14px;
		color: rgba(255, 255, 255, 0.8);
		text-transform: capitalize;
	}

	.legend-value {
		font-size: 14px;
		font-weight: 600;
		color: var(--aion-yellow);
		font-family: 'JetBrains Mono', monospace;
		margin-left: auto;
	}

	@media (max-width: 500px) {
		.chart-container {
			flex-direction: column;
		}

		.legend {
			width: 100%;
		}
	}
</style>
