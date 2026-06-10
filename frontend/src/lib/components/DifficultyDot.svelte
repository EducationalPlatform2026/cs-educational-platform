<script lang="ts">
	let { difficulty }: { difficulty: 'easy' | 'medium' | 'hard' } = $props();
</script>

<span class="dot {difficulty}" title={difficulty}></span>

<style>
	.dot {
		display: inline-block;
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
		vertical-align: middle;
	}
	.easy   { background: #10b981; }
	.medium { background: #f59e0b; }
	.hard   { background: #ef4444; }
</style>
