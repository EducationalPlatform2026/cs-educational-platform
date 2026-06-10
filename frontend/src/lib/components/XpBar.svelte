<script lang="ts">
	import { userStore } from '$lib/stores/userStore.svelte';
</script>

<div class="xp-bar">
	<span class="lv">Lv {userStore.level}</span>
	<div class="track">
		<div class="fill" style="width: {userStore.levelProgress}%"></div>
	</div>
	<span class="label">{userStore.xpIntoLevel}/500</span>
</div>

<style>
	.xp-bar {
		display: flex;
		align-items: center;
		gap: 6px;
		min-width: 148px;
	}
	.lv {
		font-size: 0.72rem;
		font-weight: 700;
		color: #7c3aed;
		white-space: nowrap;
	}
	.track {
		flex: 1;
		height: 5px;
		background: #ede9fe;
		border-radius: 99px;
		overflow: hidden;
		min-width: 60px;
	}
	.fill {
		height: 100%;
		background: #7c3aed;
		border-radius: 99px;
		transition: width 0.5s ease;
	}
	.label {
		font-size: 0.68rem;
		color: #9ca3af;
		white-space: nowrap;
	}
</style>
