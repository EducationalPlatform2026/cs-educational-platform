<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import {
		getExercise, listTestCases, createTestCase, deleteTestCase,
		type Exercise, type AnyTestCase, type TestCase
	} from '$lib/api/exercises';
	import {
		submitCode, listSubmissions, getSubmission,
		type Submission, type SubmissionDetail
	} from '$lib/api/submissions';
	import { ApiError } from '$lib/api/client';
	import { getStatusMeta } from '$lib/utils/submissionStatus';
	import { auth } from '$lib/stores/auth.svelte';
	import { userStore } from '$lib/stores/userStore.svelte';
	import DifficultyDot from '$lib/components/DifficultyDot.svelte';

	const id = $derived($page.params.id as string);

	let exercise    = $state<Exercise | null>(null);
	let testCases   = $state<AnyTestCase[]>([]);
	let loading     = $state(true);
	let error       = $state('');
	let notEnrolled = $state(false);

	const canManage = $derived(
		auth.user?.role === 'professor' || auth.user?.role === 'teaching_assistant' || auth.user?.role === 'admin'
	);
	const canSeeAll = $derived(
		auth.user?.role === 'professor' || auth.user?.role === 'teaching_assistant' || auth.user?.role === 'admin'
	);

	// Add test case form
	let showTcForm  = $state(false);
	let tcInput     = $state('');
	let tcExpected  = $state('');
	let tcHidden    = $state(false);
	let tcOrdinal   = $state(0);
	let tcLoading   = $state(false);
	let tcError     = $state('');
	let deletingTc  = $state<Record<string, boolean>>({});

	// Submission state
	let code              = $state('');
	let submitLang        = $state('');
	let submitting        = $state(false);
	let polling           = $state(false);
	let submitError       = $state('');
	let latestDetail      = $state<SubmissionDetail | null>(null);
	let submissionHistory = $state<Submission[]>([]);
	let showHistory       = $state(false);

	// Gamification
	let showXpFloat = $state(false);
	let xpAmount    = $state(100);
	let alreadySolved = $state(false);

	const xpReward = $derived(exercise?.difficulty === 'hard' ? 200 : 100);

	const visibleTc   = $derived(testCases.filter((tc): tc is TestCase => !tc.is_hidden));
	const hiddenCount = $derived(testCases.filter((tc) => tc.is_hidden).length);

	function diffClass(d: string) {
		return d === 'easy' ? 'diff-easy' : d === 'hard' ? 'diff-hard' : 'diff-medium';
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Tab') {
			e.preventDefault();
			const ta = e.currentTarget as HTMLTextAreaElement;
			const s = ta.selectionStart;
			const end = ta.selectionEnd;
			code = code.substring(0, s) + '  ' + code.substring(end);
			requestAnimationFrame(() => { ta.selectionStart = ta.selectionEnd = s + 2; });
		}
	}

	async function pollUntilDone(subId: string): Promise<SubmissionDetail> {
		polling = true;
		try {
			for (let i = 0; i < 20; i++) {
				await new Promise((r) => setTimeout(r, 1500));
				const detail = await getSubmission(subId);
				if (detail.status !== 'pending' && detail.status !== 'running') return detail;
			}
			return await getSubmission(subId);
		} finally {
			polling = false;
		}
	}

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
			code = subs[0]?.code ?? ex.template_code ?? '';
			submitLang = ex.language;
			alreadySolved = userStore.isSolved(id);
			if (subs.length > 0) {
				latestDetail = await getSubmission(subs[0].id);
			}
		} catch (err: unknown) {
			if (err instanceof ApiError && err.status === 403) notEnrolled = true;
			else error = err instanceof Error ? err.message : 'Failed to load exercise';
		} finally {
			loading = false;
		}
	});

	async function handleAddTestCase(e: SubmitEvent) {
		e.preventDefault(); tcError = ''; tcLoading = true;
		try {
			const tc = await createTestCase(id, { input: tcInput, expected_output: tcExpected, is_hidden: tcHidden, ordinal: tcOrdinal });
			testCases = [...testCases, tc];
			tcInput = ''; tcExpected = ''; tcHidden = false; tcOrdinal = testCases.length + 1;
			showTcForm = false;
		} catch (err: unknown) {
			tcError = err instanceof Error ? err.message : 'Failed to add test case';
		} finally { tcLoading = false; }
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

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		submitError = ''; submitting = true; showXpFloat = false;
		try {
			const sub = await submitCode(id, code, submitLang);
			submissionHistory = [sub, ...submissionHistory];
			// Poll for result
			const detail = await pollUntilDone(sub.id);
			latestDetail = detail;
			submissionHistory = [detail, ...submissionHistory.slice(1)];
			// Award XP on accept (only once)
			if (detail.status === 'accepted' && !alreadySolved) {
				xpAmount = xpReward;
				userStore.addXP(xpAmount);
				userStore.markSolved(id);
				alreadySolved = true;
				showXpFloat = true;
				setTimeout(() => { showXpFloat = false; }, 2500);
			}
		} catch (err: unknown) {
			submitError = err instanceof Error ? err.message : 'Submission failed';
		} finally { submitting = false; }
	}
