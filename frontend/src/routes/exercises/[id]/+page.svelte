<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { getExercise, listTestCases, createTestCase, deleteTestCase, type Exercise, type AnyTestCase, type TestCase } from '$lib/api/exercises';
	import { submitCode, listSubmissions, type Submission, type SubmissionStatus } from '$lib/api/submissions';
	import { auth } from '$lib/stores/auth.svelte';

	const id = $derived($page.params.id as string);

	let exercise = $state<Exercise | null>(null);
	let testCases = $state<AnyTestCase[]>([]);
	let loading = $state(true);
	let error = $state('');

	const canManage = $derived(auth.user?.role === 'professor' || auth.user?.role === 'admin');
	const canSeeAll = $derived(
		auth.user?.role === 'professor' ||
		auth.user?.role === 'teaching_assistant' ||
		auth.user?.role === 'admin'
	);

	// Add test case form state
	let showTcForm = $state(false);
	let tcInput = $state('');
	let tcExpected = $state('');
	let tcHidden = $state(false);
	let tcOrdinal = $state(0);
	let tcLoading = $state(false);
	let tcError = $state('');
	let deletingTc = $state<Record<string, boolean>>({});

	// Submission state
	let code = $state('');
	let submitLang = $state('');
	let submitting = $state(false);
	let submitError = $state('');
	let latestSubmission = $state<Submission | null>(null);
	let submissionHistory = $state<Submission[]>([]);
	let showHistory = $state(false);

	onMount(async () => {
		try {
			const [ex, tc, subs] = await Promise.all([
				getExercise(id),
				listTestCases(id),
				listSubmissions(id)
			]);
			exercise = ex;
			testCases = tc;
			tcOrdinal = tc.length + 1;
			submissionHistory = subs;
			if (subs.length > 0) latestSubmission = subs[0];
			// Pre-fill code editor with template or last submission
			code = subs[0]?.code ?? ex.template_code ?? '';
			submitLang = ex.language;
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load exercise';
		} finally {
			loading = false;
		}
	});

	async function handleAddTestCase(e: SubmitEvent) {
		e.preventDefault();
		tcError = '';
		tcLoading = true;
		try {
			const tc = await createTestCase(id, {
				input: tcInput,
				expected_output: tcExpected,
				is_hidden: tcHidden,
				ordinal: tcOrdinal
			});
			testCases = [...testCases, tc];
			tcInput = '';
			tcExpected = '';
			tcHidden = false;
			tcOrdinal = testCases.length + 1;
			showTcForm = false;
		} catch (err: unknown) {
			tcError = err instanceof Error ? err.message : 'Failed to add test case';
		} finally {
			tcLoading = false;
		}
	}

	async function handleDeleteTestCase(tcId: string) {
		if (!confirm('Delete this test case?')) return;
		deletingTc = { ...deletingTc, [tcId]: true };
		try {
			await deleteTestCase(tcId);
			testCases = testCases.filter((tc) => tc.id !== tcId);
		} catch (err: unknown) {
			alert(err instanceof Error ? err.message : 'Delete failed');
		} finally {
			deletingTc = { ...deletingTc, [tcId]: false };
		}
	}

	function diffColor(d: string) {
		return d === 'easy' ? 'diff-easy' : d === 'hard' ? 'diff-hard' : 'diff-medium';
	}

	function isVisible(tc: AnyTestCase): tc is TestCase {
		return !tc.is_hidden;
	}

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		submitError = '';
		submitting = true;
		try {
			const sub = await submitCode(id, code, submitLang);
			latestSubmission = sub;
			submissionHistory = [sub, ...submissionHistory];
		} catch (err: unknown) {
			submitError = err instanceof Error ? err.message : 'Submission failed';
		} finally {
			submitting = false;
		}
	}

	const statusMeta: Record<SubmissionStatus, { label: string; cls: string }> = {
		pending:       { label: 'Pending',        cls: 'status-pending'  },
		running:       { label: 'Running…',       cls: 'status-running'  },
		accepted:      { label: 'Accepted',        cls: 'status-accepted' },
		wrong_answer:  { label: 'Wrong Answer',    cls: 'status-wrong'   },
		runtime_error: { label: 'Runtime Error',   cls: 'status-error'   },
		time_limit:    { label: 'Time Limit',      cls: 'status-error'   },
		memory_limit:  { label: 'Memory Limit',    cls: 'status-error'   },
		compile_error: { label: 'Compile Error',   cls: 'status-error'   }
	};

	function statusLabel(s: SubmissionStatus) { return statusMeta[s]?.label ?? s; }
	function statusCls(s: SubmissionStatus)   { return statusMeta[s]?.cls  ?? ''; }
