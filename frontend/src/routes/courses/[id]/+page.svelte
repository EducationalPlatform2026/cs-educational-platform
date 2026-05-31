<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getCourse, getCourseMembers, enrollCourse, type Course, type Member } from '$lib/api/courses';
	import { auth } from '$lib/stores/auth.svelte';

	// `id` is always present in this route — the `[id]` segment guarantees it.
	const id = $derived($page.params.id as string);

	let course = $state<Course | null>(null);
	let members = $state<Member[]>([]);
	let loading = $state(true);
	let error = $state('');
	let enrolling = $state(false);
	let enrolled = $state(false);

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
			const [c, m] = await Promise.all([
				getCourse(id),
				canSeeMembers ? getCourseMembers(id) : Promise.resolve([])
			]);
			course = c;
			members = m;
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