</script>

<!-- XP float animation -->
{#if showXpFloat}
	<div class="xp-float">+{xpAmount} XP!</div>
{/if}

<div class="page">
	<div class="topbar">
		<a href={exercise ? `/courses/${exercise.course_id}` : '/courses'} class="back-link">← Back to course</a>
		{#if canManage && exercise}
			<a href="/exercises/{exercise.id}/edit" class="btn-outline-sm">Edit exercise</a>
		{/if}
	</div>

	{#if loading}
		<div class="split">
			<div class="pane-left"><div class="skeleton" style="height:100%"></div></div>
			<div class="pane-right"><div class="skeleton" style="height:100%"></div></div>
		</div>

	{:else if notEnrolled}
		<div class="enroll-gate">
			<div class="gate-icon">🔒</div>
			<p class="gate-title">Enrollment required</p>
			<p class="gate-sub">You need to be enrolled in this course to view and solve exercises.</p>
			<a href="/courses" class="btn-primary">Browse courses</a>
		</div>

	{:else if error}
		<div class="alert">{error}</div>

	{:else if exercise}
		<div class="split">
			<!-- ══════ LEFT PANE ══════ -->
			<div class="pane-left">
				<!-- Exercise header -->
				<div class="ex-header">
					<div class="ex-title-row">
						<h1>{exercise.title}</h1>
						{#if alreadySolved}
							<span class="solved-badge">✓ Solved</span>
						{/if}
					</div>
					<div class="badges-row">
						<DifficultyDot difficulty={exercise.difficulty as 'easy'|'medium'|'hard'} />
						<span class="badge {diffClass(exercise.difficulty)}">{exercise.difficulty}</span>
						<span class="badge lang-badge">{exercise.language}</span>
						<span class="badge neutral-badge">{Math.round(exercise.time_limit_ms / 1000)}s limit</span>
						<span class="xp-badge">⚡ +{xpReward} XP</span>
					</div>
				</div>

				<!-- Description -->
				{#if exercise.description}
					<div class="card">
						<p class="description">{exercise.description}</p>
					</div>
				{/if}

				<!-- Instructions -->
				<div class="card">
					<h2>Instructions</h2>
					<div class="instructions">{exercise.instructions}</div>
				</div>

				<!-- Starter code -->
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
								{showTcForm ? '✕' : '+ Add'}
							</button>
						{/if}
					</div>

					{#if showTcForm}
						<div class="tc-form">
							{#if tcError}<div class="alert">{tcError}</div>{/if}
							<form onsubmit={handleAddTestCase}>
								<label>Input <textarea bind:value={tcInput} rows="2" placeholder="(empty for no stdin)"></textarea></label>
								<label><span>Expected output <span class="req">*</span></span><textarea bind:value={tcExpected} rows="2" required></textarea></label>
								<div class="row-inline">
									<label class="inline-lbl"><input type="checkbox" bind:checked={tcHidden} /> Hidden</label>
									<label class="inline-lbl">Order <input type="number" bind:value={tcOrdinal} min="1" style="width:60px" /></label>
								</div>
								<div class="form-foot"><button type="submit" class="btn-primary-sm" disabled={tcLoading}>{tcLoading ? '…' : 'Add'}</button></div>
							</form>
						</div>
					{/if}

					{#if testCases.length === 0}
						<p class="empty-text">No test cases yet.</p>
					{:else}
						<div class="tc-list">
							{#each testCases as tc, i (tc.id)}
								<div class="tc-item" class:tc-hidden={tc.is_hidden}>
									<div class="tc-meta">
										<span class="tc-num">#{i + 1}</span>
										{#if tc.is_hidden}<span class="badge draft-badge">hidden</span>{/if}
									</div>
									{#if !tc.is_hidden || canSeeAll}
										{@const visible = tc as TestCase}
										<div class="tc-io">
											<div class="io-block">
												<span class="io-label">Input</span>
												<pre class="io-pre">{visible.input || '(empty)'}</pre>
											</div>
											<div class="io-block">
												<span class="io-label">Expected</span>
												<pre class="io-pre">{visible.expected_output}</pre>
											</div>
										</div>
									{:else}
										<p class="hidden-msg">🔒 Hidden — not visible to students</p>
									{/if}
									{#if canManage}
										<button class="tc-del" onclick={() => handleDeleteTestCase(tc.id)} disabled={deletingTc[tc.id]}>✕</button>
									{/if}
								</div>
							{/each}
							{#if hiddenCount > 0 && !canSeeAll}
								<p class="hidden-count">+ {hiddenCount} hidden test case{hiddenCount !== 1 ? 's' : ''}</p>
							{/if}
						</div>
					{/if}
				</div>
			</div>

			<!-- ══════ RIGHT PANE ══════ -->
			<div class="pane-right">
				<div class="editor-panel">
					<!-- Language + header -->
					<div class="editor-header">
						<span class="editor-title">Your solution</span>
						<select bind:value={submitLang} class="lang-select">
							{#each ['python','go','java','c','cpp','javascript'] as l}
								<option value={l}>{l}</option>
							{/each}
						</select>
					</div>

					<!-- Code editor -->
					<form onsubmit={handleSubmit}>
						<textarea
							bind:value={code}
							onkeydown={handleKeydown}
							rows="16"
							class="code-editor"
							placeholder="Write your solution here…"
							spellcheck="false"
								></textarea>

						{#if submitError}
							<div class="alert" style="margin:0.75rem 0 0">{submitError}</div>
						{/if}

						<div class="submit-row">
							<button type="submit" class="btn-submit" disabled={submitting || polling}>
								{#if polling}
									<span class="spinner"></span> Running…
								{:else if submitting}
									Submitting…
								{:else}
									▶ Submit
								{/if}
							</button>
							{#if alreadySolved}
								<span class="solved-hint">✅ You solved this one!</span>
							{/if}
						</div>
					</form>
				</div>

				<!-- Result panel -->
				{#if latestDetail}
					{@const meta = getStatusMeta(latestDetail.status)}
					<div class="result-panel {meta.cls}">
						<div class="result-header">
							<span class="result-status {meta.cls}">{meta.label}</span>
							{#if latestDetail.status === 'accepted'}
								<span class="result-score">{latestDetail.score}% — all tests passed!</span>
							{:else if latestDetail.status !== 'pending' && latestDetail.status !== 'running'}
								<span class="result-score">{latestDetail.score}%</span>
							{/if}
						</div>

						{#if latestDetail.stderr}
							<pre class="result-stderr">{latestDetail.stderr}</pre>
						{/if}

						{#if latestDetail.results.length > 0}
							<div class="test-dots">
								{#each latestDetail.results as r}
									<span class="dot" class:pass={r.status === 'accepted'} class:fail={r.status !== 'accepted'}
										title="Test #{latestDetail.results.indexOf(r)+1}: {r.status}"></span>
								{/each}
							</div>
						{/if}
					</div>
				{/if}

				<!-- Submission history -->
				{#if submissionHistory.length > 0}
					<div class="history-card">
						<button class="history-toggle" onclick={() => (showHistory = !showHistory)}>
							<span>History ({submissionHistory.length})</span>
							<span>{showHistory ? '▲' : '▼'}</span>
						</button>
						{#if showHistory}
							<div class="history-list">
								{#each submissionHistory as sub (sub.id)}
									{@const m = getStatusMeta(sub.status)}
									<a href="/submissions/{sub.id}" class="history-row">
										<span class="hist-badge {m.cls}">{m.label}</span>
										<span class="hist-lang">{sub.language}</span>
										<span class="hist-date">{new Date(sub.submitted_at).toLocaleString()}</span>
									</a>
								{/each}
							</div>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	/* ── XP float ── */
	.xp-float {
		position: fixed;
		top: 45%;
		left: 50%;
		transform: translateX(-50%);
		font-size: 2.25rem;
		font-weight: 900;
		color: #7c3aed;
		pointer-events: none;
		z-index: 9999;
		animation: float-up 2.5s ease-out forwards;
		text-shadow: 0 3px 12px rgba(124,58,237,0.35);
		letter-spacing: -0.02em;
	}
	@keyframes float-up {
		0%   { opacity: 1; transform: translateX(-50%) translateY(0) scale(1.3); }
		60%  { opacity: 1; }
		100% { opacity: 0; transform: translateX(-50%) translateY(-90px) scale(1); }
	}

	/* ── Layout ── */
	.page { max-width: 1100px; }

	.topbar { display:flex; align-items:center; justify-content:space-between; margin-bottom:1rem; }
	.back-link { font-size:0.875rem; color:#6b7280; }
	.back-link:hover { color:#7c3aed; }

	.split {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.25rem;
		align-items: start;
	}

	/* ── Left pane ── */
	.pane-left { display:flex; flex-direction:column; gap:1rem; }

	.ex-header { margin-bottom:0.25rem; }
	.ex-title-row { display:flex; align-items:center; gap:0.75rem; flex-wrap:wrap; margin-bottom:0.6rem; }
	h1 { font-size:1.5rem; font-weight:700; line-height:1.3; }

	.solved-badge { background:#dcfce7; color:#16a34a; font-size:0.78rem; font-weight:700; padding:3px 10px; border-radius:99px; }

	.badges-row { display:flex; align-items:center; gap:0.4rem; flex-wrap:wrap; }
	.badge { font-size:0.72rem; font-weight:600; padding:2px 8px; border-radius:99px; text-transform:capitalize; }
	.diff-easy   { background:#dcfce7; color:#166534; }
	.diff-medium { background:#fef9c3; color:#a16207; }
	.diff-hard   { background:#fee2e2; color:#b91c1c; }
	.lang-badge  { background:#ede9fe; color:#5b21b6; }
	.neutral-badge { background:#f3f4f6; color:#374151; }
	.draft-badge { background:#f3f4f6; color:#6b7280; }
	.xp-badge {
		background: #fef3c7; color: #d97706;
		font-size: 0.72rem; font-weight: 700;
		padding: 3px 9px; border-radius: 99px;
	}

	.card { background:#fff; border:1px solid #e5e7eb; border-radius:12px; padding:1.1rem 1.25rem; }
	.card h2 { font-size:0.95rem; font-weight:600; margin-bottom:0.75rem; color:#374151; }
	.description { color:#374151; line-height:1.65; font-size:0.9rem; }
	.instructions { color:#374151; line-height:1.7; white-space:pre-wrap; font-size:0.9rem; }
	.code-block { background:#1e1e2e; color:#cdd6f4; border-radius:8px; padding:0.85rem 1rem; overflow-x:auto; font-size:0.82rem; line-height:1.6; margin:0; font-family:'JetBrains Mono',monospace; }

	/* ── Test cases ── */
	.tc-header { display:flex; align-items:center; justify-content:space-between; margin-bottom:0.75rem; }
	.tc-form { border:1px solid #e5e7eb; border-radius:8px; padding:0.85rem; margin-bottom:0.85rem; background:#fafafa; }
	.tc-form form { display:flex; flex-direction:column; gap:0.65rem; }
	.tc-form label { display:flex; flex-direction:column; gap:0.25rem; font-size:0.82rem; font-weight:500; color:#374151; }
	.tc-form textarea, .tc-form input[type='number'] { padding:0.45rem 0.65rem; border:1px solid #d1d5db; border-radius:6px; font-size:0.82rem; font-family:monospace; resize:vertical; outline:none; }
	.tc-form textarea:focus, .tc-form input:focus { border-color:#7c3aed; box-shadow:0 0 0 2px rgba(124,58,237,0.1); }
	.row-inline { display:flex; gap:1.25rem; align-items:center; }
	.inline-lbl { display:flex !important; flex-direction:row !important; align-items:center; gap:0.4rem !important; font-size:0.82rem !important; }
	.form-foot { display:flex; justify-content:flex-end; }
	.req { color:#ef4444; }
	.empty-text { color:#6b7280; font-size:0.875rem; }
	.tc-list { display:flex; flex-direction:column; gap:0.6rem; }
	.tc-item { border:1px solid #e5e7eb; border-radius:8px; padding:0.75rem; position:relative; }
	.tc-item.tc-hidden { background:#f9fafb; border-color:#d1d5db; }
	.tc-meta { display:flex; align-items:center; gap:0.4rem; margin-bottom:0.5rem; }
	.tc-num { font-size:0.75rem; color:#9ca3af; font-weight:600; }
	.tc-io { display:grid; grid-template-columns:1fr 1fr; gap:0.5rem; }
	.io-block { display:flex; flex-direction:column; gap:0.2rem; }
	.io-label { font-size:0.7rem; font-weight:600; color:#6b7280; text-transform:uppercase; letter-spacing:0.04em; }
	.io-pre { background:#f3f4f6; border-radius:5px; padding:0.4rem 0.6rem; font-size:0.78rem; font-family:monospace; white-space:pre-wrap; word-break:break-all; margin:0; }
	.hidden-msg { color:#9ca3af; font-size:0.82rem; font-style:italic; }
	.hidden-count { color:#9ca3af; font-size:0.82rem; text-align:center; padding:0.4rem 0; }
	.tc-del { position:absolute; top:0.5rem; right:0.5rem; background:none; border:none; color:#9ca3af; cursor:pointer; font-size:0.82rem; padding:2px 5px; border-radius:4px; line-height:1; }
	.tc-del:hover { background:#fee2e2; color:#b91c1c; }
	.tc-del:disabled { opacity:0.4; cursor:not-allowed; }

	/* ── Right pane ── */
	.pane-right { display:flex; flex-direction:column; gap:0.85rem; position:sticky; top:70px; }

	.editor-panel { background:#1e1e2e; border-radius:12px; overflow:hidden; border:1px solid #2a2a3e; }

	.editor-header { display:flex; align-items:center; justify-content:space-between; padding:0.65rem 0.85rem; background:#181825; border-bottom:1px solid #2a2a3e; }
	.editor-title { font-size:0.82rem; color:#6272a4; font-weight:600; }
	.lang-select { background:#2a2a3e; border:1px solid #383850; color:#cdd6f4; padding:3px 8px; border-radius:6px; font-size:0.82rem; outline:none; cursor:pointer; }

	.code-editor {
		width: 100%;
		padding: 0.85rem 1rem;
		border: none;
		font-size: 0.875rem;
		font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
		line-height: 1.65;
		resize: none;
		outline: none;
		background: #1e1e2e;
		color: #cdd6f4;
		display: block;
	}

	.submit-row { display:flex; align-items:center; gap:0.75rem; padding:0.75rem 0.85rem; border-top:1px solid #2a2a3e; }
	.btn-submit {
		background: #7c3aed; color: #fff; border: none; border-radius: 8px;
		padding: 0.5rem 1.4rem; font-size: 0.9rem; font-weight: 700;
		cursor: pointer; transition: background 0.15s; display: flex; align-items: center; gap: 6px;
		font-family: inherit;
	}
	.btn-submit:hover:not(:disabled) { background: #6d28d9; }
	.btn-submit:disabled { opacity: 0.65; cursor: not-allowed; }

	.spinner {
		width: 14px; height: 14px; border: 2px solid rgba(255,255,255,0.3);
		border-top-color: #fff; border-radius: 50%;
		animation: spin 0.7s linear infinite; flex-shrink: 0;
	}
	@keyframes spin { to { transform: rotate(360deg); } }

	.solved-hint { font-size: 0.82rem; color: #16a34a; }

	/* ── Result ── */
	.result-panel { background:#fff; border:1px solid #e5e7eb; border-radius:12px; padding:1rem; }
	.result-header { display:flex; align-items:center; gap:0.75rem; margin-bottom:0.5rem; }
	.result-status { font-size:0.9rem; font-weight:700; padding:3px 10px; border-radius:99px; }
	.result-score { font-size:0.875rem; color:#6b7280; }
	.result-stderr { font-size:0.78rem; font-family:monospace; white-space:pre-wrap; word-break:break-all; background:#fef2f2; color:#b91c1c; border-radius:6px; padding:0.6rem 0.75rem; margin-top:0.5rem; }
	.test-dots { display:flex; flex-wrap:wrap; gap:5px; margin-top:0.75rem; }
	.dot { width:14px; height:14px; border-radius:50%; flex-shrink:0; }
	.dot.pass { background:#10b981; }
	.dot.fail { background:#ef4444; }

	/* ── History ── */
	.history-card { background:#fff; border:1px solid #e5e7eb; border-radius:12px; padding:0.85rem 1rem; }
	.history-toggle { background:none; border:none; font-size:0.9rem; font-weight:600; color:#374151; cursor:pointer; display:flex; align-items:center; justify-content:space-between; width:100%; padding:0; font-family:inherit; }
	.history-toggle:hover { color:#7c3aed; }
	.history-list { margin-top:0.75rem; display:flex; flex-direction:column; gap:0.4rem; }
	.history-row { display:flex; align-items:center; gap:0.6rem; padding:0.4rem 0; border-bottom:1px solid #f3f4f6; font-size:0.82rem; text-decoration:none; }
	.history-row:last-child { border-bottom:none; }
	.history-row:hover { background:#fafafa; }
	.hist-badge { font-size:0.72rem; font-weight:700; padding:2px 7px; border-radius:99px; }
	.hist-lang { color:#6b7280; }
	.hist-date { color:#9ca3af; margin-left:auto; font-size:0.78rem; }

	/* ── Status badge colors ── */
	:global(.status-accepted) { background:#dcfce7; color:#166534; }
	:global(.status-pending)  { background:#f3f4f6; color:#6b7280; }
	:global(.status-running)  { background:#eff6ff; color:#1d4ed8; }
	:global(.status-wrong)    { background:#fef9c3; color:#a16207; }
	:global(.status-error)    { background:#fee2e2; color:#b91c1c; }

	/* ── Gates/Alerts ── */
	.enroll-gate { display:flex; flex-direction:column; align-items:center; gap:0.75rem; padding:4rem 2rem; border:2px dashed #ddd6fe; border-radius:14px; text-align:center; background:#faf8ff; }
	.gate-icon { font-size:2.5rem; }
	.gate-title { font-size:1.1rem; font-weight:600; color:#374151; }
	.gate-sub { font-size:0.875rem; color:#6b7280; max-width:380px; }
	.alert { background:#fef2f2; color:#b91c1c; border:1px solid #fecaca; border-radius:8px; padding:0.75rem 1rem; font-size:0.875rem; }

	/* ── Buttons ── */
	.btn-primary { background:#7c3aed; color:#fff; border:none; border-radius:8px; padding:0.55rem 1.2rem; font-size:0.9rem; font-weight:600; cursor:pointer; text-decoration:none; display:inline-block; font-family:inherit; }
	.btn-primary:hover { background:#6d28d9; text-decoration:none; }
	.btn-sm { background:#7c3aed; color:#fff; border:none; border-radius:6px; padding:4px 10px; font-size:0.8rem; font-weight:600; cursor:pointer; font-family:inherit; }
	.btn-sm:hover { background:#6d28d9; }
	.btn-primary-sm { background:#7c3aed; color:#fff; border:none; border-radius:6px; padding:5px 14px; font-size:0.82rem; font-weight:600; cursor:pointer; font-family:inherit; }
	.btn-primary-sm:hover:not(:disabled) { background:#6d28d9; }
	.btn-primary-sm:disabled { opacity:0.6; cursor:not-allowed; }
	.btn-outline-sm { background:transparent; color:#374151; border:1px solid #d1d5db; border-radius:7px; padding:4px 12px; font-size:0.82rem; font-weight:500; text-decoration:none; }
	.btn-outline-sm:hover { background:#f9fafb; text-decoration:none; }

	/* ── Skeleton ── */
	.skeleton { min-height:400px; border-radius:12px; background:linear-gradient(90deg,#f0f0f0 25%,#e8e8e8 50%,#f0f0f0 75%); background-size:200% 100%; animation:shimmer 1.4s infinite; }
	@keyframes shimmer { 0%{background-position:200% 0} 100%{background-position:-200% 0} }

	/* ── Responsive ── */
	@media (max-width: 768px) {
		.split { grid-template-columns: 1fr; }
		.pane-right { position: static; }
	}
</style>
