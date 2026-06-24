<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { createExercise } from '$lib/api/exercises';
	import { getCourseMembers } from '$lib/api/courses';
	import ExerciseForm, { type ExerciseFormData } from '$lib/components/ExerciseForm.svelte';
	import { auth } from '$lib/stores/auth.svelte';
	import { onMount } from 'svelte';

	const courseId = $derived($page.params.id as string);

	let loading = $state(false);
	let error = $state('');

	onMount(async () => {
		const role = auth.user?.role;
		if (role === 'professor' || role === 'admin') return;
		if (role === 'student') {
			try {
				const members = await getCourseMembers(courseId);
				const me = members.find((m) => m.user_id === auth.user?.user_id);
				if (me?.role === 'teaching_assistant') return;
			} catch {
				// fall through to redirect
			}
		}
		goto(`/courses/${courseId}`);
	});

	async function handleSubmit(data: ExerciseFormData) {
		loading = true;
		error = '';
		try {
			const ex = await createExercise(courseId, {
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
</style>
