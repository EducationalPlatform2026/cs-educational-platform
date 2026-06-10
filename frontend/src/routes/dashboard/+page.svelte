<script lang="ts">
	import { onMount } from 'svelte';
	import { getDashboard, type DashboardStats } from '$lib/api/dashboard';
	import { listCourses, type Course } from '$lib/api/courses';
	import { getStatusMeta } from '$lib/utils/submissionStatus';
	import { auth } from '$lib/stores/auth.svelte';
	import { userStore } from '$lib/stores/userStore.svelte';
	import StatChip from '$lib/components/StatChip.svelte';
	import XpBar from '$lib/components/XpBar.svelte';

	let stats = $state<DashboardStats | null>(null);
	let courses = $state<Course[]>([]);
	let loading = $state(true);
	let error = $state('');

	const role = $derived(auth.user?.role ?? '');
	const isStudent = $derived(role === 'student' || role === 'teaching_assistant');
	const isProfessor = $derived(role === 'professor');

	const enrolledCourses = $derived(courses.filter((c) => c.is_enrolled));
	const myCourses = $derived(courses.filter((c) => c.created_by === auth.user?.user_id));

	const STRIP_COLORS = ['#7c3aed', '#0d9488', '#d97706', '#e11d48', '#2563eb'];

	function stripColor(i: number) {
		return STRIP_COLORS[i % STRIP_COLORS.length];
	}

	function timeAgo(iso: string) {
		const diff = Date.now() - new Date(iso).getTime();
		const m = Math.floor(diff / 60000);
		if (m < 1) return 'just now';
		if (m < 60) return `${m}m ago`;
		const h = Math.floor(m / 60);
		if (h < 24) return `${h}h ago`;
		return `${Math.floor(h / 24)}d ago`;
	}

	onMount(async () => {
		try {
			if (isStudent) {
				const [s, c] = await Promise.all([getDashboard(), listCourses()]);
				stats = s;
				courses = c;
				userStore.rehydrateFromDashboard(s);
			} else {
				courses = await listCourses();
			}
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load dashboard';
		} finally {
			loading = false;
		}
	});
</script>

