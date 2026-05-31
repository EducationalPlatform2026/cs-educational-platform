<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { createExercise } from '$lib/api/exercises';
	import ExerciseForm, { type ExerciseFormData } from '$lib/components/ExerciseForm.svelte';
	import { auth } from '$lib/stores/auth.svelte';
	import { onMount } from 'svelte';

	const courseId = $derived($page.params.id as string);

	let loading = $state(false);
	let error = $state('');

	onMount(() => {
		if (auth.user?.role !== 'professor' && auth.user?.role !== 'admin') {
			goto(`/courses/${courseId}`);
		}
	});

	async function handleSubmit(data: ExerciseFormData) {
		loading = true;
		error = '';
		try {
			const ex = await createExercise(courseId, {
				...data,
				description: data.description || undefined,
				template_code: data.template_code || undefined
			});
			goto(`/exercises/${ex.id}`);
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to create exercise';
			loading = false;
		}
	}
</script>

<div class="page">
	<a href="/courses/{courseId}" class="back-link">← Back to course</a>
	<div class="form-card">
		<h1>New exercise</h1>
		<ExerciseForm {loading} {error} submitLabel="Create exercise" onsubmit={handleSubmit} />
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
</style>
