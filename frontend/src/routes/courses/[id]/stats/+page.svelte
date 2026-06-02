<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getCourseStats, type CourseStats, type StudentStat } from '$lib/api/coursestats';
	import { auth } from '$lib/stores/auth.svelte';

	const courseId = $derived($page.params.id as string);

	let stats = $state<CourseStats | null>(null);
	let loading = $state(true);
	let error = $state('');

	// Sorting
	type SortKey = 'name' | 'solved' | 'submissions' | 'rate' | 'last_active';
	let sortKey = $state<SortKey>('solved');
	let sortAsc = $state(false);

	const canView = $derived(
		auth.user?.role === 'professor' ||
		auth.user?.role === 'teaching_assistant' ||
		auth.user?.role === 'admin'
	);

	const sortedStudents = $derived(
		stats
			? [...stats.students].sort((a, b) => {
					let va: number | string = 0;
					let vb: number | string = 0;
					if (sortKey === 'name')        { va = `${a.last_name} ${a.first_name}`; vb = `${b.last_name} ${b.first_name}`; }
					if (sortKey === 'solved')       { va = a.exercises_solved; vb = b.exercises_solved; }
					if (sortKey === 'submissions')  { va = a.total_submissions; vb = b.total_submissions; }
					if (sortKey === 'rate')         { va = a.acceptance_rate; vb = b.acceptance_rate; }
					if (sortKey === 'last_active')  {
						va = a.last_submission_at ? new Date(a.last_submission_at).getTime() : 0;
						vb = b.last_submission_at ? new Date(b.last_submission_at).getTime() : 0;
					}
					const cmp = typeof va === 'string' ? va.localeCompare(vb as string) : (va as number) - (vb as number);
					return sortAsc ? cmp : -cmp;
			  })
			: []
	);

	function toggleSort(key: SortKey) {
		if (sortKey === key) sortAsc = !sortAsc;
		else { sortKey = key; sortAsc = false; }
	}

	function sortIcon(key: SortKey) {
		if (sortKey !== key) return '↕';
		return sortAsc ? '↑' : '↓';
	}

	function diffColor(d: string) {
		if (d === 'easy') return '#16a34a';
		if (d === 'hard') return '#b91c1c';
		return '#a16207';
	}

	function fmtDate(iso?: string) {
		if (!iso) return '—';
		return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
	}

	function fmtDateTime(iso?: string) {
		if (!iso) return 'Never';
		return new Date(iso).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	onMount(async () => {
		if (!canView) { goto(`/courses/${courseId}`); return; }
		try {
			stats = await getCourseStats(courseId);
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load stats';
		} finally {
			loading = false;
		}
	});
</script>

<div class="page">
	<a href="/courses/{courseId}" class="back-link">← Back to course</a>

	{#if loading}
		<div class="skeleton-title"></div>
		<div class="stat-grid">
			{#each [1,2,3,4,5,6] as _}<div class="stat-card skeleton"></div>{/each}
		</div>

	{:else if error}
		<div class="alert">{error}</div>

	{:else if stats}
		<div class="page-header">
			<div>
				<h1>Course Report</h1>
				<p class="subtitle">Statistical overview for professors and lab attendants</p>
			</div>
		</div>

		<!-- ── Summary cards ──────────────────────────────────────────────── -->
		<div class="stat-grid">
			<div class="stat-card">
				<span class="stat-icon">👨‍🎓</span>
				<div>
					<div class="stat-val">{stats.summary.total_students}</div>
					<div class="stat-lbl">Students enrolled</div>
				</div>
			</div>
			<div class="stat-card">
				<span class="stat-icon">🧪</span>
				<div>
					<div class="stat-val">{stats.summary.total_teaching_assistants}</div>
					<div class="stat-lbl">Lab attendants</div>
				</div>
			</div>
			<div class="stat-card">
				<span class="stat-icon">💡</span>
				<div>
					<div class="stat-val">{stats.summary.total_exercises}</div>
					<div class="stat-lbl">Published exercises</div>
				</div>
			</div>
			<div class="stat-card">
				<span class="stat-icon">📨</span>
				<div>
					<div class="stat-val">{stats.summary.total_submissions}</div>
					<div class="stat-lbl">Total submissions</div>
				</div>
			</div>
			<div class="stat-card accent-green">
				<span class="stat-icon">✅</span>
				<div>
					<div class="stat-val">
						{stats.summary.total_submissions === 0 ? '—' : stats.summary.course_acceptance_rate.toFixed(1) + '%'}
					</div>
					<div class="stat-lbl">Course acceptance rate</div>
				</div>
			</div>
			<div class="stat-card accent-blue">
				<span class="stat-icon">🏃</span>
				<div>
					<div class="stat-val">
						{stats.summary.total_students === 0 ? '—'
							: Math.round(stats.summary.students_with_submission / stats.summary.total_students * 100) + '%'}
					</div>
					<div class="stat-lbl">Students active</div>
					<div class="stat-sub">{stats.summary.students_with_submission} / {stats.summary.total_students}</div>
				</div>
			</div>
		</div>

		<!-- ── Staff ──────────────────────────────────────────────────────── -->
		<section class="section">
			<h2>Course staff</h2>
			<div class="staff-grid">
				<div class="staff-card professor">
					<span class="staff-role-badge">Professor</span>
					<p class="staff-name">{stats.professor.first_name} {stats.professor.last_name}</p>
					<p class="staff-email">{stats.professor.email}</p>
				</div>
				{#each stats.teaching_assistants as ta (ta.user_id)}
					<div class="staff-card ta">
						<span class="staff-role-badge ta-badge">Lab Attendant</span>
						<p class="staff-name">{ta.first_name} {ta.last_name}</p>
						<p class="staff-email">{ta.email}</p>
					</div>
				{/each}
				{#if stats.teaching_assistants.length === 0}
					<p class="empty-staff">No lab attendants assigned yet.</p>
				{/if}
			</div>
		</section>

		<!-- ── Exercise solve rates ───────────────────────────────────────── -->
		{#if stats.exercises.length > 0}
		<section class="section">
			<h2>Exercise solve rates</h2>
			<div class="ex-bars">
				{#each stats.exercises as ex (ex.id)}
					<div class="ex-bar-row">
						<div class="ex-bar-meta">
							<a href="/exercises/{ex.id}" class="ex-bar-title">{ex.title}</a>
							<div class="ex-bar-details">
								<span class="diff-dot" style="color:{diffColor(ex.difficulty)}">●</span>
								<span class="muted">{ex.difficulty}</span>
								<span class="muted">·</span>
								<span class="muted">{ex.students_solved}/{stats.summary.total_students} solved</span>
								<span class="muted">·</span>
								<span class="muted">{ex.total_submissions} submissions</span>
							</div>
						</div>
						<div class="bar-wrap">
							<div class="bar-track">
								<div
									class="bar-fill"
									style="width:{Math.min(ex.solve_rate, 100)}%; background:{ex.solve_rate >= 70 ? '#16a34a' : ex.solve_rate >= 40 ? '#ca8a04' : '#dc2626'}"
								></div>
							</div>
							<span class="bar-pct">{ex.solve_rate.toFixed(0)}%</span>
						</div>
					</div>
				{/each}
			</div>
		</section>
		{/if}

		<!-- ── Student roster ─────────────────────────────────────────────── -->
		<section class="section">
			<h2>Student roster ({stats.summary.total_students})</h2>

			{#if stats.students.length === 0}
				<p class="empty-text">No students enrolled yet.</p>
			{:else}
				<div class="table-wrap">
					<table class="roster-table">
						<thead>
							<tr>
								<th>
									<button class="sort-btn" onclick={() => toggleSort('name')}>
										Student {sortIcon('name')}
									</button>
								</th>
								<th>Enrolled</th>
								<th>
									<button class="sort-btn" onclick={() => toggleSort('solved')}>
										Solved {sortIcon('solved')}
									</button>
								</th>
								<th>
									<button class="sort-btn" onclick={() => toggleSort('submissions')}>
										Submissions {sortIcon('submissions')}
									</button>
								</th>
								<th>
									<button class="sort-btn" onclick={() => toggleSort('rate')}>
										Acceptance {sortIcon('rate')}
									</button>
								</th>
								<th>
									<button class="sort-btn" onclick={() => toggleSort('last_active')}>
										Last active {sortIcon('last_active')}
									</button>
								</th>
								<th>Progress</th>
							</tr>
						</thead>
						<tbody>
							{#each sortedStudents as s (s.user_id)}
								<tr>
									<td class="td-name">
										<span class="student-name">{s.first_name} {s.last_name}</span>
										<span class="student-email">{s.email}</span>
									</td>
									<td class="td-meta">{fmtDate(s.enrolled_at)}</td>
									<td class="td-center">
										<span class="solved-badge">{s.exercises_solved}/{stats.summary.total_exercises}</span>
									</td>
									<td class="td-center">{s.total_submissions}</td>
									<td class="td-center">
										{#if s.total_submissions === 0}
											<span class="muted">—</span>
										{:else}
											<span class="rate-chip" style="background:{s.acceptance_rate >= 70 ? '#dcfce7' : s.acceptance_rate >= 40 ? '#fef9c3' : '#fee2e2'}; color:{s.acceptance_rate >= 70 ? '#166534' : s.acceptance_rate >= 40 ? '#a16207' : '#b91c1c'}">
												{s.acceptance_rate.toFixed(0)}%
											</span>
										{/if}
									</td>
									<td class="td-meta">{fmtDateTime(s.last_submission_at)}</td>
									<td class="td-progress">
										<div class="mini-bar-track">
											<div class="mini-bar-fill" style="width:{stats.summary.total_exercises > 0 ? (s.exercises_solved / stats.summary.total_exercises * 100) : 0}%"></div>
										</div>
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
	.page { max-width: 1000px; }

	.back-link { display: inline-block; font-size: 0.875rem; color: #6b7280; margin-bottom: 1.5rem; }
	.back-link:hover { color: #4f46e5; }

	.page-header { margin-bottom: 1.75rem; }
	h1 { font-size: 1.75rem; font-weight: 700; }
	.subtitle { color: #6b7280; font-size: 0.875rem; margin-top: 0.25rem; }

	/* ── Summary cards ── */
	.stat-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
		gap: 1rem;
		margin-bottom: 2.5rem;
	}

	.stat-card {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		padding: 1.1rem 1.2rem;
		display: flex;
		align-items: center;
		gap: 0.9rem;
	}

	.stat-card.accent-green { border-color: #bbf7d0; background: linear-gradient(135deg,#f0fdf4,#fff); }
	.stat-card.accent-blue  { border-color: #bfdbfe; background: linear-gradient(135deg,#eff6ff,#fff); }

	.stat-icon { font-size: 1.5rem; flex-shrink: 0; }
	.stat-val  { font-size: 1.5rem; font-weight: 700; line-height: 1.1; }
	.stat-lbl  { font-size: 0.78rem; color: #6b7280; margin-top: 0.15rem; }
	.stat-sub  { font-size: 0.72rem; color: #9ca3af; }

	/* ── Sections ── */
	.section { margin-bottom: 2.5rem; }
	.section h2 { font-size: 1.05rem; font-weight: 600; margin-bottom: 1rem; }

	/* ── Staff ── */
	.staff-grid {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
	}

	.staff-card {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		padding: 0.9rem 1.1rem;
		min-width: 200px;
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
	}

	.staff-card.professor { border-color: #a5b4fc; background: #eef2ff; }
	.staff-card.ta        { border-color: #d1d5db; }

	.staff-role-badge {
		font-size: 0.7rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: #4f46e5;
		margin-bottom: 0.25rem;
	}

	.ta-badge { color: #6b7280; }

	.staff-name  { font-size: 0.95rem; font-weight: 600; color: #1a1a2e; }
	.staff-email { font-size: 0.8rem; color: #6b7280; }
	.empty-staff { font-size: 0.875rem; color: #9ca3af; align-self: center; }

	/* ── Exercise bars ── */
	.ex-bars { display: flex; flex-direction: column; gap: 0.9rem; }

	.ex-bar-row {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		padding: 0.85rem 1.1rem;
		display: flex;
		align-items: center;
		gap: 1.25rem;
	}

	.ex-bar-meta { flex: 1; min-width: 0; }

	.ex-bar-title {
		font-size: 0.9rem;
		font-weight: 500;
		color: #1a1a2e;
		display: block;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.ex-bar-title:hover { color: #4f46e5; }

	.ex-bar-details {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		margin-top: 0.2rem;
		font-size: 0.78rem;
	}

	.diff-dot { font-size: 0.6rem; }

	.bar-wrap {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		width: 220px;
		flex-shrink: 0;
	}

	.bar-track {
		flex: 1;
		height: 10px;
		background: #f3f4f6;
		border-radius: 99px;
		overflow: hidden;
	}

	.bar-fill {
		height: 100%;
		border-radius: 99px;
		transition: width 0.4s ease;
	}

	.bar-pct {
		font-size: 0.8rem;
		font-weight: 600;
		color: #374151;
		width: 2.5rem;
		text-align: right;
	}

	/* ── Roster table ── */
	.table-wrap {
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		overflow: hidden;
	}

	.roster-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.roster-table th {
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
		padding: 0.6rem 0.85rem;
		text-align: left;
		white-space: nowrap;
	}

	.sort-btn {
		background: none;
		border: none;
		cursor: pointer;
		font-size: 0.75rem;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		padding: 0;
		display: flex;
		align-items: center;
		gap: 0.3rem;
	}

	.sort-btn:hover { color: #4f46e5; }

	.roster-table td {
		padding: 0.75rem 0.85rem;
		border-bottom: 1px solid #f3f4f6;
		vertical-align: middle;
	}

	.roster-table tr:last-child td { border-bottom: none; }
	.roster-table tr:hover td { background: #fafafa; }

	.td-name { min-width: 160px; }
	.student-name  { display: block; font-weight: 500; color: #1a1a2e; font-size: 0.875rem; }
	.student-email { display: block; font-size: 0.75rem; color: #9ca3af; }

	.td-meta   { font-size: 0.8rem; color: #9ca3af; white-space: nowrap; }
	.td-center { text-align: center; }

	.solved-badge {
		font-size: 0.82rem;
		font-weight: 600;
		color: #374151;
	}

	.rate-chip {
		display: inline-block;
		font-size: 0.78rem;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 99px;
	}

	.td-progress { width: 80px; }

	.mini-bar-track {
		height: 6px;
		background: #f3f4f6;
		border-radius: 99px;
		overflow: hidden;
	}

	.mini-bar-fill {
		height: 100%;
		background: #4f46e5;
		border-radius: 99px;
	}

	/* ── Misc ── */
	.muted     { color: #9ca3af; }
	.empty-text { color: #9ca3af; font-size: 0.875rem; }

	.alert {
		background: #fef2f2; color: #b91c1c;
		border: 1px solid #fecaca;
		border-radius: 8px; padding: 0.75rem 1rem; font-size: 0.9rem;
	}

	/* ── Skeletons ── */
	.skeleton-title {
		height: 60px; border-radius: 8px; margin-bottom: 1.5rem;
		background: linear-gradient(90deg,#f0f0f0 25%,#e8e8e8 50%,#f0f0f0 75%);
		background-size: 200% 100%; animation: shimmer 1.4s infinite;
	}

	.skeleton {
		height: 80px;
		background: linear-gradient(90deg,#f0f0f0 25%,#e8e8e8 50%,#f0f0f0 75%);
		background-size: 200% 100%; animation: shimmer 1.4s infinite;
	}

	@keyframes shimmer {
		0%   { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}
</style>