</script>

<div class="page">
	<a href={exercise ? `/courses/${exercise.course_id}` : '/courses'} class="back-link">← Back to course</a>

	{#if loading}
		<div class="skeleton"></div>

	{:else if error}
		<div class="alert">{error}</div>

	{:else if exercise}
		<div class="ex-header">
			<div>
				<h1>{exercise.title}</h1>
				<div class="badges">
					<span class="badge {diffColor(exercise.difficulty)}">{exercise.difficulty}</span>
					<span class="badge lang">{exercise.language}</span>
					{#if !exercise.is_published}<span class="badge draft">draft</span>{/if}
					<span class="badge neutral">{Math.round(exercise.time_limit_ms / 1000)}s limit</span>
					<span class="badge neutral">{Math.round(exercise.memory_limit_kb / 1024)} MB</span>
				</div>
			</div>
			{#if canManage}
				<a href="/exercises/{exercise.id}/edit" class="btn-outline">Edit exercise</a>
			{/if}
		</div>

		{#if exercise.description}
			<div class="card">
				<p class="description">{exercise.description}</p>
			</div>
		{/if}

		<div class="card">
			<h2>Instructions</h2>
			<div class="instructions">{exercise.instructions}</div>
		</div>

		{#if exercise.template_code}
			<div class="card">
				<h2>Starter code</h2>
				<pre class="code-block"><code>{exercise.template_code}</code></pre>
			</div>
		{/if}

		<!-- Test cases -->
		<div class="card">
			<div class="tc-header">
				<h2>Test cases ({testCases.length})</h2>
				{#if canManage}
					<button class="btn-sm" onclick={() => (showTcForm = !showTcForm)}>
						{showTcForm ? '✕ Cancel' : '+ Add test case'}
					</button>
				{/if}
			</div>

			{#if showTcForm}
				<div class="tc-form">
					{#if tcError}<div class="alert">{tcError}</div>{/if}
					<form onsubmit={handleAddTestCase}>
						<label>
							Input <span class="hint">(stdin, leave blank if none)</span>
							<textarea bind:value={tcInput} rows="3" placeholder=""></textarea>
						</label>
						<label>
							Expected output <span class="req">*</span>
							<textarea bind:value={tcExpected} rows="3" placeholder="hello world" required></textarea>
						</label>
						<div class="tc-form-row">
							<label class="inline-label">
								<input type="checkbox" bind:checked={tcHidden} />
								Hidden from students
							</label>
							<label class="inline-label">
								Order
								<input type="number" bind:value={tcOrdinal} min="1" style="width:70px" />
							</label>
						</div>
						<div class="form-actions">
							<button type="submit" class="btn-primary" disabled={tcLoading}>
								{tcLoading ? 'Adding…' : 'Add test case'}
							</button>
						</div>
					</form>
				</div>
			{/if}

			{#if testCases.length === 0}
				<p class="empty-text">No test cases yet.</p>
			{:else}
				<div class="tc-list">
					{#each testCases as tc, i (tc.id)}
						<div class="tc-item {tc.is_hidden ? 'tc-hidden' : ''}">
							<div class="tc-meta">
								<span class="tc-num">#{i + 1}</span>
								{#if tc.is_hidden}
									<span class="badge draft">hidden</span>
								{/if}
							</div>

							{#if isVisible(tc) || canSeeAll}
								{#if isVisible(tc)}
									<div class="tc-io">
										<div class="io-block">
											<span class="io-label">Input</span>
											<pre class="io-pre">{tc.input || '(empty)'}</pre>
										</div>
										<div class="io-block">
											<span class="io-label">Expected output</span>
											<pre class="io-pre">{tc.expected_output}</pre>
										</div>
									</div>
								{:else}
									<!-- privileged user viewing a hidden case -->
									<div class="tc-io">
										<div class="io-block">
											<span class="io-label">Input</span>
											<pre class="io-pre">{(tc as unknown as TestCase).input || '(empty)'}</pre>
										</div>
										<div class="io-block">
											<span class="io-label">Expected output</span>
											<pre class="io-pre">{(tc as unknown as TestCase).expected_output}</pre>
										</div>
									</div>
								{/if}
							{:else}
								<p class="tc-hidden-msg">Hidden test case — not visible to students</p>
							{/if}

							{#if canManage}
								<button
									class="tc-delete"
									onclick={() => handleDeleteTestCase(tc.id)}
									disabled={deletingTc[tc.id]}
								>✕</button>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>
		<!-- Submit solution -->
		<div class="card">
			<h2>Submit your solution</h2>

			{#if submitError}
				<div class="alert">{submitError}</div>
			{/if}

			{#if latestSubmission}
				<div class="result-banner {statusCls(latestSubmission.status)}">
					<span class="result-label">{statusLabel(latestSubmission.status)}</span>
					{#if latestSubmission.status === 'accepted'}
						<span class="result-score">{latestSubmission.score}%</span>
					{/if}
					{#if latestSubmission.stderr}
						<pre class="result-stderr">{latestSubmission.stderr}</pre>
					{/if}
				</div>
			{/if}

			<form onsubmit={handleSubmit}>
				<div class="submit-lang-row">
					<label class="inline-label">
						Language
						<select bind:value={submitLang}>
							{#each ['python','go','java','c','cpp','javascript'] as l}
								<option value={l}>{l}</option>
							{/each}
						</select>
					</label>
				</div>
				<textarea
					bind:value={code}
					rows="14"
					placeholder="Write your solution here…"
					class="code-editor"
					spellcheck="false"
				></textarea>
				<div class="form-actions">
					<button type="submit" class="btn-primary" disabled={submitting}>
						{submitting ? 'Submitting…' : 'Submit'}
					</button>
				</div>
			</form>
		</div>

		<!-- Submission history -->
		{#if submissionHistory.length > 0}
			<div class="card">
				<button class="history-toggle" onclick={() => (showHistory = !showHistory)}>
					Submission history ({submissionHistory.length})
					<span>{showHistory ? '▲' : '▼'}</span>
				</button>

				{#if showHistory}
					<div class="history-list">
						{#each submissionHistory as sub (sub.id)}
							<div class="history-row">
								<span class="history-badge {statusCls(sub.status)}">{statusLabel(sub.status)}</span>
								<span class="history-lang">{sub.language}</span>
								<span class="history-date">
									{new Date(sub.submitted_at).toLocaleString()}
								</span>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<style>
	.page { max-width: 800px; }

	.back-link {
		display: inline-block;
		font-size: 0.875rem;
		color: #6b7280;
		margin-bottom: 1.5rem;
	}
	.back-link:hover { color: #4f46e5; }

	.skeleton {
		height: 120px;
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
		margin-bottom: 1rem;
	}

	.ex-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	h1 {
		font-size: 1.75rem;
		font-weight: 700;
		line-height: 1.3;
		margin-bottom: 0.6rem;
	}

	.badges { display: flex; gap: 0.4rem; flex-wrap: wrap; }

	.badge {
		font-size: 0.72rem;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 99px;
		text-transform: capitalize;
	}
	.diff-easy   { background: #dcfce7; color: #166534; }
	.diff-medium { background: #fef9c3; color: #a16207; }
	.diff-hard   { background: #fee2e2; color: #b91c1c; }
	.lang   { background: #e0e7ff; color: #3730a3; }
	.draft  { background: #f3f4f6; color: #6b7280; }
	.neutral { background: #f3f4f6; color: #374151; }

	.card {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		padding: 1.25rem 1.5rem;
		margin-bottom: 1.25rem;
	}

	.card h2 {
		font-size: 1rem;
		font-weight: 600;
		margin-bottom: 0.85rem;
		color: #374151;
	}

	.description { color: #374151; line-height: 1.65; }

	.instructions {
		color: #374151;
		line-height: 1.7;
		white-space: pre-wrap;
	}

	.code-block {
		background: #1e1e2e;
		color: #cdd6f4;
		border-radius: 8px;
		padding: 1rem 1.25rem;
		overflow-x: auto;
		font-size: 0.875rem;
		line-height: 1.6;
		margin: 0;
	}

	.tc-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
	}

	.tc-form {
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		padding: 1rem;
		margin-bottom: 1rem;
		background: #f9fafb;
	}

	.tc-form form { display: flex; flex-direction: column; gap: 0.85rem; }

	.tc-form label {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
		font-size: 0.85rem;
		font-weight: 500;
		color: #374151;
	}

	.tc-form textarea, .tc-form input[type='number'] {
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 6px;
		font-size: 0.875rem;
		font-family: monospace;
		resize: vertical;
		outline: none;
		background: #fff;
	}

	.tc-form textarea:focus, .tc-form input:focus {
		border-color: #4f46e5;
		box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
	}

	.tc-form-row { display: flex; gap: 1.5rem; align-items: center; }

	.inline-label {
		display: flex !important;
		flex-direction: row !important;
		align-items: center;
		gap: 0.4rem !important;
		font-size: 0.85rem !important;
	}

	.hint { color: #9ca3af; font-weight: 400; font-size: 0.8rem; }
	.req  { color: #ef4444; }

	.form-actions { display: flex; justify-content: flex-end; }

	.empty-text { color: #6b7280; font-size: 0.9rem; }

	.tc-list { display: flex; flex-direction: column; gap: 0.75rem; }

	.tc-item {
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		padding: 0.85rem 1rem;
		position: relative;
		background: #fff;
	}

	.tc-item.tc-hidden { border-color: #d1d5db; background: #f9fafb; }

	.tc-meta {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.6rem;
	}

	.tc-num { font-size: 0.78rem; color: #9ca3af; font-weight: 600; }

	.tc-io { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }

	.io-block { display: flex; flex-direction: column; gap: 0.3rem; }

	.io-label { font-size: 0.75rem; font-weight: 600; color: #6b7280; text-transform: uppercase; letter-spacing: 0.04em; }

	.io-pre {
		background: #f3f4f6;
		border-radius: 6px;
		padding: 0.5rem 0.75rem;
		font-size: 0.82rem;
		font-family: monospace;
		white-space: pre-wrap;
		word-break: break-all;
		margin: 0;
		color: #1a1a2e;
	}

	.tc-hidden-msg { color: #9ca3af; font-size: 0.875rem; font-style: italic; }

	.tc-delete {
		position: absolute;
		top: 0.6rem;
		right: 0.6rem;
		background: none;
		border: none;
		color: #9ca3af;
		cursor: pointer;
		font-size: 0.85rem;
		padding: 2px 5px;
		border-radius: 4px;
		line-height: 1;
	}
	.tc-delete:hover { background: #fee2e2; color: #b91c1c; }
	.tc-delete:disabled { opacity: 0.4; cursor: not-allowed; }

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
	.btn-primary:hover:not(:disabled) { background: #4338ca; }
	.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

	.btn-sm {
		background: #4f46e5;
		color: #fff;
		border: none;
		border-radius: 7px;
		padding: 5px 12px;
		font-size: 0.82rem;
		font-weight: 600;
		cursor: pointer;
	}
	.btn-sm:hover { background: #4338ca; }

	/* ── Submission ── */
	.result-banner {
		border-radius: 8px;
		padding: 0.75rem 1rem;
		margin-bottom: 1rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		flex-wrap: wrap;
	}
	.result-label { font-weight: 700; font-size: 0.95rem; }
	.result-score { font-size: 0.9rem; opacity: 0.85; }
	.result-stderr {
		width: 100%;
		margin: 0;
		font-size: 0.8rem;
		font-family: monospace;
		white-space: pre-wrap;
		word-break: break-all;
		opacity: 0.9;
	}

	.status-pending  { background: #f3f4f6; color: #374151; }
	.status-running  { background: #eff6ff; color: #1d4ed8; }
	.status-accepted { background: #dcfce7; color: #166534; }
	.status-wrong    { background: #fef9c3; color: #a16207; }
	.status-error    { background: #fee2e2; color: #b91c1c; }

	.submit-lang-row { margin-bottom: 0.75rem; }

	.inline-label {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}

	.inline-label select {
		padding: 0.35rem 0.65rem;
		border: 1px solid #d1d5db;
		border-radius: 6px;
		font-size: 0.875rem;
		outline: none;
		background: #fff;
	}

	.code-editor {
		width: 100%;
		padding: 0.75rem 1rem;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		font-size: 0.875rem;
		font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
		line-height: 1.6;
		resize: vertical;
		outline: none;
		background: #1e1e2e;
		color: #cdd6f4;
		transition: border-color 0.15s;
	}
	.code-editor:focus { border-color: #4f46e5; }

	.history-toggle {
		background: none;
		border: none;
		font-size: 0.9rem;
		font-weight: 600;
		color: #374151;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		padding: 0;
	}
	.history-toggle:hover { color: #4f46e5; }

	.history-list {
		margin-top: 0.85rem;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.history-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.5rem 0;
		border-bottom: 1px solid #f3f4f6;
		font-size: 0.875rem;
	}
	.history-row:last-child { border-bottom: none; }

	.history-badge {
		font-size: 0.75rem;
		font-weight: 700;
		padding: 2px 8px;
		border-radius: 99px;
	}
	.history-lang { color: #6b7280; font-size: 0.82rem; }
	.history-date { color: #9ca3af; font-size: 0.8rem; margin-left: auto; }

	.btn-outline {
		background: transparent;
		color: #374151;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
		font-weight: 500;
		text-decoration: none;
		transition: all 0.15s;
		white-space: nowrap;
	}
	.btn-outline:hover { background: #f9fafb; text-decoration: none; }
</style>
