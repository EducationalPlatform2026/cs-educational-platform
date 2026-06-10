<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth.svelte';

	const LANGS = ['python', 'javascript', 'go', 'java', 'c', 'cpp'];
	const LS_KEY = 'cp_sandbox';

	let language = $state('python');
	let code = $state('');
	let notes = $state('');
	let saved = $state(false);

	const STARTERS: Record<string, string> = {
		python: '# Write your Python code here\n\ndef main():\n    print("Hello, World!")\n\nmain()\n',
		javascript: '// Write your JavaScript here\n\nfunction main() {\n  console.log("Hello, World!");\n}\n\nmain();\n',
		go: 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("Hello, World!")\n}\n',
		java: 'public class Main {\n    public static void main(String[] args) {\n        System.out.println("Hello, World!");\n    }\n}\n',
		c: '#include <stdio.h>\n\nint main() {\n    printf("Hello, World!\\n");\n    return 0;\n}\n',
		cpp: '#include <iostream>\n\nint main() {\n    std::cout << "Hello, World!" << std::endl;\n    return 0;\n}\n'
	};

	onMount(() => {
		try {
			const raw = localStorage.getItem(LS_KEY);
			if (raw) {
				const data = JSON.parse(raw);
				language = data.language ?? 'python';
				code = data.code ?? '';
				notes = data.notes ?? '';
			} else {
				code = STARTERS[language];
			}
		} catch {
			code = STARTERS[language];
		}
	});

	function saveSnippet() {
		localStorage.setItem(LS_KEY, JSON.stringify({ language, code, notes }));
		saved = true;
		setTimeout(() => { saved = false; }, 2000);
	}

	function clearAll() {
		if (!confirm('Clear all code and notes?')) return;
		code = STARTERS[language];
		notes = '';
		localStorage.removeItem(LS_KEY);
	}

	function handleLangChange() {
		if (!code.trim() || code === STARTERS[language]) {
			code = STARTERS[language] ?? '';
		}
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

	// Update language starter when language changes
	$effect(() => {
		void language; // track dependency
		handleLangChange();
	});
</script>

<div class="page">
	<!-- Header -->
	<div class="page-header">
		<div>
			<h1>Sandbox</h1>
			<p class="sub">Scratch pad for writing and planning code — save snippets locally</p>
		</div>
		<div class="header-actions">
			<button class="btn-save" onclick={saveSnippet} class:saved>
				{saved ? '✓ Saved' : '💾 Save'}
			</button>
			<button class="btn-clear" onclick={clearAll}>Clear</button>
		</div>
	</div>

	<!-- Info banner -->
	<div class="info-banner">
		<span class="info-icon">💡</span>
		<span>Sandbox saves your code locally. To run code against test cases, <a href="/courses">open an exercise</a> from any course.</span>
	</div>

	<!-- Split pane -->
	<div class="split">
		<!-- Left: notes -->
		<div class="pane pane-notes">
			<div class="pane-header">
				<span class="pane-title">📝 Notes</span>
				<span class="pane-hint">Plan your approach, pseudocode, ideas</span>
			</div>
			<textarea
				bind:value={notes}
				class="notes-editor"
				placeholder="Write your notes, pseudocode, or algorithm outline here…"
				spellcheck="true"
			></textarea>
		</div>

		<!-- Right: code editor -->
		<div class="pane pane-code">
			<div class="pane-header">
				<span class="pane-title">⌨️ Code Editor</span>
				<select bind:value={language} class="lang-select">
					{#each LANGS as l}
						<option value={l}>{l}</option>
					{/each}
				</select>
			</div>
			<textarea
				bind:value={code}
				onkeydown={handleKeydown}
				class="code-editor"
				placeholder="Write your code here…"
				spellcheck="false"
			></textarea>
			<div class="editor-footer">
				<span class="shortcut">Tab → indent · Ctrl+S → save mentally 😄</span>
				<a href="/courses" class="cta-link">Browse exercises to test your code →</a>
			</div>
		</div>
	</div>

	<!-- Quick links -->
	<div class="quick-links">
		<p class="ql-title">Ready to solve some problems?</p>
		<div class="ql-row">
			<a href="/courses" class="ql-card">
				<span class="ql-icon">📚</span>
				<span class="ql-label">Browse Courses</span>
			</a>
			<a href="/dashboard" class="ql-card">
				<span class="ql-icon">📊</span>
				<span class="ql-label">My Dashboard</span>
			</a>
		</div>
	</div>
</div>

<style>
	.page { max-width: 1100px; }

	.page-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		margin-bottom: 1rem;
		gap: 1rem;
	}

	h1 { font-size: 1.75rem; font-weight: 700; }

	.sub { color: #6b7280; font-size: 0.875rem; margin-top: 0.2rem; }

	.header-actions { display: flex; gap: 0.5rem; align-items: center; }

	.btn-save {
		background: #7c3aed;
		color: #fff;
		border: none;
		border-radius: 8px;
		padding: 0.5rem 1.1rem;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.15s;
		font-family: inherit;
	}
	.btn-save:hover { background: #6d28d9; }
	.btn-save.saved { background: #16a34a; }

	.btn-clear {
		background: transparent;
		border: 1px solid #e5e7eb;
		color: #6b7280;
		border-radius: 8px;
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
		cursor: pointer;
		font-family: inherit;
	}
	.btn-clear:hover { border-color: #d1d5db; color: #374151; }

	.info-banner {
		background: #ede9fe;
		border: 1px solid #ddd6fe;
		border-radius: 10px;
		padding: 0.7rem 1rem;
		font-size: 0.875rem;
		color: #5b21b6;
		display: flex;
		align-items: center;
		gap: 0.6rem;
		margin-bottom: 1.25rem;
	}
	.info-banner a { color: #7c3aed; font-weight: 600; }

	.split {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.pane {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.pane-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem 1rem;
		border-bottom: 1px solid #f3f4f6;
		background: #fafafa;
	}

	.pane-title { font-size: 0.875rem; font-weight: 600; color: #374151; }

	.pane-hint { font-size: 0.78rem; color: #9ca3af; }

	.lang-select {
		padding: 3px 8px;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		font-size: 0.82rem;
		color: #374151;
		background: #fff;
		outline: none;
		cursor: pointer;
	}

	.notes-editor {
		flex: 1;
		width: 100%;
		min-height: 420px;
		padding: 1rem;
		border: none;
		font-size: 0.875rem;
		line-height: 1.7;
		font-family: inherit;
		color: #374151;
		resize: none;
		outline: none;
		background: #fff;
	}

	.code-editor {
		flex: 1;
		width: 100%;
		min-height: 380px;
		padding: 1rem;
		border: none;
		font-size: 0.875rem;
		line-height: 1.65;
		font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
		color: #cdd6f4;
		background: #1e1e2e;
		resize: none;
		outline: none;
	}

	.editor-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.5rem 0.85rem;
		border-top: 1px solid #2a2a3e;
		background: #1e1e2e;
	}

	.shortcut { font-size: 0.72rem; color: #6272a4; font-family: monospace; }

	.cta-link {
		font-size: 0.78rem;
		color: #bd93f9;
		font-weight: 500;
		text-decoration: none;
	}
	.cta-link:hover { text-decoration: underline; }

	.quick-links {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		padding: 1.25rem;
	}

	.ql-title {
		font-size: 0.9rem;
		font-weight: 600;
		color: #374151;
		margin-bottom: 0.85rem;
	}

	.ql-row { display: flex; gap: 0.75rem; }

	.ql-card {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.65rem 1.1rem;
		background: #faf8ff;
		border: 1px solid #ddd6fe;
		border-radius: 10px;
		font-size: 0.875rem;
		font-weight: 500;
		color: #7c3aed;
		text-decoration: none;
		transition: all 0.15s;
	}
	.ql-card:hover { background: #ede9fe; text-decoration: none; }

	.ql-icon { font-size: 1.1rem; }

	@media (max-width: 700px) {
		.split { grid-template-columns: 1fr; }
	}
</style>
