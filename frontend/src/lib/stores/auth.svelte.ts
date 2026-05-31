export interface User {
	user_id: string;
	role: string;
	first_name: string;
	last_name: string;
}

class AuthStore {
	token = $state<string | null>(null);
	user = $state<User | null>(null);

	get isLoggedIn(): boolean {
		return this.token !== null;
	}

	/** Restore session from localStorage (call once on app start). */
	init() {
		this.token = localStorage.getItem('token');
		const raw = localStorage.getItem('user');
		this.user = raw ? (JSON.parse(raw) as User) : null;
	}

	/** Persist a successful login/register response. */
	set(token: string, user: User) {
		this.token = token;
		this.user = user;
		localStorage.setItem('token', token);
		localStorage.setItem('user', JSON.stringify(user));
	}

	/** Clear session. */
	logout() {
		this.token = null;
		this.user = null;
		localStorage.removeItem('token');
		localStorage.removeItem('user');
	}
}

export const auth = new AuthStore();
