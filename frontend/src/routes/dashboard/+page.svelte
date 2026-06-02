<script lang="ts">
	import { onMount } from 'svelte';
	import { getDashboard, type DashboardStats } from '$lib/api/dashboard';
	import { auth } from '$lib/stores/auth.svelte';

	let stats = $state<DashboardStats | null>(null);
	let loading = $state(true);
	let error = $state('');

	const statusMeta: Record<string, { label: string; cls: string }> = {
		pending:       { label: 'Pending',       cls: 'status-pending' },
		running:       { label: 'Running',        cls: 'status-running' },
		accepted:      { label: 'Accepted',       cls: 'status-accepted' },
		wrong_answer:  { label: 'Wrong Answer',   cls: 'status-wrong' },
		runtime_error: { label: 'Runtime Error',  cls: 'status-error' },
		time_limit:    { label: 'Time Limit',     cls: 'status-error' },
		memory_limit:  { label: 'Memory Limit',   cls: 'status-error' },
		compile_error: { label: 'Compile Error',  cls: 'status-error' }
	};

	function statusInfo(s: string) {
		return statusMeta[s] ?? { label: s, cls: 'status-pending' };
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleString(undefined, {
			month: 'short', day: 'numeric',
			hour: '2-digit', minute: '2-digit'
		});
	}

	onMount(async () => {
		try {
			stats = await getDashboard();
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load dashboard';
		} finally {
			loading = false;
		}
	});
</script>

