<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth } from '$lib/stores/auth.svelte';
	import { userStore } from '$lib/stores/userStore.svelte';
	import { theme } from '$lib/stores/theme.svelte';

	const { children } = $props();
	const publicRoutes = ['/auth/login', '/auth/register'];

	onMount(() => {
		auth.init();
		userStore.init();
		theme.init();
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
				<span>CS Educational Platform</span>
			</a>

			{#if auth.isLoggedIn}
				<div class="nav-links">
					<a href="/dashboard" class:active={$page.url.pathname.startsWith('/dashboard')}>Dashboard</a>
					<a href="/courses"   class:active={$page.url.pathname.startsWith('/courses')}>Courses</a>
					<a href="/sandbox"   class:active={$page.url.pathname.startsWith('/sandbox')}>Sandbox</a>
					{#if auth.user?.role === 'admin'}
						<a href="/admin" class:active={$page.url.pathname.startsWith('/admin')} class="admin-link">Admin</a>
					{/if}
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
					<button
						class="btn-theme"
						onclick={() => theme.toggle()}
						title={theme.dark ? 'Switch to light mode' : 'Switch to dark mode'}
						aria-label="Toggle dark mode"
					>
						{theme.dark ? '☀' : '☾'}
					</button>
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
	/* ── CSS variables ─────────────────────────────────────────────────────── */
	:global(:root) {
		--bg:           #faf8ff;
		--bg-card:      #ffffff;
		--bg-surface:   #f9fafb;
		--bg-input:     #ffffff;
		--border:       #e5e7eb;
		--border-light: #f3f4f6;
		--text:         #1a1a2e;
		--text-2:       #374151;
		--text-3:       #6b7280;
		--text-4:       #9ca3af;
		--primary:      #7c3aed;
		--primary-h:    #6d28d9;
		--primary-bg:   #ede9fe;
		--primary-faint:#f5f3ff;
	}

	:global([data-theme='dark']) {
		--bg:           #0e0e17;
		--bg-card:      #1a1a27;
		--bg-surface:   #14141f;
		--bg-input:     #1e1e2d;
		--border:       #2c2c40;
		--border-light: #222232;
		--text:         #eeeef8;
		--text-2:       #b5b5cc;
		--text-3:       #8080a0;
		--text-4:       #60607a;
		--primary:      #8b5cf6;
		--primary-h:    #7c3aed;
		--primary-bg:   #2a1d58;
		--primary-faint:#1b1238;
	}

	/* ── Global resets ─────────────────────────────────────────────────────── */
	:global(*, *::before, *::after) { box-sizing: border-box; margin: 0; padding: 0; }

	:global(body) {
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
		background: var(--bg);
		color: var(--text);
		line-height: 1.5;
		transition: background 0.2s, color 0.2s;
	}

	:global(a) { color: var(--primary); text-decoration: none; }
	:global(a:hover) { text-decoration: underline; }

	/* ── Global form elements ──────────────────────────────────────────────── */
	:global(input:not([type='checkbox']):not([type='radio'])),
	:global(select),
	:global(textarea) {
		background: var(--bg-input);
		color: var(--text);
		border-color: var(--border);
	}

	/* ── App shell ─────────────────────────────────────────────────────────── */
	.app { min-height: 100vh; display: flex; flex-direction: column; }

	header {
		background: var(--bg-card);
		border-bottom: 1px solid var(--border);
		padding: 0 2rem;
		position: sticky;
		top: 0;
		z-index: 100;
		transition: background 0.2s, border-color 0.2s;
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
		color: var(--primary);
		text-decoration: none;
		flex-shrink: 0;
	}
	.brand:hover { text-decoration: none; }

	.brand-mark {
		background: var(--primary);
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
		color: var(--text-3);
		font-weight: 500;
		padding: 5px 12px;
		border-radius: 8px;
		transition: all 0.15s;
		text-decoration: none;
	}
	.nav-links a:hover { background: var(--primary-faint); color: var(--primary); text-decoration: none; }
	.nav-links a.active { background: var(--primary-bg); color: var(--primary); font-weight: 600; }
	.nav-links a.admin-link { color: #b91c1c; }
	.nav-links a.admin-link:hover { background: #fee2e2; color: #b91c1c; }
	.nav-links a.admin-link.active { background: #fee2e2; color: #b91c1c; }

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
		background: var(--primary-bg);
		color: var(--primary);
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
		color: var(--text-2);
	}

	.btn-theme {
		background: none;
		border: 1px solid var(--border);
		color: var(--text-3);
		width: 32px;
		height: 32px;
		border-radius: 6px;
		font-size: 1rem;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.15s;
		font-family: inherit;
		flex-shrink: 0;
	}
	.btn-theme:hover { border-color: var(--primary); color: var(--primary); }

	.btn-logout {
		background: none;
		border: 1px solid var(--border);
		color: var(--text-4);
		padding: 4px 12px;
		border-radius: 6px;
		font-size: 0.82rem;
		cursor: pointer;
		transition: all 0.15s;
		font-family: inherit;
	}
	.btn-logout:hover { border-color: var(--border); color: var(--text-3); }

	main {
		flex: 1;
		max-width: 1200px;
		margin: 0 auto;
		width: 100%;
		padding: 2rem;
	}
</style>
