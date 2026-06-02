<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getCourse, updateCourse } from '$lib/api/courses';
	import { auth } from '$lib/stores/auth.svelte';

	const id = $derived($page.params.id as string);

	let loading = $state(true);
	let saving = $state(false);
	let error = $state('');
	let formError = $state('');

	let title = $state('');
	let description = $state('');
	let isPublished = $state(false);

	const canEdit = $derived(auth.user?.role === 'professor' || auth.user?.role === 'admin');

	onMount(async () => {
		if (!canEdit) { goto(`/courses/${id}`); return; }
		try {
			const course = await getCourse(id);
			title       = course.title;
			description = course.description ?? '';
			isPublished = course.is_published;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load course';
		} finally {
			loading = false;
		}
	});

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!title.trim()) { formError = 'Title is required'; return; }
		formError = '';
		saving = true;
		try {
			await updateCourse(id, {
				title: title.trim(),
				description: description.trim() || undefined,
				is_published: isPublished
			});
			goto(`/courses/${id}`);
		} catch (err: unknown) {
			formError = err instanceof Error ? err.message : 'Failed to save';
		} finally {
			saving = false;
		}
	}
</script>

<div class="page">
	<a href="/courses/{id}" class="back-link">← Back to course</a>

	<h1>Edit course</h1>

	{#if loading}
		<div class="skeleton"></div>

	{:else if error}
		<div class="alert">{error}</div>

	{:else}
		<div class="form-card">
			{#if formError}
				<div class="alert">{formError}</div>
			{/if}

			<form onsubmit={handleSubmit}>
				<label>
					Title <span class="req">*</span>
					<input bind:value={title} placeholder="Course title" required />
				</label>

				<label>
					Description
					<textarea bind:value={description} placeholder="What will students learn?" rows="4"></textarea>
				</label>

				<label class="checkbox-label">
					<input type="checkbox" bind:checked={isPublished} />
					Published (visible to students)
				</label>

				<div class="form-actions">
					<a href="/courses/{id}" class="btn-outline">Cancel</a>
					<button type="submit" class="btn-primary" disabled={saving}>
						{saving ? 'Saving…' : 'Save changes'}
					</button>
				</div>
			</form>
		</div>
	{/if}
</div>

<style>
	.page { max-width: 600px; }

	.back-link {
		display: inline-block;
		font-size: 0.875rem;
		color: #6b7280;
		margin-bottom: 1.5rem;
	}
	.back-link:hover { color: #4f46e5; }

	h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; }

	.form-card {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		padding: 1.75rem;
	}

	form { display: flex; flex-direction: column; gap: 1.1rem; }

	label {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}

	input:not([type='checkbox']),
	textarea {
		padding: 0.6rem 0.85rem;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		font-size: 0.9rem;
		font-family: inherit;
		outline: none;
		resize: vertical;
	}

	input:focus, textarea:focus {
		border-color: #4f46e5;
		box-shadow: 0 0 0 3px rgba(79,70,229,0.1);
	}

	.checkbox-label {
		flex-direction: row !important;
		align-items: center;
		gap: 0.5rem !important;
	}

	.req { color: #ef4444; }

	.form-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		margin-top: 0.5rem;
	}

	.btn-primary {
		background: #4f46e5;
		color: #fff;
		border: none;
		border-radius: 8px;
		padding: 0.55rem 1.2rem;
		font-size: 0.9rem;
		font-weight: 600;
		cursor: pointer;
	}
	.btn-primary:hover:not(:disabled) { background: #4338ca; }
	.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

	.btn-outline {
		background: transparent;
		color: #374151;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
		font-weight: 500;
		text-decoration: none;
	}
	.btn-outline:hover { background: #f9fafb; text-decoration: none; }

	.alert {
		background: #fef2f2;
		color: #b91c1c;
		border: 1px solid #fecaca;
		border-radius: 8px;
		padding: 0.75rem 1rem;
		font-size: 0.9rem;
		margin-bottom: 1rem;
	}

	.skeleton {
		height: 300px;
		border-radius: 12px;
		background: linear-gradient(90deg,#f0f0f0 25%,#e8e8e8 50%,#f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
	}
	@keyframes shimmer {
		0%   { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}
</style>
