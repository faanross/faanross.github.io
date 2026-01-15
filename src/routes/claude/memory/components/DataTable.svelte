<script lang="ts">
	type Column = {
		key: string;
		label: string;
		sortable?: boolean;
		align?: 'left' | 'center' | 'right';
		width?: string;
		format?: (value: unknown) => string;
	};

	type Row = Record<string, unknown>;

	let {
		columns = [],
		rows = [],
		title = '',
		onRowClick = undefined,
		keyField = 'id'
	}: {
		columns: Column[];
		rows: Row[];
		title?: string;
		onRowClick?: (row: Row) => void;
		keyField?: string;
	} = $props();

	let sortKey = $state<string | null>(null);
	let sortDirection = $state<'asc' | 'desc'>('desc');

	function handleSort(key: string) {
		if (sortKey === key) {
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			sortKey = key;
			sortDirection = 'desc';
		}
	}

	const sortedRows = $derived(() => {
		if (!sortKey) return rows;

		return [...rows].sort((a, b) => {
			const aVal = a[sortKey!];
			const bVal = b[sortKey!];

			if (typeof aVal === 'number' && typeof bVal === 'number') {
				return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
			}

			const aStr = String(aVal ?? '');
			const bStr = String(bVal ?? '');
			return sortDirection === 'asc' ? aStr.localeCompare(bStr) : bStr.localeCompare(aStr);
		});
	});

	function getValue(row: Row, col: Column): string {
		const value = row[col.key];
		if (col.format) return col.format(value);
		if (value === null || value === undefined) return '-';
		if (typeof value === 'number') return value.toLocaleString();
		return String(value);
	}
</script>

<div class="data-table">
	{#if title}
		<h3>{title}</h3>
	{/if}

	<div class="table-wrapper">
		<table>
			<thead>
				<tr>
					{#each columns as col}
						<th
							class:sortable={col.sortable}
							class:sorted={sortKey === col.key}
							style:width={col.width}
							style:text-align={col.align ?? 'left'}
							onclick={() => col.sortable && handleSort(col.key)}
						>
							{col.label}
							{#if col.sortable}
								<span class="sort-icon">
									{#if sortKey === col.key}
										{sortDirection === 'asc' ? '↑' : '↓'}
									{:else}
										↕
									{/if}
								</span>
							{/if}
						</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each sortedRows() as row}
					<tr
						class:clickable={!!onRowClick}
						onclick={() => onRowClick?.(row)}
					>
						{#each columns as col}
							<td style:text-align={col.align ?? 'left'}>
								{getValue(row, col)}
							</td>
						{/each}
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	{#if rows.length === 0}
		<p class="no-data">No data available</p>
	{/if}
</div>

<style>
	.data-table {
		margin-bottom: 40px;
	}

	h3 {
		margin-bottom: 16px;
		font-size: 16px;
		font-weight: 600;
		color: var(--aion-purple);
	}

	.table-wrapper {
		overflow-x: auto;
		background: rgba(61, 61, 71, 0.2);
		border-radius: 8px;
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 13px;
	}

	thead {
		background: rgba(0, 0, 0, 0.2);
	}

	th {
		padding: 12px 16px;
		font-weight: 600;
		color: rgba(255, 255, 255, 0.7);
		text-transform: uppercase;
		font-size: 11px;
		letter-spacing: 0.05em;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
		white-space: nowrap;
	}

	th.sortable {
		cursor: pointer;
		user-select: none;
	}

	th.sortable:hover {
		color: var(--aion-purple);
	}

	th.sorted {
		color: var(--aion-yellow);
	}

	.sort-icon {
		margin-left: 4px;
		opacity: 0.5;
	}

	th.sorted .sort-icon {
		opacity: 1;
	}

	td {
		padding: 12px 16px;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
		color: rgba(255, 255, 255, 0.8);
	}

	tr.clickable {
		cursor: pointer;
		transition: background 0.15s;
	}

	tr.clickable:hover {
		background: rgba(189, 147, 249, 0.1);
	}

	tbody tr:last-child td {
		border-bottom: none;
	}

	.no-data {
		text-align: center;
		padding: 40px 0;
		color: rgba(255, 255, 255, 0.4);
	}

	/* Responsive */
	@media (max-width: 768px) {
		th, td {
			padding: 10px 12px;
		}

		.table-wrapper {
			font-size: 12px;
		}
	}
</style>