<div class="page">
	<!-- Greeting -->
	<div class="greeting-row">
		<div>
			<h1>Welcome back, {auth.user?.first_name ?? 'there'} 👋</h1>
			<p class="greeting-sub">
				{#if role === 'professor'}Professor dashboard{:else if role === 'teaching_assistant'}Teaching assistant dashboard{:else if role === 'admin'}Admin dashboard{:else}Your learning progress at a glance{/if}
			</p>
		</div>
		{#if isStudent}
			<div class="level-badge">
				<span class="level-num">Lv {userStore.level}</span>
				<XpBar />
			</div>
		{/if}
	</div>

	{#if loading}
		<div class="skeleton-grid">
			{#each [1,2,3,4] as _}
				<div class="skeleton-chip"></div>
			{/each}
		</div>
		<div class="skeleton-block" style="height:200px"></div>
		<div class="skeleton-block" style="height:300px;margin-top:1rem"></div>

	{:else if error}
		<div class="alert">{error}</div>

	{:else if isStudent && stats}
		<!-- ── Student view ─────────────────────────── -->
		<div class="chips-grid">
			<StatChip label="Enrolled courses"  value={stats.enrolled_courses}       icon="📚" color="brand" />
			<StatChip label="Exercises solved"  value={stats.accepted_submissions}   icon="✅" color="green" />
			<StatChip label="Total submissions" value={stats.total_submissions}      icon="📨" color="blue" />
			<StatChip
				label="Acceptance rate"
				value={stats.total_submissions === 0 ? '—' : stats.acceptance_rate.toFixed(0) + '%'}
				icon="🎯"
				color="xp"
			/>
		</div>

		<!-- Streak banner (only if streak > 0) -->
		{#if userStore.streak > 0}
			<div class="streak-banner">
				🔥 <strong>{userStore.streak}-day streak!</strong> Keep solving problems to maintain it.
			</div>
		{/if}

		<!-- Continue Learning -->
		{#if enrolledCourses.length > 0}
			<section class="section">
				<div class="section-hd">
					<h2>Continue learning</h2>
					<a href="/courses" class="see-all">Browse all →</a>
				</div>
				<div class="course-row">
					{#each enrolledCourses.slice(0, 4) as c, i (c.id)}
						<a href="/courses/{c.id}" class="course-pill">
							<span class="pill-strip" style="background:{stripColor(i)}"></span>
							<span class="pill-title">{c.title}</span>
							<span class="pill-arrow">→</span>
						</a>
					{/each}
				</div>
			</section>
		{:else}
			<div class="enroll-cta">
				<span class="enroll-icon">🚀</span>
				<p class="enroll-title">Start your first course</p>
				<p class="enroll-sub">Enroll in a course to start solving exercises and earning XP.</p>
				<a href="/courses" class="btn-primary">Browse courses</a>
			</div>
		{/if}

		<!-- Recent submissions -->
		<section class="section">
			<h2>Recent submissions</h2>
			{#if stats.recent_submissions.length === 0}
				<p class="empty-text">No submissions yet. <a href="/courses">Find an exercise to try.</a></p>
			{:else}
				<div class="table-wrap">
					<table class="sub-table">
						<thead>
							<tr>
								<th>Exercise</th>
								<th>Language</th>
								<th>Status</th>
								<th>Score</th>
								<th>When</th>
							</tr>
						</thead>
						<tbody>
							{#each stats.recent_submissions as sub (sub.id)}
								{@const meta = getStatusMeta(sub.status)}
								<tr>
									<td><a href="/exercises/{sub.exercise_id}" class="ex-link">{sub.exercise_title}</a></td>
									<td><span class="lang-tag">{sub.language}</span></td>
									<td><span class="status-badge {meta.cls}">{meta.label}</span></td>
									<td class="score-cell">
										{sub.status === 'pending' || sub.status === 'running' ? '—' : sub.score.toFixed(0) + '%'}
									</td>
									<td class="date-cell">{timeAgo(sub.submitted_at)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</section>

	{:else if isProfessor}
		<!-- ── Professor view ──────────────────────── -->
		<div class="chips-grid">
			<StatChip label="Courses created"  value={myCourses.length}                          icon="📚" color="brand" />
			<StatChip label="Total courses"    value={courses.filter(c=>c.is_published).length}  icon="🌐" color="blue" />
		</div>

		{#if myCourses.length === 0}
			<div class="enroll-cta">
				<span class="enroll-icon">✏️</span>
				<p class="enroll-title">Create your first course</p>
				<p class="enroll-sub">Build a course, add exercises, and track student progress.</p>
				<a href="/courses" class="btn-primary">Go to courses</a>
			</div>
		{:else}
			<section class="section">
				<div class="section-hd">
					<h2>Your courses</h2>
					<a href="/courses" class="see-all">View all →</a>
				</div>
				<div class="prof-grid">
					{#each myCourses as c, i (c.id)}
						<div class="prof-card">
							<div class="prof-strip" style="background:{stripColor(i)}"></div>
							<div class="prof-body">
								<div class="prof-title-row">
									<a href="/courses/{c.id}" class="prof-title">{c.title}</a>
									{#if !c.is_published}<span class="badge-draft">Draft</span>{/if}
								</div>
								{#if c.description}<p class="prof-desc">{c.description}</p>{/if}
								<div class="prof-actions">
									<a href="/courses/{c.id}/stats" class="btn-sm btn-outline">📊 Stats</a>
									<a href="/courses/{c.id}/exercises/new" class="btn-sm">+ Exercise</a>
									<a href="/courses/{c.id}/edit" class="btn-sm btn-outline">Edit</a>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</section>
		{/if}

	{:else}
		<!-- ── Admin / TA fallback ─────────────────── -->
		<div class="chips-grid">
			<StatChip label="Total courses" value={courses.length} icon="📚" color="brand" />
		</div>
		<div class="section">
			<div class="section-hd"><h2>All courses</h2><a href="/courses" class="see-all">View all →</a></div>
			<p class="empty-text">Browse all courses to manage content.</p>
		</div>
	{/if}
</div>

<style>
	.page { max-width: 960px; }

	/* ── Greeting ── */
	.greeting-row {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.75rem;
	}
	h1 { font-size: 1.75rem; font-weight: 700; }
	.greeting-sub { font-size: 0.875rem; color: #6b7280; margin-top: 0.2rem; }

	.level-badge {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 5px;
		padding: 0.75rem 1rem;
		background: #fff;
		border: 1px solid #ddd6fe;
		border-radius: 12px;
		flex-shrink: 0;
	}
	.level-num { font-size: 0.78rem; font-weight: 700; color: #7c3aed; }

	/* ── Stat chips ── */
	.chips-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		gap: 0.85rem;
		margin-bottom: 1.75rem;
	}

	/* ── Streak banner ── */
	.streak-banner {
		background: #fce7f3;
		border: 1px solid #fbcfe8;
		color: #9d174d;
		border-radius: 10px;
		padding: 0.65rem 1rem;
		font-size: 0.875rem;
		margin-bottom: 1.75rem;
	}

	/* ── Sections ── */
	.section { margin-bottom: 2rem; }
	.section-hd {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
	}
	h2 { font-size: 1.1rem; font-weight: 600; }
	.see-all { font-size: 0.85rem; color: #7c3aed; font-weight: 500; }

	/* ── Course pill row ── */
	.course-row { display: flex; flex-direction: column; gap: 0.5rem; }
	.course-pill {
		display: flex;
		align-items: center;
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		overflow: hidden;
		text-decoration: none;
		transition: all 0.15s;
	}
	.course-pill:hover { border-color: #a78bfa; box-shadow: 0 2px 8px rgba(124,58,237,0.1); text-decoration: none; }
	.pill-strip { width: 6px; min-height: 48px; flex-shrink: 0; }
	.pill-title { flex: 1; padding: 0.75rem 1rem; font-size: 0.9rem; font-weight: 500; color: #374151; }
	.pill-arrow { padding-right: 1rem; color: #a78bfa; font-size: 1rem; }

	/* ── Enroll CTA ── */
	.enroll-cta {
		text-align: center;
		padding: 3rem 2rem;
		background: #fff;
		border: 1px dashed #ddd6fe;
		border-radius: 14px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.6rem;
		margin-bottom: 1.75rem;
	}
	.enroll-icon { font-size: 2.5rem; line-height: 1; }
	.enroll-title { font-size: 1.1rem; font-weight: 600; color: #374151; }
	.enroll-sub { font-size: 0.875rem; color: #6b7280; max-width: 360px; }

	/* ── Professor cards ── */
	.prof-grid { display: flex; flex-direction: column; gap: 0.75rem; }
	.prof-card {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		overflow: hidden;
		display: flex;
		transition: box-shadow 0.15s;
	}
	.prof-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.07); }
	.prof-strip { width: 6px; flex-shrink: 0; }
	.prof-body { padding: 1rem 1.25rem; flex: 1; }
	.prof-title-row { display: flex; align-items: center; gap: 0.6rem; margin-bottom: 0.4rem; }
	.prof-title { font-size: 1rem; font-weight: 600; color: #1a1a2e; }
	.prof-title:hover { color: #7c3aed; }
	.badge-draft { background: #fef9c3; color: #a16207; font-size: 0.7rem; font-weight: 700; padding: 2px 7px; border-radius: 99px; }
	.prof-desc { color: #6b7280; font-size: 0.85rem; line-height: 1.5; margin-bottom: 0.75rem; display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
	.prof-actions { display: flex; gap: 0.5rem; }

	/* ── Submissions table ── */
	.table-wrap { border: 1px solid #e5e7eb; border-radius: 10px; overflow: hidden; }
	.sub-table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
	.sub-table th { text-align: left; padding: 0.6rem 1rem; font-size: 0.73rem; font-weight: 600; color: #6b7280; text-transform: uppercase; letter-spacing: 0.04em; background: #f9fafb; border-bottom: 1px solid #e5e7eb; }
	.sub-table td { padding: 0.7rem 1rem; border-bottom: 1px solid #f3f4f6; color: #374151; vertical-align: middle; }
	.sub-table tr:last-child td { border-bottom: none; }
	.sub-table tr:hover td { background: #fafafa; }
	.ex-link { font-weight: 500; color: #1a1a2e; }
	.ex-link:hover { color: #7c3aed; }
	.score-cell { font-weight: 600; font-variant-numeric: tabular-nums; }
	.date-cell { color: #9ca3af; font-size: 0.82rem; white-space: nowrap; }

	/* ── Badges ── */
	.lang-tag { background: #ede9fe; color: #5b21b6; font-size: 0.72rem; font-weight: 600; padding: 2px 7px; border-radius: 99px; }
	.status-badge { font-size: 0.72rem; font-weight: 600; padding: 3px 8px; border-radius: 99px; white-space: nowrap; }
	:global(.status-accepted) { background: #dcfce7; color: #166534; }
	:global(.status-pending)  { background: #f3f4f6; color: #6b7280; }
	:global(.status-running)  { background: #fef9c3; color: #a16207; }
	:global(.status-wrong)    { background: #fef2f2; color: #b91c1c; }
	:global(.status-error)    { background: #fef2f2; color: #b91c1c; }

	/* ── Buttons ── */
	.btn-primary {
		background: #7c3aed; color: #fff; border: none; border-radius: 8px;
		padding: 0.55rem 1.2rem; font-size: 0.9rem; font-weight: 600;
		cursor: pointer; text-decoration: none; display: inline-block;
	}
	.btn-primary:hover { background: #6d28d9; text-decoration: none; }

	.btn-sm {
		background: #7c3aed; color: #fff; border: none; border-radius: 6px;
		padding: 5px 12px; font-size: 0.8rem; font-weight: 600; cursor: pointer;
		text-decoration: none; display: inline-block;
	}
	.btn-sm:hover { background: #6d28d9; text-decoration: none; }
	.btn-sm.btn-outline { background: transparent; color: #374151; border: 1px solid #d1d5db; }
	.btn-sm.btn-outline:hover { background: #f9fafb; text-decoration: none; }

	/* ── Skeletons ── */
	.skeleton-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px,1fr)); gap: 0.85rem; margin-bottom: 1.75rem; }
	.skeleton-chip { height: 80px; border-radius: 14px; background: linear-gradient(90deg,#f0f0f0 25%,#e8e8e8 50%,#f0f0f0 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
	.skeleton-block { border-radius: 12px; background: linear-gradient(90deg,#f0f0f0 25%,#e8e8e8 50%,#f0f0f0 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
	@keyframes shimmer { 0%{background-position:200% 0} 100%{background-position:-200% 0} }

	.alert { background:#fef2f2; color:#b91c1c; border:1px solid #fecaca; border-radius:8px; padding:0.75rem 1rem; font-size:0.9rem; }
	.empty-text { color: #6b7280; font-size: 0.9rem; }
</style>
