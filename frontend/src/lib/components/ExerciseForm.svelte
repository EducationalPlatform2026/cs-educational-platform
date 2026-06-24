<script lang="ts">
	import { untrack } from 'svelte';
	import type { Exercise } from '$lib/api/exercises';

	interface Props {
		initial?: Partial<Exercise>;
		loading?: boolean;
		error?: string;
		submitLabel?: string;
		onsubmit: (data: ExerciseFormData) => void;
	}

	export interface ExerciseFormData {
		title: string;
		description: string;
		instructions: string;
		difficulty: string;
		exercise_type: string;
		language: string;
		template_code: string;
		time_limit_ms: number;
		memory_limit_kb: number;
		is_published: boolean;
		quiz_options: string[];
		quiz_correct: number[];
		quiz_allow_multiple: boolean;
	}

	let { initial = {}, loading = false, error = '', submitLabel = 'Save', onsubmit }: Props = $props();

	// untrack: form fields are initialized once from props — intentionally not reactive to prop changes.
	let title        = $state(untrack(() => initial.title         ?? ''));
	let description  = $state(untrack(() => initial.description   ?? ''));
	let instructions = $state(untrack(() => initial.instructions  ?? ''));
	let difficulty   = $state(untrack(() => initial.difficulty    ?? 'medium'));
	let exerciseType = $state(untrack(() => initial.exercise_type ?? 'coding'));
	let language     = $state(untrack(() => initial.language      ?? 'python'));
	let templateCode = $state(untrack(() => initial.template_code ?? ''));
	let timeLimitMs  = $state(untrack(() => initial.time_limit_ms  ?? 2000));
	let memLimitKb   = $state(untrack(() => initial.memory_limit_kb ?? 65536));
	let isPublished  = $state(untrack(() => initial.is_published  ?? false));
	// Quiz fields — default to 4 blank options
	let quizOptions       = $state<string[]>(untrack(() =>
		initial.quiz_options && initial.quiz_options.length > 0
			? [...initial.quiz_options, ...Array(Math.max(0, 2 - initial.quiz_options.length)).fill('')]
			: ['', '', '', '']
	));
	let quizCorrect       = $state<number[]>(untrack(() => initial.quiz_correct ?? [0]));
	let quizAllowMultiple = $state(untrack(() => initial.quiz_allow_multiple ?? false));

	const difficulties = ['easy', 'medium', 'hard'];
	const languages = ['python', 'go', 'java', 'c', 'cpp', 'javascript'];

	function addOption() {
		if (quizOptions.length < 4) quizOptions = [...quizOptions, ''];
	}
	function removeOption(i: number) {
		if (quizOptions.length <= 2) return;
		quizOptions = quizOptions.filter((_, idx) => idx !== i);
		// Remove or clamp correct indices that referenced the removed option
		quizCorrect = quizCorrect
			.filter((v) => v !== i)
			.map((v) => (v > i ? v - 1 : v));
		if (quizCorrect.length === 0) quizCorrect = [0];
	}

	function toggleCorrect(i: number, checked: boolean) {
		if (quizAllowMultiple) {
			quizCorrect = checked
				? [...quizCorrect, i].sort((a, b) => a - b)
				: quizCorrect.filter((v) => v !== i);
			if (quizCorrect.length === 0) quizCorrect = [i]; // must have at least one
		} else {
			quizCorrect = [i];
		}
	}

	function handleMultipleToggle() {
		quizAllowMultiple = !quizAllowMultiple;
		// When switching to single-choice, keep only the first correct answer
		if (!quizAllowMultiple && quizCorrect.length > 1) {
			quizCorrect = [quizCorrect[0]];
		}
	}

	function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		const filteredOptions = quizOptions.map((o) => o.trim()).filter(Boolean);
		onsubmit({
			title,
			description,
			instructions,
			difficulty,
			exercise_type: exerciseType,
			language: exerciseType === 'coding' ? language : '',
			template_code: templateCode,
			time_limit_ms: timeLimitMs,
			memory_limit_kb: memLimitKb,
			is_published: isPublished,
			quiz_options: exerciseType === 'quiz' ? filteredOptions : [],
			quiz_correct: exerciseType === 'quiz' ? quizCorrect : [],
			quiz_allow_multiple: exerciseType === 'quiz' ? quizAllowMultiple : false
		});
	}
