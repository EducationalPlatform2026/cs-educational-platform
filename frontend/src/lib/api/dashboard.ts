import { apiFetch } from './client';

export interface RecentSubmission {
	id: string;
	exercise_id: string;
	exercise_title: string;
	language: string;
	status: string;
	score: number;
	submitted_at: string;
}

export interface DashboardStats {
	enrolled_courses: number;
	exercises_attempted: number;
	total_submissions: number;
	accepted_submissions: number;
	acceptance_rate: number;
	recent_submissions: RecentSubmission[];
}

export function getDashboard(): Promise<DashboardStats> {
	return apiFetch('/dashboard');
}
