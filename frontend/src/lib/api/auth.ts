import { apiFetch } from './client';

export interface AuthResponse {
	token: string;
	user_id: string;
	role: string;
	first_name: string;
	last_name: string;
}

export function login(email: string, password: string): Promise<AuthResponse> {
	return apiFetch('/auth/login', {
		method: 'POST',
		body: JSON.stringify({ email, password })
	});
}

export function register(
	email: string,
	password: string,
	firstName: string,
	lastName: string,
	role: string
): Promise<AuthResponse> {
	return apiFetch('/auth/register', {
		method: 'POST',
		body: JSON.stringify({
			email,
			password,
			first_name: firstName,
			last_name: lastName,
			role
		})
	});
}
