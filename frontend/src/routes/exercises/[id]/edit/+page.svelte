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
		if (auth.user?.role !== 'professor' && auth.user?.role !== 'teaching_assistant' && auth.user?.role !== 'admin') {
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
				title: data.title,
				description: data.description || undefined,
				instructions: data.instructions,
				difficulty: data.difficulty,
				exercise_type: data.exercise_type,
				language: data.exercise_type === 'coding' ? data.language : undefined,
				template_code: data.exercise_type === 'coding' ? (data.template_code || undefined) : undefined,
				time_limit_ms: data.exercise_type === 'coding' ? data.time_limit_ms : undefined,
				memory_limit_kb: data.exercise_type === 'coding' ? data.memory_limit_kb : undefined,
				is_published: data.is_published,
				quiz_options: data.exercise_type === 'quiz' ? data.quiz_options : undefined,
				quiz_correct: data.exercise_type === 'quiz' ? data.quiz_correct : undefined,
				quiz_allow_multiple: data.exercise_type === 'quiz' ? data.quiz_allow_multiple : undefined
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
		color: var(--text-3);
		margin-bottom: 1.5rem;
	}
	.back-link:hover { color: var(--primary); }

	.form-card {
		background: var(--bg-card);
		border: 1px solid var(--border);
		border-radius: 12px;
		padding: 2rem;
	}

	h1 {
		font-size: 1.4rem;
		font-weight: 700;
		margin-bottom: 1.5rem;
		color: var(--text);
	}

	.alert {
		background: #fef2f2;
		color: #b91c1c;
		border: 1px solid #fecaca;
		border-radius: 8px;
		padding: 0.75rem 1rem;
		font-size: 0.9rem;
	}

	.loading-text { color: var(--text-3); font-size: 0.9rem; }
</style>
