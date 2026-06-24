<script lang="ts">
	import { onMount } from 'svelte';
	import { listCourses, createCourse, deleteCourse, enrollCourse, type Course } from '$lib/api/courses';
	import { auth } from '$lib/stores/auth.svelte';
	import DifficultyDot from '$lib/components/DifficultyDot.svelte';

	let courses = $state<Course[]>([]);
	let loading = $state(true);
	let error = $state('');

	let showForm = $state(false);
	let newTitle = $state('');
	let newDesc = $state('');
	let newPublished = $state(true);
	let formError = $state('');
	let formLoading = $state(false);

	let filter = $state<'all' | 'enrolled' | 'available'>('all');
	let actionState = $state<Record<string, string>>({});

	const canManage = $derived(auth.user?.role === 'professor' || auth.user?.role === 'admin');
	const canEnroll  = $derived(auth.user?.role === 'student' || auth.user?.role === 'teaching_assistant');

	const STRIP_COLORS = ['#7c3aed', '#0d9488', '#d97706', '#e11d48', '#2563eb', '#059669'];
	function stripColor(i: number) { return STRIP_COLORS[i % STRIP_COLORS.length]; }

	const filtered = $derived(
		filter === 'enrolled'
			? courses.filter((c) => c.is_enrolled)
			: filter === 'available'
			? courses.filter((c) => !c.is_enrolled && c.is_published)
			: courses
	);

	onMount(async () => {
		await load();
	});

	async function load() {
		loading = true; error = '';
		try { courses = await listCourses(); }
		catch (err: unknown) { error = err instanceof Error ? err.message : 'Failed to load courses'; }
		finally { loading = false; }
	}

	async function handleCreate(e: SubmitEvent) {
		e.preventDefault(); formError = ''; formLoading = true;
		try {
			const course = await createCourse({ title: newTitle, description: newDesc || undefined, is_published: newPublished });
			courses = [course, ...courses];
			newTitle = ''; newDesc = ''; newPublished = true; showForm = false;
		} catch (err: unknown) {
			formError = err instanceof Error ? err.message : 'Failed to create course';
		} finally { formLoading = false; }
	}

	async function handleEnroll(id: string) {
		actionState = { ...actionState, [id]: 'enrolling' };
		try {
			await enrollCourse(id);
			courses = courses.map((c) => c.id === id ? { ...c, is_enrolled: true } : c);
			actionState = { ...actionState, [id]: 'done' };
		} catch (err: unknown) {
			actionState = { ...actionState, [id]: '' };
			alert(err instanceof Error ? err.message : 'Enroll failed');
		}
	}

	async function handleDelete(id: string, title: string) {
		if (!confirm(`Delete "${title}"? This cannot be undone.`)) return;
		actionState = { ...actionState, [id]: 'deleting' };
		try {
			await deleteCourse(id);
			courses = courses.filter((c) => c.id !== id);
		} catch (err: unknown) {
			actionState = { ...actionState, [id]: '' };
			alert(err instanceof Error ? err.message : 'Delete failed');
		}
	}
</script>

