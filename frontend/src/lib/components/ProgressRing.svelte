<script lang="ts">
	let {
		percent = 0,
		size = 72,
		stroke = 6,
		label = ''
	}: {
		percent: number;
		size?: number;
		stroke?: number;
		label?: string;
	} = $props();

	const r = $derived((size - stroke) / 2);
	const circumference = $derived(2 * Math.PI * r);
	const offset = $derived(circumference - (Math.min(percent, 100) / 100) * circumference);
</script>

<div class="ring" style="width:{size}px;height:{size}px;">
	<svg width={size} height={size} viewBox="0 0 {size} {size}" style="transform:rotate(-90deg);display:block;">
		<circle cx={size / 2} cy={size / 2} r={r} fill="none" stroke="#ede9fe" stroke-width={stroke} />
		<circle
			cx={size / 2} cy={size / 2} r={r}
			fill="none"
			stroke="#7c3aed"
			stroke-width={stroke}
			stroke-linecap="round"
			stroke-dasharray={circumference}
			stroke-dashoffset={offset}
			style="transition: stroke-dashoffset 0.5s ease"
		/>
	</svg>
	{#if label}
		<div class="ring-label" style="font-size:{Math.round(size * 0.22)}px">{label}</div>
	{/if}
</div>

<style>
	.ring {
		position: relative;
		display: inline-block;
		flex-shrink: 0;
	}
	.ring-label {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		color: #7c3aed;
		line-height: 1;
	}
</style>
