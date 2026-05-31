<script lang="ts">
	import { onMount } from 'svelte';
	import { listCourses, createCourse, deleteCourse, enrollCourse, type Course } from '$lib/api/courses';
	import { auth } from '$lib/stores/auth.svelte';

	let courses = $state<Course[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Create-course form (professors/admins only)
	let showForm = $state(false);
	let newTitle = $state('');
	let newDesc = $state('');
	let newPublished = $state(true);
	let formError = $state('');
	let formLoading = $state(false);

	// Per-card action state
	let actionState = $state<Record<string, string>>({}); // courseId → 'enrolling' | 'deleting' | 'done'

	const canManage =
		auth.user?.role === 'professor' || auth.user?.role === 'admin';
	const canEnroll =
		auth.user?.role === 'student' || auth.user?.role === 'teaching_assistant';

	onMount(async () => {
		await load();
	});

	async function load() {
		loading = true;
		error = '';
		try {
			courses = await listCourses();
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load courses';
		} finally {
			loading = false;
		}
	}

	async function handleCreate(e: SubmitEvent) {
		e.preventDefault();
		formError = '';
		formLoading = true;
		try {
			const course = await createCourse({
				title: newTitle,
				description: newDesc || undefined,
				is_published: newPublished
			});
			courses = [course, ...courses];
			newTitle = '';
			newDesc = '';
			newPublished = true;
			showForm = false;
		} catch (err: unknown) {
			formError = err instanceof Error ? err.message : 'Failed to create course';
		} finally {
			formLoading = false;
		}
	}

	async function handleEnroll(id: string) {
		actionState = { ...actionState, [id]: 'enrolling' };
		try {
			await enrollCourse(id);
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
				{#if loading}
					Loading…
				{:else}
					{courses.length} course{courses.length !== 1 ? 's' : ''} available
				{/if}
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
			{#if formError}
				<div class="alert">{formError}</div>
			{/if}
			<form onsubmit={handleCreate}>
				<label>
					Title <span class="req">*</span>
					<input bind:value={newTitle} placeholder="Introduction to Algorithms" required />
				</label>
				<label>
					Description
					<textarea bind:value={newDesc} placeholder="What will students learn?" rows="3"></textarea>
				</label>
				<label class="checkbox-label">
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

	<!-- Error state -->
	{#if error}
		<div class="alert">{error}</div>
	{/if}

	<!-- Loading skeleton -->
	{#if loading}
		<div class="grid">
			{#each [1, 2, 3] as _}
				<div class="course-card skeleton"></div>
			{/each}
		</div>

	<!-- Empty state -->
	{:else if courses.length === 0}
		<div class="empty">
			<p>No courses yet.</p>
			{#if canManage}
				<p>Create the first one using the button above.</p>
			{/if}
		</div>

	<!-- Course grid -->
	{:else}
		<div class="grid">
			{#each courses as course (course.id)}
				<div class="course-card">
					<div class="card-header">
						<a href="/courses/{course.id}" class="card-title">{course.title}</a>
						{#if !course.is_published}
							<span class="badge-draft">Draft</span>
						{/if}
					</div>
					{#if course.description}
						<p class="card-desc">{course.description}</p>
					{/if}
					<div class="card-footer">
						<span class="card-date">
							{new Date(course.created_at).toLocaleDateString()}
						</span>
						<div class="card-actions">
							{#if canEnroll}
								{#if actionState[course.id] === 'done'}
									<span class="enrolled-badge">✓ Enrolled</span>
								{:else}
									<button
										class="btn-sm"
										onclick={() => handleEnroll(course.id)}
										disabled={actionState[course.id] === 'enrolling'}
									>
										{actionState[course.id] === 'enrolling' ? '…' : 'Enroll'}
									</button>
								{/if}
							{/if}
							{#if canManage}
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
			{/each}
		</div>
	{/if}
</div>

<style>
	.page-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		margin-bottom: 1.75rem;
		gap: 1rem;
	}

	h1 {
		font-size: 1.75rem;
		font-weight: 700;
	}

	.page-sub {
		color: #6b7280;
		font-size: 0.875rem;
		margin-top: 0.2rem;
	}

	.form-card {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		padding: 1.75rem;
		margin-bottom: 1.75rem;
	}

	.form-card h2 {
		font-size: 1.1rem;
		font-weight: 600;
		margin-bottom: 1.25rem;
	}

	.form-card form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.form-card label {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}

	.form-card input:not([type='checkbox']),
	.form-card textarea {
		padding: 0.6rem 0.85rem;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		font-size: 0.9rem;
		outline: none;
		resize: vertical;
		font-family: inherit;
	}

	.form-card input:focus,
	.form-card textarea:focus {
		border-color: #4f46e5;
		box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
	}

	.checkbox-label {
		flex-direction: row !important;
		align-items: center;
		gap: 0.5rem !important;
		font-size: 0.9rem !important;
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
	}

	.req {
		color: #ef4444;
	}

	.alert {
		background: #fef2f2;
		color: #b91c1c;
		border: 1px solid #fecaca;
		border-radius: 8px;
		padding: 0.75rem 1rem;
		font-size: 0.9rem;
		margin-bottom: 1.25rem;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1.25rem;
	}

	.course-card {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		padding: 1.25rem 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		transition: box-shadow 0.15s;
	}

	.course-card:hover {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
	}

	.skeleton {
		height: 140px;
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
	}

	@keyframes shimmer {
		0% { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}

	.card-header {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
	}

	.card-title {
		font-size: 1rem;
		font-weight: 600;
		color: #1a1a2e;
		flex: 1;
		line-height: 1.4;
	}

	.card-title:hover {
		color: #4f46e5;
	}

	.badge-draft {
		background: #fef9c3;
		color: #a16207;
		font-size: 0.7rem;
		font-weight: 700;
		padding: 2px 7px;
		border-radius: 99px;
		white-space: nowrap;
	}

	.card-desc {
		color: #6b7280;
		font-size: 0.875rem;
		line-height: 1.5;
		flex: 1;
		display: -webkit-box;
		-webkit-line-clamp: 3;
		line-clamp: 3;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.card-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
		margin-top: auto;
	}

	.card-date {
		font-size: 0.78rem;
		color: #9ca3af;
	}

	.card-actions {
		display: flex;
		gap: 0.5rem;
		align-items: center;
	}

	.btn-primary {
		background: #4f46e5;
		color: #fff;
		border: none;
		border-radius: 8px;
		padding: 0.55rem 1.1rem;
		font-size: 0.9rem;
		font-weight: 600;
		cursor: pointer;
		transition: background 0.15s;
		white-space: nowrap;
	}

	.btn-primary:hover:not(:disabled) {
		background: #4338ca;
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-sm {
		background: #4f46e5;
		color: #fff;
		border: none;
		border-radius: 6px;
		padding: 4px 10px;
		font-size: 0.8rem;
		font-weight: 600;
		cursor: pointer;
		transition: background 0.15s;
		text-decoration: none;
		display: inline-block;
	}

	.btn-sm:hover:not(:disabled) {
		background: #4338ca;
	}

	.btn-sm:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-outline {
		background: transparent;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-outline:hover {
		background: #f9fafb !important;
		color: #374151 !important;
	}

	.btn-danger {
		background: #ef4444;
	}

	.btn-danger:hover:not(:disabled) {
		background: #dc2626 !important;
	}

	.enrolled-badge {
		font-size: 0.8rem;
		color: #16a34a;
		font-weight: 600;
	}

	.empty {
		text-align: center;
		padding: 4rem 2rem;
		color: #6b7280;
	}

	.empty p {
		font-size: 0.95rem;
	}
</style>
