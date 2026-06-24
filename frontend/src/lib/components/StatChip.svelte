<script lang="ts">
	let {
		label,
		value,
		icon,
		color = 'brand'
	}: {
		label: string;
		value: string | number;
		icon: string;
		color?: 'brand' | 'xp' | 'streak' | 'blue' | 'green';
	} = $props();
</script>

<div class="chip {color}">
	<span class="icon">{icon}</span>
	<div class="body">
		<div class="value">{value}</div>
		<div class="label">{label}</div>
	</div>
</div>

<style>
	.chip {
		background: var(--bg-card);
		border-radius: 14px;
		padding: 1rem 1.25rem;
		display: flex;
		align-items: center;
		gap: 0.85rem;
		border: 1px solid var(--border);
		transition: box-shadow 0.15s, transform 0.15s;
	}
	.chip:hover {
		box-shadow: 0 4px 14px rgba(0, 0, 0, 0.12);
		transform: translateY(-1px);
	}
	.icon { font-size: 1.5rem; line-height: 1; flex-shrink: 0; }
	.value { font-size: 1.55rem; font-weight: 700; line-height: 1.1; }
	.label { font-size: 0.76rem; color: var(--text-3); margin-top: 0.1rem; }

	.brand { border-color: var(--primary-bg); }
	.brand .value { color: var(--primary); }

	.xp { border-color: #fde68a; }
	.xp .value { color: #d97706; }

	.streak { border-color: #fbcfe8; }
	.streak .value { color: #db2777; }

	.blue { border-color: #bfdbfe; }
	.blue .value { color: #2563eb; }

	.green { border-color: #bbf7d0; }
	.green .value { color: #16a34a; }
</style>
