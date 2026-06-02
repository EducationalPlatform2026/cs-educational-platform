<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth } from '$lib/stores/auth.svelte';

	const { children } = $props();

	const publicRoutes = ['/auth/login', '/auth/register'];

	onMount(() => {
		auth.init();
		const isPublic = publicRoutes.some((r) => $page.url.pathname.startsWith(r));
		if (!auth.isLoggedIn && !isPublic) {
			goto('/auth/login');
		}
	});

	function handleLogout() {
		auth.logout();
		goto('/auth/login');
	}
</script>

<div class="app">
	<header>
		<nav>
			<a href="/" class="brand">CS Platform</a>
			{#if auth.isLoggedIn}
				<div class="nav-links">
					<a href="/dashboard">Dashboard</a>
					<a href="/courses">Courses</a>
				</div>
				<div class="nav-user">
					<span class="user-name">{auth.user?.first_name} {auth.user?.last_name}</span>
					<span class="user-role">{auth.user?.role}</span>
					<button onclick={handleLogout} class="btn-logout">Log out</button>
				</div>
			{/if}
		</nav>
	</header>

	<main>
		{@render children()}
	</main>
</div>

<style>
	:global(*, *::before, *::after) {
		box-sizing: border-box;
		margin: 0;
		padding: 0;
	}

	:global(body) {
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
		background: #f5f7fa;
		color: #1a1a2e;
		line-height: 1.5;
	}

	:global(a) {
		color: #4f46e5;
		text-decoration: none;
	}

	:global(a:hover) {
		text-decoration: underline;
	}

	.app {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
	}

	header {
		background: #fff;
		border-bottom: 1px solid #e5e7eb;
		padding: 0 2rem;
		position: sticky;
		top: 0;
		z-index: 10;
	}

	nav {
		max-width: 1100px;
		margin: 0 auto;
		height: 56px;
		display: flex;
		align-items: center;
		gap: 2rem;
	}

	.brand {
		font-size: 1.1rem;
		font-weight: 700;
		color: #4f46e5;
	}

	.nav-links {
		display: flex;
		gap: 1.5rem;
		flex: 1;
	}

	.nav-links a {
		font-size: 0.95rem;
		color: #374151;
		font-weight: 500;
	}

	.nav-user {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-left: auto;
	}

	.user-name {
		font-weight: 600;
		font-size: 0.9rem;
	}

	.user-role {
		background: #eef2ff;
		color: #4f46e5;
		font-size: 0.75rem;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 99px;
		text-transform: capitalize;
	}

	.btn-logout {
		background: none;
		border: 1px solid #d1d5db;
		color: #6b7280;
		padding: 4px 12px;
		border-radius: 6px;
		font-size: 0.85rem;
		cursor: pointer;
		transition: all 0.15s;
	}

	.btn-logout:hover {
		border-color: #9ca3af;
		color: #374151;
	}

	main {
		flex: 1;
		max-width: 1100px;
		margin: 0 auto;
		width: 100%;
		padding: 2rem;
	}
</style>
