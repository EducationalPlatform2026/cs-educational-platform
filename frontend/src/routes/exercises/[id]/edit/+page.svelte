<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { getExercise, updateExercise, type Exercise } from '$lib/api/exercises';
	import ExerciseForm, { type ExerciseFormData } from '$lib/components/ExerciseForm.svelte';
	import { auth } from '$lib/stores/auth.svelte';

	const id = $derived($page.params.id as string);

	let exercise = $state<Exercise | null>(null);
	let loading = $state(false);
	let fetchError = $state('');
	let saveError = $state('');

	onMount(async () => {
		if (auth.user?.role !== 'professor' && auth.user?.role !== 'admin') {
			goto(`/exercises/${id}`);
			return;
		}
		try {
			exercise = await getExercise(id);
		} catch (err: unknown) {
			fetchError = err instanceof Error ? err.message : 'Failed to load exercise';
		}
	});

	async function handleSubmit(data: ExerciseFormData) {
		loading = true;
		saveError = '';
		try {
			await updateExercise(id, {
				...data,
				description: data.description || undefined,
				template_code: data.template_code || undefined
			});
			goto(`/exercises/${id}`);
		} catch (err: unknown) {
			saveError = err instanceof Error ? err.message : 'Failed to save exercise';
			loading = false;
		}
	}
</script>

<div class="page">
	<a href="/exercises/{id}" class="back-link">← Back to exercise</a>

	<div class="form-card">
		<h1>Edit exercise</h1>

		{#if fetchError}
			<div class="alert">{fetchError}</div>
		{:else if !exercise}
			<div class="loading-text">Loading…</div>
		{:else}
			<ExerciseForm
				initial={exercise}
				{loading}
				error={saveError}
				submitLabel="Save changes"
				onsubmit={handleSubmit}
			/>
		{/if}
	</div>
</div>

<style>
	.page { max-width: 700px; }

	.back-link {
		display: inline-block;
		font-size: 0.875rem;
		color: #6b7280;
		margin-bottom: 1.5rem;
	}
	.back-link:hover { color: #4f46e5; }

	.form-card {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		padding: 2rem;
	}

	h1 {
		font-size: 1.4rem;
		font-weight: 700;
		margin-bottom: 1.5rem;
	}

	.alert {
		background: #fef2f2;
		color: #b91c1c;
		border: 1px solid #fecaca;
		border-radius: 8px;
		padding: 0.75rem 1rem;
		font-size: 0.9rem;
	}

	.loading-text { color: #6b7280; font-size: 0.9rem; }
</style>
