import type { DashboardStats } from '$lib/api/dashboard';

interface GamificationData {
	xp: number;
	streak: number;
	solvedExercises: string[];
}

class UserStore {
	xp = $state(0);
	streak = $state(0);
	solvedExercises = $state<Set<string>>(new Set());

	get level(): number {
		return Math.floor(this.xp / 500) + 1;
	}

	get levelProgress(): number {
		return ((this.xp % 500) / 500) * 100;
	}

	get xpIntoLevel(): number {
		return this.xp % 500;
	}

	init() {
		try {
			const raw = localStorage.getItem('cp_gamification');
			if (raw) {
				const data: GamificationData = JSON.parse(raw);
				this.xp = data.xp ?? 0;
				this.streak = data.streak ?? 0;
				this.solvedExercises = new Set(data.solvedExercises ?? []);
			}
		} catch {
			// ignore corrupt data
		}
	}

	rehydrateFromDashboard(stats: DashboardStats) {
		this.xp = (stats.accepted_submissions ?? 0) * 100;
		this.streak = this._computeStreak(stats.recent_submissions ?? []);
		this._persist();
	}

	addXP(amount: number) {
		this.xp += amount;
		this._persist();
	}

	markSolved(exerciseId: string) {
		this.solvedExercises = new Set([...this.solvedExercises, exerciseId]);
		this._persist();
	}

	isSolved(exerciseId: string): boolean {
		return this.solvedExercises.has(exerciseId);
	}

	reset() {
		this.xp = 0;
		this.streak = 0;
		this.solvedExercises = new Set();
		localStorage.removeItem('cp_gamification');
	}

	private _computeStreak(submissions: Array<{ submitted_at: string; status: string }>): number {
		const acceptedDays = new Set(
			submissions
				.filter((s) => s.status === 'accepted')
				.map((s) => new Date(s.submitted_at).toDateString())
		);
		if (acceptedDays.size === 0) return 0;
		let streak = 0;
		const today = new Date();
		for (let i = 0; i < 30; i++) {
			const d = new Date(today);
			d.setDate(d.getDate() - i);
			if (acceptedDays.has(d.toDateString())) {
				streak++;
			} else {
				break;
			}
		}
		return streak;
	}

	private _persist() {
		const data: GamificationData = {
			xp: this.xp,
			streak: this.streak,
			solvedExercises: [...this.solvedExercises]
		};
		localStorage.setItem('cp_gamification', JSON.stringify(data));
	}
}

export const userStore = new UserStore();
