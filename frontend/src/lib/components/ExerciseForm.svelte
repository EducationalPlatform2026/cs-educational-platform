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
		language: string;
		template_code: string;
		time_limit_ms: number;
		memory_limit_kb: number;
		is_published: boolean;
	}

	let { initial = {}, loading = false, error = '', submitLabel = 'Save', onsubmit }: Props = $props();

	// untrack: form fields are initialized once from props — intentionally not reactive to prop changes.
	let title        = $state(untrack(() => initial.title         ?? ''));
	let description  = $state(untrack(() => initial.description   ?? ''));
	let instructions = $state(untrack(() => initial.instructions  ?? ''));
	let difficulty   = $state(untrack(() => initial.difficulty    ?? 'medium'));
	let language     = $state(untrack(() => initial.language      ?? 'python'));
	let templateCode = $state(untrack(() => initial.template_code ?? ''));
	let timeLimitMs  = $state(untrack(() => initial.time_limit_ms  ?? 2000));
	let memLimitKb   = $state(untrack(() => initial.memory_limit_kb ?? 65536));
	let isPublished  = $state(untrack(() => initial.is_published  ?? false));

	const difficulties = ['easy', 'medium', 'hard'];
	const languages = ['python', 'go', 'java', 'c', 'cpp', 'javascript'];

	function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		onsubmit({
			title,
			description,
			instructions,
			difficulty,
			language,
			template_code: templateCode,
			time_limit_ms: timeLimitMs,
			memory_limit_kb: memLimitKb,
			is_published: isPublished
		});
	}
</script>

{#if error}
	<div class="alert">{error}</div>
{/if}

<form onsubmit={handleSubmit}>
	<label>
		Title <span class="req">*</span>
		<input bind:value={title} placeholder="Binary Search" required />
	</label>

	<label>
		Description
		<input bind:value={description} placeholder="Short summary shown in the exercise list" />
	</label>

	<label>
		Instructions <span class="req">*</span>
		<textarea bind:value={instructions} rows="6" placeholder="Describe the problem, constraints, and examples…" required></textarea>
	</label>

	<div class="row">
		<label>
			Difficulty
			<select bind:value={difficulty}>
				{#each difficulties as d}
					<option value={d}>{d.charAt(0).toUpperCase() + d.slice(1)}</option>
				{/each}
			</select>
		</label>

		<label>
			Language <span class="req">*</span>
			<select bind:value={language}>
				{#each languages as l}
					<option value={l}>{l}</option>
				{/each}
			</select>
		</label>
	</div>

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

	input:not([type='checkbox']),
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
		border-color: #4f46e5;
		box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
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
		background: #4f46e5;
		color: #fff;
		border: none;
		border-radius: 8px;
		padding: 0.65rem 1.5rem;
		font-size: 0.95rem;
		font-weight: 600;
		cursor: pointer;
		transition: background 0.15s;
	}

	.btn-primary:hover:not(:disabled) { background: #4338ca; }
	.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

	.req  { color: #ef4444; }
	.hint { color: #9ca3af; font-weight: 400; font-size: 0.8rem; }
</style>
