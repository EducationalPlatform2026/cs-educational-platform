<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { getCourse, getCourseMembers, enrollCourse, type Course, type Member } from '$lib/api/courses';
	import { listExercises, deleteExercise, type Exercise } from '$lib/api/exercises';
	import { auth } from '$lib/stores/auth.svelte';
	import { userStore } from '$lib/stores/userStore.svelte';
	import LearningPathNode from '$lib/components/LearningPathNode.svelte';
	import DifficultyDot from '$lib/components/DifficultyDot.svelte';
	import ProgressRing from '$lib/components/ProgressRing.svelte';

	const id = $derived($page.params.id as string);

	let course = $state<Course | null>(null);
	let members = $state<Member[]>([]);
	let exList = $state<Exercise[]>([]);
	let loading = $state(true);
	let error = $state('');
	let enrolling = $state(false);
	let enrolled = $state(false);
	let deletingEx = $state<Record<string, boolean>>({});

	const canManage = $derived(auth.user?.role === 'professor' || auth.user?.role === 'admin');
	const canManageExercises = $derived(
		canManage || (auth.user?.role === 'teaching_assistant' && enrolled)
	);
	const canEnroll = $derived(auth.user?.role === 'student' || auth.user?.role === 'teaching_assistant');
	const canSeeMembers = $derived(
		auth.user?.role === 'professor' || auth.user?.role === 'teaching_assistant' || auth.user?.role === 'admin'
	);
	const canAccess = $derived(enrolled || canManage);

	const solvedCount = $derived(exList.filter((e) => userStore.isSolved(e.id)).length);
	const progressPct = $derived(exList.length > 0 ? Math.round((solvedCount / exList.length) * 100) : 0);

	const STRIP_COLORS = ['#7c3aed', '#0d9488', '#d97706', '#e11d48', '#2563eb'];

	function nodeState(ex: Exercise, i: number): 'done' | 'current' | 'locked' {
		if (userStore.isSolved(ex.id)) return 'done';
		const firstUnsolved = exList.findIndex((e) => !userStore.isSolved(e.id));
		if (i === firstUnsolved) return 'current';
		return 'locked';
	}

	function xpFor(difficulty: string) {
		return difficulty === 'hard' ? 200 : 100;
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' });
	}

	function roleLabel(role: string) { return role.replace('_', ' '); }

	onMount(async () => {
		try {
			const [c, m, ex] = await Promise.all([
				getCourse(id),
				canSeeMembers ? getCourseMembers(id) : Promise.resolve([]),
				listExercises(id)
			]);
			course = c;
			enrolled = c.is_enrolled;
			members = m;
			exList = ex;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load course';
		} finally {
			loading = false;
		}
	});

	async function handleEnroll() {
		enrolling = true;
		try {
			await enrollCourse(id);
			enrolled = true;
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Enroll failed');
		} finally {
			enrolling = false;
		}
	}

	async function handleDeleteExercise(exId: string, title: string) {
		if (!confirm(`Delete exercise "${title}"?`)) return;
		deletingEx = { ...deletingEx, [exId]: true };
		try {
			await deleteExercise(exId);
			exList = exList.filter((e) => e.id !== exId);
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Delete failed');
		} finally {
			deletingEx = { ...deletingEx, [exId]: false };
		}
	}
</script>

