<script lang="ts">
	let {
		state = 'locked',
		label = '',
		sub = '',
		href = ''
	}: {
		state: 'done' | 'current' | 'locked';
		label: string;
		sub?: string;
		href?: string;
	} = $props();
</script>

<div class="node-row">
	{#if state !== 'locked' && href}
		<a {href} class="node {state}" aria-label={label}>
			{#if state === 'done'}✓{:else}●{/if}
		</a>
	{:else}
		<div class="node {state}" aria-label={label}>
			{#if state === 'locked'}🔒{:else if state === 'done'}✓{:else}●{/if}
		</div>
	{/if}
	<div class="node-text">
		<span class="node-label {state === 'locked' ? 'muted' : ''}">{label}</span>
		{#if sub}<span class="node-sub">{sub}</span>{/if}
	</div>
</div>

<style>
	.node-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.node {
		width: 30px;
		height: 30px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.78rem;
		font-weight: 700;
		flex-shrink: 0;
		text-decoration: none;
		transition: transform 0.15s;
		line-height: 1;
	}
	.done {
		background: var(--primary);
		color: #fff;
	}
	.current {
		background: var(--bg-card);
		border: 2px solid var(--primary);
		color: var(--primary);
		animation: pulse-ring 2s ease-in-out infinite;
	}
	.locked {
		background: var(--primary-faint);
		color: #a78bfa;
		font-size: 0.65rem;
		cursor: default;
	}
	a.done:hover, a.current:hover {
		transform: scale(1.12);
		box-shadow: 0 2px 8px rgba(124, 58, 237, 0.3);
	}
	.node-label { font-size: 0.875rem; font-weight: 500; color: var(--text-2); }
	.node-label.muted { color: var(--text-4); }
	.node-sub { display: block; font-size: 0.72rem; color: var(--text-4); margin-top: 1px; }

	@keyframes pulse-ring {
		0%, 100% { box-shadow: 0 0 0 0 rgba(124, 58, 237, 0.35); }
		50% { box-shadow: 0 0 0 6px rgba(124, 58, 237, 0); }
	}
</style>
