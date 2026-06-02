<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { getSubmission, type SubmissionDetail, type SubmissionStatus } from '$lib/api/submissions';
	import { getStatusMeta } from '$lib/utils/submissionStatus';

	const id = $derived($page.params.id as string);

	let sub = $state<SubmissionDetail | null>(null);
	let loading = $state(true);
	let error = $state('');

	const statusInfo = getStatusMeta;

	function formatDate(iso: string) {
		return new Date(iso).toLocaleString(undefined, {
			year: 'numeric', month: 'short', day: 'numeric',
			hour: '2-digit', minute: '2-digit', second: '2-digit'
		});
	}

	onMount(async () => {
		try {
			sub = await getSubmission(id);
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Failed to load submission';
		} finally {
			loading = false;
		}
	});
</script>

<div class="page">
	<a href="/dashboard" class="back-link">← Dashboard</a>

	{#if loading}
		<div class="skeleton-header"></div>
		<div class="skeleton-body"></div>

	{:else if error}
		<div class="alert">{error}</div>

	{:else if sub}
		{@const meta = statusInfo(sub.status)}

		<div class="submission-header">
			<div>
				<h1>Submission</h1>
				<p class="meta">
					<a href="/exercises/{sub.exercise_id}" class="ex-link">View exercise</a>
					· {formatDate(sub.submitted_at)}
				</p>
			</div>
			<div class="header-right">
				<span class="status-badge {meta.cls}">{meta.label}</span>
				{#if sub.status !== 'pending' && sub.status !== 'running'}
					<span class="score">{sub.score.toFixed(0)}%</span>
				{/if}
			</div>
		</div>

		<!-- Stderr / compiler output -->
		{#if sub.stderr}
			<div class="error-block">
				<div class="block-title">Compiler / runtime output</div>
				<pre class="error-pre">{sub.stderr}</pre>
			</div>
		{/if}

		<!-- Submitted code -->
		<div class="code-block">
			<div class="block-title">
				Submitted code
				<span class="lang-badge">{sub.language}</span>
			</div>
			<pre class="code-pre">{sub.code}</pre>
		</div>

		<!-- Test case results -->
		{#if sub.results.length > 0}
			<section class="results-section">
				<h2>Test case results ({sub.results.length})</h2>
				<div class="table-wrap">
					<table class="results-table">
						<thead>
							<tr>
								<th>#</th>
								<th>Status</th>
								<th>Runtime</th>
								<th>Actual output</th>
							</tr>
						</thead>
						<tbody>
							{#each sub.results as r, i (r.id)}
								{@const rm = statusInfo(r.status as SubmissionStatus)}
								<tr>
									<td class="td-num">{i + 1}</td>
									<td><span class="status-badge {rm.cls}">{rm.label}</span></td>
									<td class="td-runtime">
										{r.runtime_ms != null ? r.runtime_ms + ' ms' : '—'}
									</td>
									<td class="td-output">
										{#if r.actual_output != null}
											<code>{r.actual_output.slice(0, 200)}{r.actual_output.length > 200 ? '…' : ''}</code>
										{:else}
											<span class="muted">—</span>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>
		{:else if sub.status !== 'pending' && sub.status !== 'running'}
			<p class="muted-note">No per-test-case results recorded.</p>
		{/if}
	{/if}
</div>

<style>
	.page { max-width: 860px; }

	.back-link {
		display: inline-block;
		font-size: 0.875rem;
		color: #6b7280;
		margin-bottom: 1.5rem;
	}
	.back-link:hover { color: #4f46e5; }

	/* ── Header ── */
	.submission-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.75rem;
	}

	h1 { font-size: 1.6rem; font-weight: 700; }

	.meta {
		font-size: 0.85rem;
		color: #6b7280;
		margin-top: 0.3rem;
	}

	.ex-link { color: #4f46e5; }
	.ex-link:hover { text-decoration: underline; }

	.header-right {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-shrink: 0;
	}

	.score {
		font-size: 1.5rem;
		font-weight: 700;
		color: #1a1a2e;
	}

	/* ── Code / error blocks ── */
	.error-block, .code-block {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		margin-bottom: 1.5rem;
		overflow: hidden;
	}

	.block-title {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.65rem 1rem;
		font-size: 0.8rem;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
	}

	.error-pre {
		margin: 0;
		padding: 1rem;
		font-size: 0.82rem;
		line-height: 1.6;
		color: #b91c1c;
		white-space: pre-wrap;
		word-break: break-word;
		background: #fef2f2;
	}

	.code-pre {
		margin: 0;
		padding: 1rem;
		font-size: 0.85rem;
		line-height: 1.65;
		color: #e2e8f0;
		background: #1e1e2e;
		white-space: pre;
		overflow-x: auto;
	}

	.lang-badge {
		background: #e0e7ff;
		color: #3730a3;
		font-size: 0.7rem;
		font-weight: 600;
		padding: 2px 7px;
		border-radius: 99px;
		text-transform: capitalize;
	}

	/* ── Results table ── */
	.results-section { margin-top: 0.5rem; }

	.results-section h2 {
		font-size: 1.05rem;
		font-weight: 600;
		margin-bottom: 0.75rem;
	}

	.table-wrap {
		border: 1px solid #e5e7eb;
		border-radius: 10px;
		overflow: hidden;
	}

	.results-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.results-table th {
		text-align: left;
		padding: 0.6rem 1rem;
		font-size: 0.75rem;
		font-weight: 600;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
	}

	.results-table td {
		padding: 0.7rem 1rem;
		border-bottom: 1px solid #f3f4f6;
		vertical-align: top;
	}

	.results-table tr:last-child td { border-bottom: none; }

	.td-num { color: #9ca3af; font-size: 0.8rem; width: 2rem; }
	.td-runtime { color: #6b7280; font-size: 0.82rem; white-space: nowrap; }
	.td-output code {
		font-size: 0.8rem;
		background: #f3f4f6;
		padding: 2px 6px;
		border-radius: 4px;
		word-break: break-all;
	}

	/* ── Status badges ── */
	.status-badge {
		font-size: 0.75rem;
		font-weight: 600;
		padding: 3px 9px;
		border-radius: 99px;
		white-space: nowrap;
	}
	.status-accepted { background: #dcfce7; color: #166534; }
	.status-pending  { background: #f3f4f6; color: #6b7280; }
	.status-running  { background: #fef9c3; color: #a16207; }
	.status-wrong    { background: #fef2f2; color: #b91c1c; }
	.status-error    { background: #fef2f2; color: #b91c1c; }

	/* ── Misc ── */
	.muted { color: #9ca3af; }
	.muted-note { color: #9ca3af; font-size: 0.875rem; margin-top: 1rem; }

	.alert {
		background: #fef2f2;
		color: #b91c1c;
		border: 1px solid #fecaca;
		border-radius: 8px;
		padding: 0.75rem 1rem;
		font-size: 0.9rem;
	}

	/* ── Skeletons ── */
	.skeleton-header {
		height: 80px;
		border-radius: 10px;
		margin-bottom: 1.5rem;
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
	}
	.skeleton-body {
		height: 260px;
		border-radius: 10px;
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
	}
	@keyframes shimmer {
		0%   { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}
</style>