<div class="page">
	<a href="/courses" class="back-link">← All courses</a>

	{#if loading}
		<div class="skeleton-header"></div>

	{:else if error}
		<div class="alert">{error}</div>

	{:else if course}
		<!-- ── Course header ───────────────── -->
		<div class="course-hero">
			<div class="hero-strip" style="background:{STRIP_COLORS[0]}"></div>
			<div class="hero-body">
				<div class="hero-main">
					<div>
						<div class="hero-title-row">
							<h1>{course.title}</h1>
							{#if !course.is_published}<span class="badge-draft">Draft</span>{/if}
						</div>
						{#if course.description}<p class="hero-desc">{course.description}</p>{/if}
						<p class="hero-meta">Created {formatDate(course.created_at)}</p>
					</div>

					<!-- Progress ring (only when enrolled/managing) -->
					{#if canAccess && exList.length > 0}
						<div class="progress-wrap">
							<ProgressRing percent={progressPct} size={72} label="{progressPct}%" />
							<div class="progress-label">
								<span class="prog-num">{solvedCount}/{exList.length}</span>
								<span class="prog-sub">solved</span>
							</div>
						</div>
					{/if}
				</div>

				<div class="hero-actions">
					{#if canEnroll && course.is_published}
						{#if enrolled}
							<span class="enrolled-badge">✓ Enrolled</span>
						{:else}
							<button class="btn-primary" onclick={handleEnroll} disabled={enrolling}>
								{enrolling ? 'Enrolling…' : 'Enroll now'}
							</button>
						{/if}
					{/if}
					{#if canSeeMembers}
						<a href="/courses/{course.id}/stats" class="btn-outline">📊 Stats</a>
					{/if}
					{#if canManage}
						<a href="/courses/{course.id}/edit" class="btn-outline">Edit</a>
					{/if}
				</div>
			</div>
		</div>

		<!-- ── Learning path ──────────────── -->
		<section class="section">
			<div class="section-hd">
				<h2>Learning path
					{#if canAccess}({exList.length} exercises){/if}
				</h2>
				{#if canManageExercises}
					<a href="/courses/{id}/exercises/new" class="btn-sm">+ Add exercise</a>
				{/if}
			</div>

			{#if !canManageExercises && !enrolled && canEnroll}
				<div class="enroll-gate">
					<div class="gate-icon">🔒</div>
					<p class="gate-title">Enroll to unlock exercises</p>
					<p class="gate-sub">Join this course to access and solve exercises, track your progress, and earn XP.</p>
					{#if course.is_published}
						<button class="btn-primary" onclick={handleEnroll} disabled={enrolling}>
							{enrolling ? 'Enrolling…' : 'Enroll now — it\'s free'}
						</button>
					{/if}
				</div>

			{:else if exList.length === 0}
				<p class="empty-text">
					{canManageExercises ? 'No exercises yet. Add the first one above.' : 'No exercises available yet.'}
				</p>

			{:else}
				<div class="path">
					{#each exList as ex, i (ex.id)}
						<div class="path-step">
							<LearningPathNode
								state={canAccess ? nodeState(ex, i) : 'locked'}
								label={ex.title}
								sub="{ex.difficulty} · {ex.language} · +{xpFor(ex.difficulty)} XP"
								href={canAccess ? `/exercises/${ex.id}` : ''}
							/>
							{#if canManageExercises}
								<div class="ex-mgmt">
									<a href="/exercises/{ex.id}/edit" class="btn-xs btn-outline">Edit</a>
									<button
										class="btn-xs btn-danger"
										onclick={() => handleDeleteExercise(ex.id, ex.title)}
										disabled={deletingEx[ex.id]}
									>Del</button>
								</div>
							{/if}
						</div>
						{#if i < exList.length - 1}
							<div class="path-connector" class:done={userStore.isSolved(ex.id)}></div>
						{/if}
					{/each}
				</div>
			{/if}
		</section>

		<!-- ── Members (privileged) ─────── -->
		{#if canSeeMembers && members.length > 0}
			<section class="section">
				<h2>Enrolled members ({members.length})</h2>
				<div class="members-wrap">
					<table class="members-table">
						<thead>
							<tr><th>Name</th><th>Role</th><th>Enrolled</th></tr>
						</thead>
						<tbody>
							{#each members as m (m.user_id)}
								<tr>
									<td>{m.first_name} {m.last_name}</td>
									<td><span class="role-badge">{roleLabel(m.role)}</span></td>
									<td>{formatDate(m.enrolled_at)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>
		{/if}
	{/if}
</div>

<style>
	.page { max-width: 860px; }

	.back-link { display:inline-block; font-size:0.875rem; color:#6b7280; margin-bottom:1.5rem; }
	.back-link:hover { color:#7c3aed; }

	.skeleton-header { height:120px; border-radius:14px; background:linear-gradient(90deg,#f0f0f0 25%,#e8e8e8 50%,#f0f0f0 75%); background-size:200% 100%; animation:shimmer 1.4s infinite; }
	@keyframes shimmer { 0%{background-position:200% 0} 100%{background-position:-200% 0} }

	.alert { background:#fef2f2; color:#b91c1c; border:1px solid #fecaca; border-radius:8px; padding:0.75rem 1rem; font-size:0.9rem; }

	/* ── Hero ── */
	.course-hero {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 14px;
		overflow: hidden;
		margin-bottom: 2rem;
	}
	.hero-strip { height: 6px; }
	.hero-body { padding: 1.5rem; }
	.hero-main { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; margin-bottom: 1.25rem; }
	.hero-title-row { display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; margin-bottom: 0.4rem; }
	h1 { font-size: 1.65rem; font-weight: 700; line-height: 1.3; }
	.badge-draft { background:#fef9c3; color:#a16207; font-size:0.7rem; font-weight:700; padding:2px 7px; border-radius:99px; }
	.hero-desc { color:#6b7280; font-size:0.9rem; line-height:1.6; margin-bottom:0.4rem; }
	.hero-meta { color:#9ca3af; font-size:0.82rem; }

	.progress-wrap { display:flex; align-items:center; gap:0.75rem; flex-shrink:0; }
	.progress-label { display:flex; flex-direction:column; }
	.prog-num { font-size:1rem; font-weight:700; color:#7c3aed; }
	.prog-sub { font-size:0.72rem; color:#9ca3af; }

	.hero-actions { display:flex; gap:0.6rem; align-items:center; flex-wrap:wrap; }
	.enrolled-badge { font-size:0.9rem; color:#16a34a; font-weight:600; }

	/* ── Section ── */
	.section { margin-bottom:2.5rem; }
	.section-hd { display:flex; align-items:center; justify-content:space-between; margin-bottom:1.25rem; }
	h2 { font-size:1.1rem; font-weight:600; }

	/* ── Learning path ── */
	.path { display:flex; flex-direction:column; }

	.path-step {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 6px 0;
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		padding: 0.75rem 1rem;
	}
	.path-step:hover { border-color: #ddd6fe; background: #faf8ff; }

	.path-connector {
		margin-left: 15px;
		width: 2px;
		height: 16px;
		background: repeating-linear-gradient(to bottom, #ddd6fe, #ddd6fe 4px, transparent 4px, transparent 8px);
	}
	.path-connector.done {
		background: repeating-linear-gradient(to bottom, #a78bfa, #a78bfa 4px, transparent 4px, transparent 8px);
	}

	.ex-mgmt { display:flex; gap:0.4rem; margin-left:auto; }

	/* ── Enroll gate ── */
	.enroll-gate {
		display:flex; flex-direction:column; align-items:center;
		gap:0.75rem; padding:3rem 2rem;
		border:2px dashed #ddd6fe; border-radius:14px;
		text-align:center; background:#faf8ff;
	}
	.gate-icon { font-size:2.5rem; line-height:1; }
	.gate-title { font-size:1.1rem; font-weight:600; color:#374151; }
	.gate-sub { font-size:0.875rem; color:#6b7280; max-width:380px; }

	/* ── Members ── */
	.members-wrap { overflow-x:auto; border:1px solid #e5e7eb; border-radius:10px; }
	.members-table { width:100%; border-collapse:collapse; font-size:0.875rem; }
	.members-table th { text-align:left; padding:0.65rem 1rem; font-size:0.75rem; font-weight:600; color:#6b7280; text-transform:uppercase; letter-spacing:0.04em; background:#f9fafb; border-bottom:1px solid #e5e7eb; }
	.members-table td { padding:0.75rem 1rem; border-bottom:1px solid #f3f4f6; color:#374151; }
	.members-table tr:last-child td { border-bottom:none; }
	.role-badge { background:#ede9fe; color:#7c3aed; font-size:0.75rem; font-weight:600; padding:2px 8px; border-radius:99px; text-transform:capitalize; }

	/* ── Buttons ── */
	.btn-primary { background:#7c3aed; color:#fff; border:none; border-radius:8px; padding:0.55rem 1.1rem; font-size:0.9rem; font-weight:600; cursor:pointer; transition:background 0.15s; font-family:inherit; }
	.btn-primary:hover:not(:disabled) { background:#6d28d9; }
	.btn-primary:disabled { opacity:0.6; cursor:not-allowed; }
	.btn-outline { background:transparent; color:#374151; border:1px solid #d1d5db; border-radius:8px; padding:0.5rem 1rem; font-size:0.875rem; font-weight:500; cursor:pointer; text-decoration:none; transition:all 0.15s; font-family:inherit; }
	.btn-outline:hover { background:#f9fafb; text-decoration:none; }
	.btn-sm { background:#7c3aed; color:#fff; border:none; border-radius:7px; padding:5px 12px; font-size:0.82rem; font-weight:600; cursor:pointer; text-decoration:none; }
	.btn-sm:hover { background:#6d28d9; text-decoration:none; }
	.btn-xs { font-size:0.78rem; font-weight:600; padding:3px 9px; border-radius:5px; border:none; cursor:pointer; text-decoration:none; display:inline-block; font-family:inherit; }
	.btn-xs.btn-outline { background:transparent; color:#374151; border:1px solid #d1d5db; }
	.btn-xs.btn-outline:hover { background:#f9fafb; }
	.btn-xs.btn-danger { background:#ef4444; color:#fff; }
	.btn-xs.btn-danger:hover:not(:disabled) { background:#dc2626; }
	.btn-xs:disabled { opacity:0.5; cursor:not-allowed; }

	.empty-text { color:#6b7280; font-size:0.9rem; }
</style>
