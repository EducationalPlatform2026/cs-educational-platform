<script lang="ts">
	import { goto } from '$app/navigation';
	import { register } from '$lib/api/auth';
	import { auth } from '$lib/stores/auth.svelte';

	let firstName = $state('');
	let lastName = $state('');
	let email = $state('');
	let password = $state('');
	let role = $state('student');
	let error = $state('');
	let loading = $state(false);

	const roles = [
		{ value: 'student', label: 'Student' },
		{ value: 'teaching_assistant', label: 'Teaching Assistant' },
		{ value: 'professor', label: 'Professor' }
	];

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const res = await register(email, password, firstName, lastName, role);
			auth.set(res.token, {
				user_id: res.user_id,
				role: res.role,
				first_name: res.first_name,
				last_name: res.last_name
			});
			goto('/courses');
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Registration failed';
		} finally {
			loading = false;
		}
	}
</script>

<div class="auth-page">
	<div class="auth-card">
		<h1>Create account</h1>
		<p class="subtitle">Join CS Educational Platform today</p>

		{#if error}
			<div class="alert">{error}</div>
		{/if}

		<form onsubmit={handleSubmit}>
			<div class="row">
				<label>
					First name
					<input
						type="text"
						bind:value={firstName}
						placeholder="Jane"
						required
						autocomplete="given-name"
					/>
				</label>
				<label>
					Last name
					<input
						type="text"
						bind:value={lastName}
						placeholder="Doe"
						required
						autocomplete="family-name"
					/>
				</label>
			</div>

			<label>
				Email
				<input
					type="email"
					bind:value={email}
					placeholder="you@example.com"
					required
					autocomplete="email"
				/>
			</label>

			<label>
				Password
				<input
					type="password"
					bind:value={password}
					placeholder="••••••••"
					required
					minlength="8"
					autocomplete="new-password"
				/>
			</label>

			<label>
				Role
				<select bind:value={role}>
					{#each roles as r}
						<option value={r.value}>{r.label}</option>
					{/each}
				</select>
			</label>

			<button type="submit" class="btn-primary" disabled={loading}>
				{loading ? 'Creating account…' : 'Create account'}
			</button>
		</form>

		<p class="switch-link">
			Already have an account? <a href="/auth/login">Sign in</a>
		</p>
	</div>
</div>

<style>
	.auth-page {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: calc(100vh - 56px);
		padding: 2rem;
	}

	.auth-card {
		background: #fff;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		padding: 2.5rem;
		width: 100%;
		max-width: 640px;
		box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
	}

	h1 {
		font-size: 1.6rem;
		font-weight: 700;
		margin-bottom: 0.25rem;
	}

	.subtitle {
		color: #6b7280;
		font-size: 0.9rem;
		margin-bottom: 1.75rem;
	}

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

	.row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.75rem;
	}

	label {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
	}

	input,
	select {
		padding: 0.6rem 0.85rem;
		border: 1px solid #d1d5db;
		border-radius: 8px;
		font-size: 0.95rem;
		transition: border-color 0.15s;
		outline: none;
		background: #fff;
	}

	input:focus,
	select:focus {
		border-color: #4f46e5;
		box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
	}

	.btn-primary {
		margin-top: 0.5rem;
		background: #4f46e5;
		color: #fff;
		border: none;
		border-radius: 8px;
		padding: 0.7rem;
		font-size: 0.95rem;
		font-weight: 600;
		cursor: pointer;
		transition: background 0.15s;
	}

	.btn-primary:hover:not(:disabled) {
		background: #4338ca;
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.switch-link {
		margin-top: 1.25rem;
		text-align: center;
		font-size: 0.875rem;
		color: #6b7280;
	}
</style>