</script>

{#if error}
	<div class="alert">{error}</div>
{/if}

<form onsubmit={handleSubmit}>
	<!-- Exercise type toggle -->
	<div class="type-toggle">
		<span class="toggle-label">Exercise type</span>
		<div class="toggle-btns">
			<button
				type="button"
				class="toggle-btn"
				class:active={exerciseType === 'coding'}
				onclick={() => (exerciseType = 'coding')}
			>
				💻 Coding problem
			</button>
			<button
				type="button"
				class="toggle-btn"
				class:active={exerciseType === 'quiz'}
				onclick={() => (exerciseType = 'quiz')}
			>
				❓ Quiz
			</button>
		</div>
	</div>

	<label>
		<span>Title <span class="req">*</span></span>
		<input bind:value={title} placeholder={exerciseType === 'quiz' ? 'What is Big-O notation?' : 'Binary Search'} required />
	</label>

	<label>
		Description
		<input bind:value={description} placeholder="Short summary shown in the exercise list" />
	</label>

	<label>
		<span>Instructions <span class="req">*</span></span>
		<textarea bind:value={instructions} rows="5" placeholder={exerciseType === 'quiz' ? 'Explain the question context here…' : 'Describe the problem, constraints, and examples…'} required></textarea>
	</label>

	<label>
		Difficulty
		<select bind:value={difficulty}>
			{#each difficulties as d}
				<option value={d}>{d.charAt(0).toUpperCase() + d.slice(1)}</option>
			{/each}
		</select>
	</label>

	<!-- ── Coding-specific fields ── -->
	{#if exerciseType === 'coding'}
		<label>
			<span>Language <span class="req">*</span></span>
			<select bind:value={language}>
				{#each languages as l}
					<option value={l}>{l}</option>
				{/each}
			</select>
		</label>

		<label>
			Starter code <span class="hint">(optional — shown to students as a template)</span>
			<textarea bind:value={templateCode} rows="5" placeholder="def solution():\n    pass" style="font-family: monospace;"></textarea>
		</label>

		<div class="row">
			<label>
				Time limit (ms)
				<input type="number" bind:value={timeLimitMs} min="100" max="30000" step="100" />
			</label>
			<label>
				Memory limit (KB)
				<input type="number" bind:value={memLimitKb} min="1024" max="524288" step="1024" />
			</label>
		</div>
	{/if}

	<!-- ── Quiz-specific fields ── -->
	{#if exerciseType === 'quiz'}
		<div class="quiz-section">
			<div class="quiz-header">
				<span class="quiz-section-label">Answer options <span class="req">*</span></span>
				<span class="quiz-hint">
					{#if quizAllowMultiple}
						Check all correct answers.
					{:else}
						Select the radio button next to the correct answer.
					{/if}
				</span>
			</div>

			{#each quizOptions as opt, i}
				<div class="quiz-option-row">
					{#if quizAllowMultiple}
						<input
							type="checkbox"
							checked={quizCorrect.includes(i)}
							onchange={(e) => toggleCorrect(i, (e.target as HTMLInputElement).checked)}
							class="quiz-radio"
						/>
					{:else}
						<input
							type="radio"
							name="quiz-correct"
							value={i}
							checked={quizCorrect[0] === i}
							onchange={() => toggleCorrect(i, true)}
							class="quiz-radio"
						/>
					{/if}
					<span class="option-label">{String.fromCharCode(65 + i)}</span>
					<input
						bind:value={quizOptions[i]}
						placeholder="Option {String.fromCharCode(65 + i)}"
						class="option-input"
						required
					/>
					{#if quizOptions.length > 2}
						<button type="button" class="remove-btn" onclick={() => removeOption(i)}>✕</button>
					{/if}
				</div>
			{/each}

			{#if quizOptions.length < 4}
				<button type="button" class="add-option-btn" onclick={addOption}>+ Add option</button>
			{/if}

			<label class="multiple-toggle">
				<input type="checkbox" checked={quizAllowMultiple} onchange={handleMultipleToggle} />
				Allow multiple correct answers
			</label>
		</div>
	{/if}

	<label class="checkbox-label">
		<input type="checkbox" bind:checked={isPublished} />
		Publish immediately
	</label>

	<div class="form-actions">
		<button type="submit" class="btn-primary" disabled={loading}>
			{loading ? 'Saving…' : submitLabel}
		</button>
	</div>
</form>

<style>
	.alert {
		background: #fef2f2;
		color: #b91c1c;
		border: 1px solid #fecaca;
		border-radius: 8px;
		padding: 0.75rem 1rem;
		font-size: 0.9rem;
		margin-bottom: 1.25rem;
	}

	form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	/* ── Type toggle ── */
	.type-toggle {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}
	.toggle-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}
	.toggle-btns {
		display: flex;
		gap: 0.5rem;
	}
	.toggle-btn {
		padding: 0.5rem 1rem;
		border: 2px solid #d1d5db;
		border-radius: 8px;
		background: #fff;
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
		cursor: pointer;
		transition: all 0.15s;
		font-family: inherit;
	}
	.toggle-btn:hover { border-color: #7c3aed; color: #7c3aed; }
	.toggle-btn.active {
		border-color: #7c3aed;
		background: #f5f3ff;
		color: #7c3aed;
	}

	label {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}

	.row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.75rem;
	}

	input:not([type='checkbox']):not([type='radio']),
	select,
	textarea {
		padding: 0.6rem 0.85rem;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		font-size: 0.9rem;
		outline: none;
		background: #fff;
		resize: vertical;
		font-family: inherit;
		transition: border-color 0.15s;
	}

	input:focus,
	select:focus,
	textarea:focus {
		border-color: #7c3aed;
		box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.1);
	}

	/* ── Quiz section ── */
	.quiz-section {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		padding: 1rem;
		background: #fafafa;
	}
	.quiz-header {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		margin-bottom: 0.25rem;
	}
	.quiz-section-label { font-size: 0.875rem; font-weight: 500; color: #374151; }
	.quiz-hint { font-size: 0.78rem; color: #9ca3af; }
	.quiz-option-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.quiz-radio { flex-shrink: 0; width: 16px; height: 16px; accent-color: #7c3aed; cursor: pointer; }
	.option-label {
		flex-shrink: 0;
		width: 20px;
		font-size: 0.875rem;
		font-weight: 600;
		color: #6b7280;
	}
	.option-input {
		flex: 1;
		padding: 0.5rem 0.75rem !important;
		border-radius: 6px !important;
	}
	.remove-btn {
		flex-shrink: 0;
		background: none;
		border: none;
		color: #9ca3af;
		cursor: pointer;
		font-size: 0.85rem;
		padding: 2px 6px;
		border-radius: 4px;
		line-height: 1;
		font-family: inherit;
	}
	.remove-btn:hover { background: #fee2e2; color: #b91c1c; }
	.add-option-btn {
		align-self: flex-start;
		background: none;
		border: 1px dashed #d1d5db;
		border-radius: 6px;
		padding: 0.4rem 0.85rem;
		font-size: 0.82rem;
		color: #6b7280;
		cursor: pointer;
		font-family: inherit;
		transition: all 0.15s;
	}
	.add-option-btn:hover { border-color: #7c3aed; color: #7c3aed; }

	.multiple-toggle {
		display: flex !important;
		flex-direction: row !important;
		align-items: center;
		gap: 0.5rem !important;
		font-size: 0.82rem !important;
		color: #6b7280;
		font-weight: 400 !important;
		padding-top: 0.25rem;
		border-top: 1px solid #e5e7eb;
		margin-top: 0.25rem;
	}

	.checkbox-label {
		flex-direction: row !important;
		align-items: center;
		gap: 0.5rem !important;
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
		padding-top: 0.5rem;
	}

	.btn-primary {
		background: #7c3aed;
		color: #fff;
		border: none;
		border-radius: 8px;
		padding: 0.65rem 1.5rem;
		font-size: 0.95rem;
		font-weight: 600;
		cursor: pointer;
		transition: background 0.15s;
		font-family: inherit;
	}

	.btn-primary:hover:not(:disabled) { background: #6d28d9; }
	.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

	.req  { color: #ef4444; }
	.hint { color: #9ca3af; font-weight: 400; font-size: 0.8rem; }
</style>
