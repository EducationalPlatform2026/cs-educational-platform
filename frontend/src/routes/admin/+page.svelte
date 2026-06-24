<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth.svelte';
	import {
		getAdminStats, listAdminUsers, listAdminCourses,
		updateUserRole, toggleUserActive,
		type AdminUser, type AdminCourse, type AdminStats
	} from '$lib/api/admin';

	let tab = $state<'overview' | 'users' | 'courses'>('overview');
	let stats = $state<AdminStats | null>(null);
	let users = $state<AdminUser[]>([]);
	let courses = $state<AdminCourse[]>([]);
	let loading = $state(true);
	let error = $state('');
	let search = $state('');
	let updatingRole = $state<Record<string, boolean>>({});
	let togglingActive = $state<Record<string, boolean>>({});
	let saveMsg = $state<Record<string, string>>({});

	const ROLES = ['student', 'teaching_assistant', 'professor', 'admin'];

	const filteredUsers = $derived(
		search.trim()
			? users.filter(u =>
				`${u.first_name} ${u.last_name} ${u.email}`.toLowerCase().includes(search.toLowerCase())
			)
			: users
	);

	const filteredCourses = $derived(
		search.trim()
			? courses.filter(c =>
				`${c.title} ${c.creator_name}`.toLowerCase().includes(search.toLowerCase())
			)
			: courses
	);

	onMount(async () => {
		auth.init();
		if (auth.user?.role !== 'admin') { goto('/dashboard'); return; }
		try {
			const [s, u, c] = await Promise.all([getAdminStats(), listAdminUsers(), listAdminCourses()]);
			stats = s;
			users = u;
			courses = c;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load admin data';
		} finally {
			loading = false;
		}
	});

	async function handleRoleChange(user: AdminUser, newRole: string) {
		updatingRole = { ...updatingRole, [user.id]: true };
		try {
			await updateUserRole(user.id, newRole);
			users = users.map(u => u.id === user.id ? { ...u, role: newRole } : u);
			saveMsg = { ...saveMsg, [user.id]: '✓ Saved' };
			setTimeout(() => { saveMsg = { ...saveMsg, [user.id]: '' }; }, 2000);
		} catch (err: unknown) {
			saveMsg = { ...saveMsg, [user.id]: '✗ Failed' };
			setTimeout(() => { saveMsg = { ...saveMsg, [user.id]: '' }; }, 2000);
		} finally {
			updatingRole = { ...updatingRole, [user.id]: false };
		}
	}

	async function handleToggleActive(user: AdminUser) {
		togglingActive = { ...togglingActive, [user.id]: true };
		try {
			const res = await toggleUserActive(user.id);
			users = users.map(u => u.id === user.id ? { ...u, is_active: res.is_active } : u);
			if (stats) {
				stats = { ...stats, active_users: stats.active_users + (res.is_active ? 1 : -1) };
			}
		} catch {
			alert('Failed to toggle user status');
		} finally {
			togglingActive = { ...togglingActive, [user.id]: false };
		}
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
	}

	function roleColor(role: string) {
		return { student: 'blue', teaching_assistant: 'teal', professor: 'purple', admin: 'red' }[role] ?? 'gray';
	}
</script>

