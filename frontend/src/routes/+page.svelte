<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth.svelte';

	onMount(() => {
		auth.init();
		goto(auth.isLoggedIn ? '/courses' : '/auth/login', { replaceState: true });
	});
</script>
