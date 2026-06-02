<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getCourse, getCourseMembers, enrollCourse, type Course, type Member } from '$lib/api/courses';
	import { listExercises, deleteExercise, type Exercise } from '$lib/api/exercises';
	import { auth } from '$lib/stores/auth.svelte';

	// `id` is always present in this route — the `[id]` segment guarantees it.
	const id = $derived($page.params.id as string);

	let course = $state<Course | null>(null);
	let members = $state<Member[]>([]);
	let exList = $state<Exercise[]>([]);
	let loading = $state(true);
	let error = $state('');
	let enrolling = $state(false);
	let enrolled = $state(false);
	let deletingEx = $state<Record<string, boolean>>({});

	const canManage = $derived(
		auth.user?.role === 'professor' || auth.user?.role === 'admin'
	);
	const canEnroll = $derived(
		auth.user?.role === 'student' || auth.user?.role === 'teaching_assistant'
	);
	const canSeeMembers = $derived(
		auth.user?.role === 'professor' ||
		auth.user?.role === 'teaching_assistant' ||
		auth.user?.role === 'admin'
	);

	onMount(async () => {
		try {
			const [c, m, ex] = await Promise.all([
				getCourse(id),
				canSeeMembers ? getCourseMembers(id) : Promise.resolve([]),
				listExercises(id)
			]);
			course = c;
			enrolled = c.is_enrolled; // restore persisted enrollment state on every load
			members = m;
			exList = ex;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load course';
		} finally {
			loading = false;
		}
	});

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

	function formatDate(iso: string) {
		return new Date(iso).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function roleLabel(role: string) {
		return role.replace('_', ' ');
	}
</script>

<div class="page">
	<a href="/courses" class="back-link">← All courses</a>

	{#if loading}
		<div class="skeleton-header"></div>

	{:else if error}
		<div class="alert">{error}</div>

	{:else if course}
		<div class="course-header">
			<div>
				<div class="title-row">
					<h1>{course.title}</h1>
					{#if !course.is_published}
						<span class="badge-draft">Draft</span>
					{/if}
				</div>
				<p class="meta">Created {formatDate(course.created_at)}</p>
			</div>

			<div class="header-actions">
				{#if canEnroll && course.is_published}
					{#if enrolled}
						<span class="enrolled-badge">✓ Enrolled</span>
					{:else}
						<button class="btn-primary" onclick={handleEnroll} disabled={enrolling}>
							{enrolling ? 'Enrolling…' : 'Enroll'}
						</button>
					{/if}
				{/if}
				{#if canManage}
					<a href="/courses/{course.id}/edit" class="btn-outline">Edit course</a>
				{/if}
			</div>
		</div>

		{#if course.description}
			<div class="description-card">
				<p>{course.description}</p>
			</div>
		{/if}

		<!-- Exercises section -->
		<section class="exercises-section">
			<div class="section-header">
				<h2>Exercises ({exList.length})</h2>
				{#if canManage}
					<a href="/courses/{id}/exercises/new" class="btn-sm">+ Add exercise</a>
				{/if}
			</div>

			{#if exList.length === 0}
				<p class="empty-text">
					{canManage ? 'No exercises yet. Add the first one.' : 'No exercises available yet.'}
				</p>
			{:else}
				<div class="exercise-list">
					{#each exList as ex (ex.id)}
						<div class="exercise-row">
							<div class="ex-info">
								<a href="/exercises/{ex.id}" class="ex-title">{ex.title}</a>
								<div class="ex-badges">
									<span class="badge diff-{ex.difficulty}">{ex.difficulty}</span>
									<span class="badge lang">{ex.language}</span>
									{#if !ex.is_published}
										<span class="badge draft">draft</span>
									{/if}
								</div>
							</div>
							{#if canManage}
								<div class="ex-actions">
									<a href="/exercises/{ex.id}/edit" class="btn-xs btn-outline">Edit</a>
									<button
										class="btn-xs btn-danger"
										onclick={() => handleDeleteExercise(ex.id, ex.title)}
										disabled={deletingEx[ex.id]}
									>Delete</button>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</section>

		{#if canSeeMembers}
			<section class="members-section">
				<h2>Enrolled members ({members.length})</h2>
				{#if members.length === 0}
					<p class="empty-text">No members enrolled yet.</p>
				{:else}
					<div class="members-table-wrap">
						<table class="members-table">
							<thead>
								<tr>
									<th>Name</th>
									<th>Role</th>
									<th>Enrolled</th>
								</tr>
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
				{/if}
			</section>
		{/if}
	{/if}
</div>

<style>
	.page {
		max-width: 800px;
	}

	.back-link {
		display: inline-block;
		font-size: 0.875rem;
		color: #6b7280;
		margin-bottom: 1.5rem;
	}

	.back-link:hover {
		color: #4f46e5;
	}

	.skeleton-header {
		height: 100px;
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
		border-radius: 12px;
	}

	@keyframes shimmer {
		0% { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}

	.alert {
		background: #fef2f2;
		color: #b91c1c;
		border: 1px solid #fecaca;
		border-radius: 8px;
		padding: 0.75rem 1rem;
		font-size: 0.9rem;
	}

	.course-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.title-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-wrap: wrap;
	}

	h1 {
		font-size: 1.75rem;
		font-weight: 700;
		line-height: 1.3;
	}

	.meta {
		color: #6b7280;
		font-size: 0.875rem;
		margin-top: 0.4rem;
	}

	.badge-draft {
		background: #fef9c3;
		color: #a16207;
		font-size: 0.7rem;
		font-weight: 700;
		padding: 2px 7px;
		border-radius: 99px;
	}

	.header-actions {
		display: flex;
		gap: 0.75rem;
		align-items: center;
		flex-shrink: 0;
	}

	.description-card {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		padding: 1.25rem 1.5rem;
		margin-bottom: 2rem;
		color: #374151;
		line-height: 1.65;
	}

	.exercises-section {
		margin-top: 2rem;
	}

	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
	}

	.section-header h2 {
		font-size: 1.1rem;
		font-weight: 600;
	}

	.exercise-list {
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		overflow: hidden;
	}

	.exercise-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.85rem 1.1rem;
		border-bottom: 1px solid #f3f4f6;
		gap: 1rem;
	}

	.exercise-row:last-child {
		border-bottom: none;
	}

	.exercise-row:hover {
		background: #fafafa;
	}

	.ex-info {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-wrap: wrap;
		flex: 1;
	}

	.ex-title {
		font-size: 0.925rem;
		font-weight: 500;
		color: #1a1a2e;
	}

	.ex-title:hover {
		color: #4f46e5;
	}

	.ex-badges {
		display: flex;
		gap: 0.4rem;
	}

	.badge {
		font-size: 0.7rem;
		font-weight: 600;
		padding: 2px 7px;
		border-radius: 99px;
		text-transform: capitalize;
	}

	.diff-easy   { background: #dcfce7; color: #166534; }
	.diff-medium { background: #fef9c3; color: #a16207; }
	.diff-hard   { background: #fee2e2; color: #b91c1c; }
	.lang        { background: #e0e7ff; color: #3730a3; }
	.draft       { background: #f3f4f6; color: #6b7280; }

	.ex-actions {
		display: flex;
		gap: 0.4rem;
		flex-shrink: 0;
	}

	.btn-xs {
		font-size: 0.78rem;
		font-weight: 600;
		padding: 3px 9px;
		border-radius: 5px;
		border: none;
		cursor: pointer;
		transition: all 0.15s;
		text-decoration: none;
		display: inline-block;
	}

	.btn-sm {
		background: #4f46e5;
		color: #fff;
		border: none;
		border-radius: 7px;
		padding: 5px 12px;
		font-size: 0.82rem;
		font-weight: 600;
		cursor: pointer;
		text-decoration: none;
	}

	.btn-sm:hover { background: #4338ca; text-decoration: none; }

	.btn-xs.btn-outline {
		background: transparent;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-xs.btn-outline:hover { background: #f9fafb; }

	.btn-xs.btn-danger { background: #ef4444; color: #fff; }
	.btn-xs.btn-danger:hover:not(:disabled) { background: #dc2626; }
	.btn-xs:disabled { opacity: 0.5; cursor: not-allowed; }

	.members-section {
		margin-top: 2rem;
	}

	.members-section h2 {
		font-size: 1.1rem;
		font-weight: 600;
		margin-bottom: 1rem;
	}

	.empty-text {
		color: #6b7280;
		font-size: 0.9rem;
	}

	.members-table-wrap {
		overflow-x: auto;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
	}

	.members-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.9rem;
	}

	.members-table th {
		text-align: left;
		padding: 0.75rem 1rem;
		font-size: 0.8rem;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
	}

	.members-table td {
		padding: 0.75rem 1rem;
		border-bottom: 1px solid #f3f4f6;
		color: #374151;
	}

	.members-table tr:last-child td {
		border-bottom: none;
	}

	.role-badge {
		background: #eef2ff;
		color: #4f46e5;
		font-size: 0.75rem;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 99px;
		text-transform: capitalize;
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
	}

	.btn-primary:hover:not(:disabled) {
		background: #4338ca;
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-outline {
		background: transparent;
		color: #374151;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		text-decoration: none;
		transition: all 0.15s;
	}

	.btn-outline:hover {
		background: #f9fafb;
		text-decoration: none;
	}

	.enrolled-badge {
		font-size: 0.9rem;
		color: #16a34a;
		font-weight: 600;
	}
</style>