<div class="page">
	<div class="page-header">
		<div>
			<h1>Admin Panel</h1>
			<p class="sub">Platform management — users, courses, and system stats</p>
		</div>
	</div>

	{#if error}
		<div class="alert">{error}</div>
	{:else if loading}
		<div class="skeleton"></div>
	{:else}
		<!-- Tabs -->
		<div class="tabs">
			<button class="tab" class:active={tab === 'overview'} onclick={() => { tab = 'overview'; search = ''; }}>
				📊 Overview
			</button>
			<button class="tab" class:active={tab === 'users'} onclick={() => { tab = 'users'; search = ''; }}>
				👥 Users {#if stats}<span class="tab-count">{stats.total_users}</span>{/if}
			</button>
			<button class="tab" class:active={tab === 'courses'} onclick={() => { tab = 'courses'; search = ''; }}>
				📚 Courses {#if stats}<span class="tab-count">{stats.total_courses}</span>{/if}
			</button>
		</div>

		<!-- ── Overview ── -->
		{#if tab === 'overview' && stats}
			<div class="stat-grid">
				<div class="stat-card brand">
					<div class="stat-icon">👥</div>
					<div class="stat-body">
						<div class="stat-value">{stats.total_users}</div>
						<div class="stat-label">Total users</div>
						<div class="stat-sub">{stats.active_users} active</div>
					</div>
				</div>
				<div class="stat-card teal">
					<div class="stat-icon">📚</div>
					<div class="stat-body">
						<div class="stat-value">{stats.total_courses}</div>
						<div class="stat-label">Total courses</div>
						<div class="stat-sub">{stats.published_courses} published</div>
					</div>
				</div>
				<div class="stat-card amber">
					<div class="stat-icon">⚡</div>
					<div class="stat-body">
						<div class="stat-value">{stats.total_exercises}</div>
						<div class="stat-label">Total exercises</div>
					</div>
				</div>
				<div class="stat-card pink">
					<div class="stat-icon">📨</div>
					<div class="stat-body">
						<div class="stat-value">{stats.total_submissions}</div>
						<div class="stat-label">Total submissions</div>
					</div>
				</div>
			</div>

			<!-- Role breakdown -->
			<div class="section-card">
				<h2>Role breakdown</h2>
				<div class="role-breakdown">
					{#each ROLES as role}
						{@const count = users.filter(u => u.role === role).length}
						<div class="role-row">
							<span class="role-badge {roleColor(role)}">{role.replace('_', ' ')}</span>
							<div class="role-bar-wrap">
								<div class="role-bar" style="width: {users.length ? (count/users.length*100) : 0}%"></div>
							</div>
							<span class="role-count">{count}</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}

		<!-- ── Users ── -->
		{#if tab === 'users'}
			<div class="toolbar">
				<input class="search" bind:value={search} placeholder="Search by name or email…" />
				<span class="result-count">{filteredUsers.length} user{filteredUsers.length !== 1 ? 's' : ''}</span>
			</div>
			<div class="table-wrap">
				<table class="data-table">
					<thead>
						<tr>
							<th>Name</th>
							<th>Email</th>
							<th>Role</th>
							<th>Status</th>
							<th>Joined</th>
							<th>Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each filteredUsers as user (user.id)}
							<tr class:inactive={!user.is_active}>
								<td class="name-cell">
									<div class="avatar-sm">{user.first_name[0]}{user.last_name[0]}</div>
									<span>{user.first_name} {user.last_name}</span>
									{#if user.id === auth.user?.user_id}<span class="you-badge">you</span>{/if}
								</td>
								<td class="muted">{user.email}</td>
								<td>
									{#if user.id === auth.user?.user_id}
										<span class="role-badge {roleColor(user.role)}">{user.role.replace('_', ' ')}</span>
									{:else}
										<select
											class="role-select"
											value={user.role}
											disabled={updatingRole[user.id]}
											onchange={(e) => handleRoleChange(user, (e.currentTarget as HTMLSelectElement).value)}
										>
											{#each ROLES as r}
												<option value={r}>{r.replace('_', ' ')}</option>
											{/each}
										</select>
										{#if saveMsg[user.id]}
											<span class="save-msg" class:ok={saveMsg[user.id].startsWith('✓')}>
												{saveMsg[user.id]}
											</span>
										{/if}
									{/if}
								</td>
								<td>
									<span class="status-badge" class:active={user.is_active}>
										{user.is_active ? 'Active' : 'Inactive'}
									</span>
								</td>
								<td class="muted">{formatDate(user.created_at)}</td>
								<td>
									{#if user.id !== auth.user?.user_id}
										<button
											class="btn-toggle"
											class:deactivate={user.is_active}
											onclick={() => handleToggleActive(user)}
											disabled={togglingActive[user.id]}
										>
											{togglingActive[user.id] ? '…' : user.is_active ? 'Deactivate' : 'Activate'}
										</button>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}

		<!-- ── Courses ── -->
		{#if tab === 'courses'}
			<div class="toolbar">
				<input class="search" bind:value={search} placeholder="Search by title or creator…" />
				<span class="result-count">{filteredCourses.length} course{filteredCourses.length !== 1 ? 's' : ''}</span>
			</div>
			<div class="table-wrap">
				<table class="data-table">
					<thead>
						<tr>
							<th>Title</th>
							<th>Creator</th>
							<th>Status</th>
							<th>Students</th>
							<th>Exercises</th>
							<th>Created</th>
							<th>Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each filteredCourses as course (course.id)}
							<tr>
								<td class="title-cell">
									<a href="/courses/{course.id}">{course.title}</a>
								</td>
								<td class="muted">{course.creator_name}</td>
								<td>
									<span class="status-badge" class:active={course.is_published}>
										{course.is_published ? 'Published' : 'Draft'}
									</span>
								</td>
								<td class="center">{course.enroll_count}</td>
								<td class="center">{course.exercise_count}</td>
								<td class="muted">{formatDate(course.created_at)}</td>
								<td>
									<a href="/courses/{course.id}" class="btn-view">View →</a>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	{/if}
</div>

<style>
	.page { max-width: 1100px; }

	.page-header { margin-bottom: 1.5rem; }
	h1 { font-size: 1.75rem; font-weight: 700; }
	.sub { color: #6b7280; font-size: 0.875rem; margin-top: 0.2rem; }

	.alert { background: #fef2f2; color: #b91c1c; border: 1px solid #fecaca; border-radius: 8px; padding: 0.75rem 1rem; margin-bottom: 1rem; }

	.skeleton { height: 200px; border-radius: 14px; background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
	@keyframes shimmer { 0% { background-position: 200% 0 } 100% { background-position: -200% 0 } }

	/* Tabs */
	.tabs { display: flex; gap: 0.25rem; margin-bottom: 1.5rem; border-bottom: 2px solid #f3f4f6; }
	.tab {
		background: none; border: none; padding: 0.6rem 1.1rem; font-size: 0.9rem; font-weight: 500;
		color: #6b7280; cursor: pointer; border-bottom: 2px solid transparent; margin-bottom: -2px;
		transition: all 0.15s; border-radius: 6px 6px 0 0; font-family: inherit; display: flex; align-items: center; gap: 0.4rem;
	}
	.tab:hover { color: #7c3aed; background: #f5f3ff; }
	.tab.active { color: #7c3aed; border-bottom-color: #7c3aed; font-weight: 600; }
	.tab-count { background: #ede9fe; color: #7c3aed; font-size: 0.72rem; font-weight: 700; padding: 1px 7px; border-radius: 99px; }

	/* Stats grid */
	.stat-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; margin-bottom: 1.5rem; }
	.stat-card {
		background: #fff; border: 1px solid #e5e7eb; border-radius: 14px;
		padding: 1.25rem; display: flex; gap: 1rem; align-items: center;
	}
	.stat-icon { font-size: 1.75rem; line-height: 1; }
	.stat-value { font-size: 1.6rem; font-weight: 800; }
	.stat-label { font-size: 0.8rem; color: #6b7280; font-weight: 500; }
	.stat-sub { font-size: 0.75rem; color: #9ca3af; margin-top: 2px; }

	.stat-card.brand .stat-value { color: #7c3aed; }
	.stat-card.teal  .stat-value { color: #0d9488; }
	.stat-card.amber .stat-value { color: #d97706; }
	.stat-card.pink  .stat-value { color: #db2777; }

	/* Role breakdown */
	.section-card { background: #fff; border: 1px solid #e5e7eb; border-radius: 14px; padding: 1.25rem; }
	h2 { font-size: 1rem; font-weight: 600; margin-bottom: 1rem; }
	.role-breakdown { display: flex; flex-direction: column; gap: 0.65rem; }
	.role-row { display: flex; align-items: center; gap: 0.75rem; }
	.role-bar-wrap { flex: 1; height: 8px; background: #f3f4f6; border-radius: 99px; overflow: hidden; }
	.role-bar { height: 100%; background: #7c3aed; border-radius: 99px; transition: width 0.4s ease; min-width: 2px; }
	.role-count { font-size: 0.82rem; font-weight: 600; color: #374151; min-width: 24px; text-align: right; }

	/* Toolbar */
	.toolbar { display: flex; align-items: center; gap: 1rem; margin-bottom: 1rem; }
	.search {
		flex: 1; max-width: 360px; padding: 0.5rem 0.85rem;
		border: 1px solid #e5e7eb; border-radius: 8px; font-size: 0.875rem;
		outline: none; font-family: inherit;
	}
	.search:focus { border-color: #a78bfa; }
	.result-count { font-size: 0.82rem; color: #9ca3af; }

	/* Table */
	.table-wrap { border: 1px solid #e5e7eb; border-radius: 12px; overflow-x: auto; }
	.data-table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
	.data-table th {
		text-align: left; padding: 0.65rem 1rem; font-size: 0.75rem; font-weight: 600;
		color: #6b7280; text-transform: uppercase; letter-spacing: 0.04em;
		background: #f9fafb; border-bottom: 1px solid #e5e7eb;
	}
	.data-table td { padding: 0.75rem 1rem; border-bottom: 1px solid #f3f4f6; color: #374151; vertical-align: middle; }
	.data-table tr:last-child td { border-bottom: none; }
	.data-table tr.inactive td { opacity: 0.5; }
	.data-table tr:hover td { background: #fafafa; }

	.name-cell { display: flex; align-items: center; gap: 0.6rem; }
	.avatar-sm {
		width: 28px; height: 28px; border-radius: 50%; background: #ede9fe; color: #7c3aed;
		font-size: 0.68rem; font-weight: 700; display: flex; align-items: center; justify-content: center; flex-shrink: 0; text-transform: uppercase;
	}
	.you-badge { background: #fef9c3; color: #92400e; font-size: 0.68rem; font-weight: 700; padding: 1px 6px; border-radius: 99px; }
	.muted { color: #6b7280; }
	.center { text-align: center; }
	.title-cell a { color: #7c3aed; font-weight: 500; }
	.title-cell a:hover { text-decoration: underline; }

	/* Role badge */
	.role-badge { display: inline-block; font-size: 0.75rem; font-weight: 600; padding: 2px 9px; border-radius: 99px; text-transform: capitalize; }
	.role-badge.blue   { background: #dbeafe; color: #1d4ed8; }
	.role-badge.teal   { background: #ccfbf1; color: #0f766e; }
	.role-badge.purple { background: #ede9fe; color: #7c3aed; }
	.role-badge.red    { background: #fee2e2; color: #b91c1c; }
	.role-badge.gray   { background: #f3f4f6; color: #6b7280; }

	/* Role select */
	.role-select {
		padding: 3px 8px; border: 1px solid #e5e7eb; border-radius: 6px;
		font-size: 0.8rem; color: #374151; background: #fff; outline: none; cursor: pointer; font-family: inherit;
	}
	.role-select:focus { border-color: #a78bfa; }
	.role-select:disabled { opacity: 0.5; }
	.save-msg { font-size: 0.75rem; font-weight: 600; margin-left: 0.4rem; color: #dc2626; }
	.save-msg.ok { color: #16a34a; }

	/* Status badge */
	.status-badge { font-size: 0.75rem; font-weight: 600; padding: 2px 9px; border-radius: 99px; background: #f3f4f6; color: #6b7280; }
	.status-badge.active { background: #dcfce7; color: #16a34a; }

	/* Buttons */
	.btn-toggle {
		font-size: 0.78rem; font-weight: 600; padding: 4px 10px; border-radius: 6px; border: 1px solid #e5e7eb;
		background: #f9fafb; color: #374151; cursor: pointer; font-family: inherit; transition: all 0.15s;
	}
	.btn-toggle.deactivate { border-color: #fecaca; background: #fef2f2; color: #b91c1c; }
	.btn-toggle.deactivate:hover:not(:disabled) { background: #fee2e2; }
	.btn-toggle:not(.deactivate):hover:not(:disabled) { background: #dcfce7; border-color: #bbf7d0; color: #15803d; }
	.btn-toggle:disabled { opacity: 0.5; cursor: not-allowed; }

	.btn-view { font-size: 0.8rem; color: #7c3aed; font-weight: 500; }
	.btn-view:hover { text-decoration: underline; }

	@media (max-width: 900px) {
		.stat-grid { grid-template-columns: 1fr 1fr; }
	}
	@media (max-width: 600px) {
		.stat-grid { grid-template-columns: 1fr; }
	}
</style>