<div class="page">
	<div class="page-header">
		<div>
			<h1>Dashboard</h1>
			<p class="greeting">
				Welcome back, {auth.user?.first_name ?? 'there'} 👋
			</p>
		</div>
		<a href="/courses" class="btn-outline">Browse courses</a>
	</div>

	{#if loading}
		<div class="stat-grid">
			{#each [1,2,3,4] as _}
				<div class="stat-card skeleton"></div>
			{/each}
		</div>
		<div class="skeleton-table"></div>

	{:else if error}
		<div class="alert">{error}</div>

	{:else if stats}
		<!-- Stat cards -->
		<div class="stat-grid">
			<div class="stat-card">
				<span class="stat-icon">📚</span>
				<div class="stat-body">
					<div class="stat-value">{stats.enrolled_courses}</div>
					<div class="stat-label">Enrolled courses</div>
				</div>
			</div>

			<div class="stat-card">
				<span class="stat-icon">💡</span>
				<div class="stat-body">
					<div class="stat-value">{stats.exercises_attempted}</div>
					<div class="stat-label">Exercises attempted</div>
				</div>
			</div>

			<div class="stat-card">
				<span class="stat-icon">📨</span>
				<div class="stat-body">
					<div class="stat-value">{stats.total_submissions}</div>
					<div class="stat-label">Total submissions</div>
				</div>
			</div>

			<div class="stat-card accent">
				<span class="stat-icon">✅</span>
				<div class="stat-body">
					<div class="stat-value">
						{stats.total_submissions === 0
							? '—'
							: stats.acceptance_rate.toFixed(1) + '%'}
					</div>
					<div class="stat-label">
						Acceptance rate
						{#if stats.total_submissions > 0}
							<span class="sub">({stats.accepted_submissions}/{stats.total_submissions})</span>
						{/if}
					</div>
				</div>
			</div>
		</div>

		<!-- Recent submissions -->
		<section class="recent-section">
			<h2>Recent submissions</h2>

			{#if stats.recent_submissions.length === 0}
				<div class="empty">
					<p>No submissions yet.</p>
					<a href="/courses" class="btn-primary">Find an exercise to try</a>
				</div>
			{:else}
				<div class="table-wrap">
					<table class="sub-table">
						<thead>
							<tr>
								<th>Exercise</th>
								<th>Language</th>
								<th>Status</th>
								<th>Score</th>
								<th>Submitted</th>
								<th></th>
							</tr>
						</thead>
						<tbody>
							{#each stats.recent_submissions as sub (sub.id)}
								{@const meta = statusInfo(sub.status)}
								<tr>
									<td class="td-exercise">
										<a href="/exercises/{sub.exercise_id}">{sub.exercise_title}</a>
									</td>
									<td><span class="lang-badge">{sub.language}</span></td>
									<td><span class="status-badge {meta.cls}">{meta.label}</span></td>
									<td class="td-score">
										{sub.status === 'pending' || sub.status === 'running'
											? '—'
											: sub.score.toFixed(0) + '%'}
									</td>
									<td class="td-date">{formatDate(sub.submitted_at)}</td>
									<td>
										<a href="/submissions/{sub.id}" class="link-view">View</a>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</section>
	{/if}
</div>

<style>
	.page {
		max-width: 900px;
	}

	.page-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		margin-bottom: 2rem;
		gap: 1rem;
	}

	h1 {
		font-size: 1.75rem;
		font-weight: 700;
	}

	.greeting {
		color: #6b7280;
		font-size: 0.9rem;
		margin-top: 0.25rem;
	}

	/* ── Stat cards ── */
	.stat-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
		gap: 1rem;
		margin-bottom: 2.5rem;
	}

	.stat-card {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		padding: 1.25rem 1.4rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		transition: box-shadow 0.15s;
	}

	.stat-card:hover {
		box-shadow: 0 4px 12px rgba(0,0,0,0.07);
	}

	.stat-card.accent {
		border-color: #a5b4fc;
		background: linear-gradient(135deg, #eef2ff 0%, #fff 100%);
	}

	.stat-icon {
		font-size: 1.6rem;
		line-height: 1;
		flex-shrink: 0;
	}

	.stat-body {
		min-width: 0;
	}

	.stat-value {
		font-size: 1.6rem;
		font-weight: 700;
		color: #1a1a2e;
		line-height: 1.1;
	}

	.stat-label {
		font-size: 0.8rem;
		color: #6b7280;
		margin-top: 0.2rem;
	}

	.sub {
		font-size: 0.75rem;
		color: #9ca3af;
		display: block;
	}

	/* ── Recent submissions ── */
	.recent-section h2 {
		font-size: 1.1rem;
		font-weight: 600;
		margin-bottom: 1rem;
	}

	.table-wrap {
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		overflow: hidden;
	}

	.sub-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.sub-table th {
		text-align: left;
		padding: 0.65rem 1rem;
		font-size: 0.75rem;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
	}

	.sub-table td {
		padding: 0.75rem 1rem;
		border-bottom: 1px solid #f3f4f6;
		color: #374151;
		vertical-align: middle;
	}

	.sub-table tr:last-child td {
		border-bottom: none;
	}

	.sub-table tr:hover td {
		background: #fafafa;
	}

	.td-exercise a {
		font-weight: 500;
		color: #1a1a2e;
	}

	.td-exercise a:hover {
		color: #4f46e5;
	}

	.td-score {
		font-weight: 600;
		font-variant-numeric: tabular-nums;
	}

	.td-date {
		color: #9ca3af;
		font-size: 0.82rem;
		white-space: nowrap;
	}

	.lang-badge {
		background: #e0e7ff;
		color: #3730a3;
		font-size: 0.72rem;
		font-weight: 600;
		padding: 2px 7px;
		border-radius: 99px;
		text-transform: capitalize;
	}

	.status-badge {
		font-size: 0.72rem;
		font-weight: 600;
		padding: 3px 8px;
		border-radius: 99px;
		white-space: nowrap;
	}

	.status-accepted { background: #dcfce7; color: #166534; }
	.status-pending  { background: #f3f4f6; color: #6b7280; }
	.status-running  { background: #fef9c3; color: #a16207; }
	.status-wrong    { background: #fef2f2; color: #b91c1c; }
	.status-error    { background: #fef2f2; color: #b91c1c; }

	.link-view {
		font-size: 0.8rem;
		color: #4f46e5;
		font-weight: 500;
	}

	.link-view:hover {
		text-decoration: underline;
	}

	/* ── Empty state ── */
	.empty {
		text-align: center;
		padding: 3rem 2rem;
		color: #6b7280;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
	}

	/* ── Skeleton ── */
	.skeleton {
		height: 96px;
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
	}

	.skeleton-table {
		height: 200px;
		border-radius: 10px;
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
	}

	@keyframes shimmer {
		0%   { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}

	/* ── Buttons ── */
	.btn-primary {
		background: #4f46e5;
		color: #fff;
		border: none;
		border-radius: 8px;
		padding: 0.55rem 1.2rem;
		font-size: 0.9rem;
		font-weight: 600;
		cursor: pointer;
		text-decoration: none;
		display: inline-block;
	}

	.btn-primary:hover {
		background: #4338ca;
		text-decoration: none;
	}

	.btn-outline {
		background: transparent;
		color: #374151;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
		font-weight: 500;
		text-decoration: none;
		white-space: nowrap;
		flex-shrink: 0;
	}

	.btn-outline:hover {
		background: #f9fafb;
		text-decoration: none;
	}

	.alert {
		background: #fef2f2;
		color: #b91c1c;
		border: 1px solid #fecaca;
		border-radius: 8px;
		padding: 0.75rem 1rem;
		font-size: 0.9rem;
	}
</style>