<div class="page">
	<div class="page-header">
		<div>
			<h1>Courses</h1>
			<p class="page-sub">
				{#if !loading}{courses.length} course{courses.length !== 1 ? 's' : ''} available{/if}
			</p>
		</div>
		{#if canManage}
			<button class="btn-primary" onclick={() => (showForm = !showForm)}>
				{showForm ? '✕ Cancel' : '+ New course'}
			</button>
		{/if}
	</div>

	<!-- Create course form -->
	{#if showForm}
		<div class="form-card">
			<h2>New course</h2>
			{#if formError}<div class="alert">{formError}</div>{/if}
			<form onsubmit={handleCreate}>
				<label><span>Title <span class="req">*</span></span>
					<input bind:value={newTitle} placeholder="Introduction to Algorithms" required />
				</label>
				<label>Description
					<textarea bind:value={newDesc} placeholder="What will students learn?" rows="3"></textarea>
				</label>
				<label class="check-label">
					<input type="checkbox" bind:checked={newPublished} />
					Publish immediately
				</label>
				<div class="form-actions">
					<button type="submit" class="btn-primary" disabled={formLoading}>
						{formLoading ? 'Creating…' : 'Create course'}
					</button>
				</div>
			</form>
		</div>
	{/if}

	{#if error}<div class="alert">{error}</div>{/if}

	<!-- Filter tabs -->
	{#if !loading && courses.length > 0 && canEnroll}
		<div class="filter-tabs">
			<button class:active={filter === 'all'}      onclick={() => (filter = 'all')}>All</button>
			<button class:active={filter === 'enrolled'} onclick={() => (filter = 'enrolled')}>Enrolled</button>
			<button class:active={filter === 'available'} onclick={() => (filter = 'available')}>Available</button>
		</div>
	{/if}

	{#if loading}
		<div class="grid">
			{#each [1,2,3] as _}<div class="course-card skeleton"></div>{/each}
		</div>
	{:else if filtered.length === 0}
		<div class="empty">
			{#if filter === 'enrolled'}
				<p>You haven't enrolled in any courses yet.</p>
				<button class="btn-outline" onclick={() => (filter = 'all')}>Browse all courses</button>
			{:else}
				<p>No courses yet.{canManage ? ' Create the first one above.' : ''}</p>
			{/if}
		</div>
	{:else}
		<div class="grid">
			{#each filtered as course, i (course.id)}
				{@const isEnrolled = course.is_enrolled || actionState[course.id] === 'done'}
				<div class="course-card">
					<div class="card-strip" style="background:{stripColor(i)}"></div>
					<div class="card-body">
						<div class="card-title-row">
							<a href="/courses/{course.id}" class="card-title">{course.title}</a>
							{#if !course.is_published}<span class="badge-draft">Draft</span>{/if}
						</div>
						{#if course.description}
							<p class="card-desc">{course.description}</p>
						{/if}
						<div class="card-footer">
							<span class="card-date">{new Date(course.created_at).toLocaleDateString()}</span>
							<div class="card-actions">
								{#if isEnrolled}
									<span class="enrolled-badge">✓ Enrolled</span>
									<a href="/courses/{course.id}" class="btn-sm">Continue →</a>
								{:else if canEnroll && course.is_published}
									<button
										class="btn-sm btn-enroll"
										onclick={() => handleEnroll(course.id)}
										disabled={actionState[course.id] === 'enrolling'}
									>
										{actionState[course.id] === 'enrolling' ? '…' : 'Enroll'}
									</button>
								{/if}
								{#if canManage}
									<a href="/courses/{course.id}" class="btn-sm btn-outline">View</a>
									<a href="/courses/{course.id}/edit" class="btn-sm btn-outline">Edit</a>
									<button
										class="btn-sm btn-danger"
										onclick={() => handleDelete(course.id, course.title)}
										disabled={actionState[course.id] === 'deleting'}
									>
										{actionState[course.id] === 'deleting' ? '…' : 'Delete'}
									</button>
								{/if}
							</div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.page-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 1.5rem; gap: 1rem; }
	h1 { font-size: 1.75rem; font-weight: 700; color: var(--text); }
	.page-sub { color: var(--text-3); font-size: 0.875rem; margin-top: 0.2rem; }

	/* ── Filter tabs ── */
	.filter-tabs {
		display: flex;
		gap: 4px;
		margin-bottom: 1.25rem;
		background: var(--primary-faint);
		border-radius: 10px;
		padding: 4px;
		width: fit-content;
	}
	.filter-tabs button {
		background: transparent;
		border: none;
		padding: 6px 16px;
		border-radius: 7px;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text-3);
		cursor: pointer;
		transition: all 0.15s;
		font-family: inherit;
	}
	.filter-tabs button:hover { color: var(--primary); }
	.filter-tabs button.active { background: var(--bg-card); color: var(--primary); font-weight: 600; box-shadow: 0 1px 4px rgba(0,0,0,0.08); }

	/* ── Grid ── */
	.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 1.1rem; }

	/* ── Course card ── */
	.course-card {
		background: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: 14px;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		transition: box-shadow 0.15s, transform 0.15s;
	}
	.course-card:hover { box-shadow: 0 6px 20px rgba(0,0,0,0.15); transform: translateY(-2px); }

	.card-strip { height: 6px; flex-shrink: 0; }
	.card-body { padding: 1.1rem 1.3rem; display: flex; flex-direction: column; gap: 0.65rem; flex: 1; }

	.card-title-row { display: flex; align-items: flex-start; gap: 0.6rem; }
	.card-title { font-size: 1rem; font-weight: 600; color: var(--text); flex: 1; line-height: 1.4; }
	.card-title:hover { color: var(--primary); }

	.badge-draft { background: #fef9c3; color: #a16207; font-size: 0.7rem; font-weight: 700; padding: 2px 7px; border-radius: 99px; white-space: nowrap; flex-shrink: 0; }

	.card-desc { color: var(--text-3); font-size: 0.875rem; line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; flex: 1; }

	.card-footer { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; margin-top: auto; }
	.card-date { font-size: 0.78rem; color: var(--text-4); }
	.card-actions { display: flex; gap: 0.4rem; align-items: center; flex-wrap: wrap; }

	.enrolled-badge { font-size: 0.8rem; color: #16a34a; font-weight: 600; }

	/* ── Buttons ── */
	.btn-primary { background: var(--primary); color: #fff; border: none; border-radius: 8px; padding: 0.55rem 1.1rem; font-size: 0.9rem; font-weight: 600; cursor: pointer; transition: background 0.15s; white-space: nowrap; font-family: inherit; }
	.btn-primary:hover:not(:disabled) { background: var(--primary-h); }
	.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

	.btn-sm { background: var(--primary); color: #fff; border: none; border-radius: 6px; padding: 4px 10px; font-size: 0.8rem; font-weight: 600; cursor: pointer; transition: background 0.15s; text-decoration: none; display: inline-block; font-family: inherit; white-space: nowrap; }
	.btn-sm:hover:not(:disabled) { background: var(--primary-h); text-decoration: none; }
	.btn-sm:disabled { opacity: 0.5; cursor: not-allowed; }
	.btn-sm.btn-outline { background: transparent; color: var(--text-2); border: 1px solid var(--border); }
	.btn-sm.btn-outline:hover:not(:disabled) { background: var(--bg-surface); text-decoration: none; }
	.btn-sm.btn-enroll { background: var(--primary); }
	.btn-sm.btn-danger { background: #ef4444; }
	.btn-sm.btn-danger:hover:not(:disabled) { background: #dc2626; }

	.btn-outline { background: transparent; color: var(--text-2); border: 1px solid var(--border); border-radius: 8px; padding: 0.5rem 1rem; font-size: 0.875rem; font-weight: 500; cursor: pointer; font-family: inherit; }
	.btn-outline:hover { background: var(--bg-surface); }

	/* ── Form ── */
	.form-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: 12px; padding: 1.75rem; margin-bottom: 1.75rem; }
	.form-card h2 { font-size: 1.1rem; font-weight: 600; margin-bottom: 1.25rem; color: var(--text); }
	.form-card form { display: flex; flex-direction: column; gap: 1rem; }
	.form-card label { display: flex; flex-direction: column; gap: 0.35rem; font-size: 0.875rem; font-weight: 500; color: var(--text-2); }
	.form-card input:not([type='checkbox']), .form-card textarea { padding: 0.6rem 0.85rem; border: 1px solid var(--border); border-radius: 8px; font-size: 0.9rem; outline: none; resize: vertical; font-family: inherit; background: var(--bg-input); color: var(--text); }
	.form-card input:focus, .form-card textarea:focus { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(124,58,237,0.1); }
	.check-label { flex-direction: row !important; align-items: center; gap: 0.5rem !important; }
	.form-actions { display: flex; justify-content: flex-end; }
	.req { color: #ef4444; }

	/* ── Skeleton ── */
	.skeleton { height: 160px; background: linear-gradient(90deg,var(--border) 25%,var(--border-light) 50%,var(--border) 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
	@keyframes shimmer { 0%{background-position:200% 0} 100%{background-position:-200% 0} }

	.alert { background:#fef2f2; color:#b91c1c; border:1px solid #fecaca; border-radius:8px; padding:0.75rem 1rem; font-size:0.9rem; margin-bottom:1.25rem; }
	.empty { text-align:center; padding:3.5rem 2rem; color:var(--text-3); display:flex; flex-direction:column; align-items:center; gap:1rem; }
</style>
