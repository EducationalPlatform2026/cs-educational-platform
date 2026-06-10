<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth } from '$lib/stores/auth.svelte';
	import { userStore } from '$lib/stores/userStore.svelte';

	const { children } = $props();
	const publicRoutes = ['/auth/login', '/auth/register'];

	onMount(() => {
		auth.init();
		userStore.init();
		const isPublic = publicRoutes.some((r) => $page.url.pathname.startsWith(r));
		if (!auth.isLoggedIn && !isPublic) goto('/auth/login');
	});

	function handleLogout() {
		auth.logout();
		userStore.reset();
		goto('/auth/login');
	}

	const initials = $derived(
		((auth.user?.first_name?.[0] ?? '') + (auth.user?.last_name?.[0] ?? '')).toUpperCase()
	);
</script>

<div class="app">
	<header>
		<nav>
			<a href="/" class="brand">
				<span class="brand-mark">⚡</span>
				<span>CodePath</span>
			</a>

			{#if auth.isLoggedIn}
				<div class="nav-links">
					<a href="/dashboard" class:active={$page.url.pathname.startsWith('/dashboard')}>Dashboard</a>
					<a href="/courses"   class:active={$page.url.pathname.startsWith('/courses')}>Courses</a>
					<a href="/sandbox"   class:active={$page.url.pathname.startsWith('/sandbox')}>Sandbox</a>
				</div>

				<div class="nav-right">
					<div class="xp-chip" title="Total XP">
						<span>⚡</span>
						<span>{userStore.xp.toLocaleString()} XP</span>
					</div>
					{#if userStore.streak > 0}
						<div class="streak-chip" title="Current streak">
							🔥 {userStore.streak}
						</div>
					{/if}
					<div class="avatar-wrap">
						<div class="avatar">{initials}</div>
						<span class="user-name">{auth.user?.first_name}</span>
					</div>
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
	:global(*, *::before, *::after) { box-sizing: border-box; margin: 0; padding: 0; }

	:global(body) {
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
		background: #faf8ff;
		color: #1a1a2e;
		line-height: 1.5;
	}

	:global(a) { color: #7c3aed; text-decoration: none; }
	:global(a:hover) { text-decoration: underline; }

	.app { min-height: 100vh; display: flex; flex-direction: column; }

	header {
		background: #fff;
		border-bottom: 1px solid #e5e7eb;
		padding: 0 2rem;
		position: sticky;
		top: 0;
		z-index: 100;
	}

	nav {
		max-width: 1200px;
		margin: 0 auto;
		height: 58px;
		display: flex;
		align-items: center;
		gap: 2rem;
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 7px;
		font-size: 1.05rem;
		font-weight: 800;
		color: #7c3aed;
		text-decoration: none;
		flex-shrink: 0;
	}
	.brand:hover { text-decoration: none; }

	.brand-mark {
		background: #7c3aed;
		color: #fff;
		width: 26px;
		height: 26px;
		border-radius: 7px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.85rem;
		flex-shrink: 0;
	}

	.nav-links {
		display: flex;
		gap: 0.25rem;
		flex: 1;
	}

	.nav-links a {
		font-size: 0.9rem;
		color: #6b7280;
		font-weight: 500;
		padding: 5px 12px;
		border-radius: 8px;
		transition: all 0.15s;
		text-decoration: none;
	}
	.nav-links a:hover { background: #f3f0ff; color: #7c3aed; text-decoration: none; }
	.nav-links a.active { background: #ede9fe; color: #7c3aed; font-weight: 600; }

	.nav-right {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		margin-left: auto;
	}

	.xp-chip {
		background: #fef3c7;
		color: #d97706;
		padding: 4px 11px;
		border-radius: 99px;
		font-size: 0.78rem;
		font-weight: 700;
		display: flex;
		align-items: center;
		gap: 4px;
		white-space: nowrap;
	}

	.streak-chip {
		background: #fce7f3;
		color: #db2777;
		padding: 4px 11px;
		border-radius: 99px;
		font-size: 0.78rem;
		font-weight: 700;
		white-space: nowrap;
	}

	.avatar-wrap {
		display: flex;
		align-items: center;
		gap: 7px;
	}

	.avatar {
		width: 30px;
		height: 30px;
		border-radius: 50%;
		background: #ede9fe;
		color: #7c3aed;
		font-size: 0.72rem;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.user-name {
		font-size: 0.875rem;
		font-weight: 600;
		color: #374151;
	}

	.btn-logout {
		background: none;
		border: 1px solid #e5e7eb;
		color: #9ca3af;
		padding: 4px 12px;
		border-radius: 6px;
		font-size: 0.82rem;
		cursor: pointer;
		transition: all 0.15s;
		font-family: inherit;
	}
	.btn-logout:hover { border-color: #d1d5db; color: #6b7280; }

	main {
		flex: 1;
		max-width: 1200px;
		margin: 0 auto;
		width: 100%;
		padding: 2rem;
	}
</style>
