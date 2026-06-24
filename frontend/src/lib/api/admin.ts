import { apiFetch } from './client';

export interface AdminUser {
	id: string;
	email: string;
	first_name: string;
	last_name: string;
	role: string;
	is_active: boolean;
	created_at: string;
}

export interface AdminCourse {
	id: string;
	title: string;
	created_by: string;
	creator_name: string;
	is_published: boolean;
	enroll_count: number;
	exercise_count: number;
	created_at: string;
}

export interface AdminStats {
	total_users: number;
	total_courses: number;
	total_exercises: number;
	total_submissions: number;
	active_users: number;
	published_courses: number;
}

export const getAdminStats  = () => apiFetch<AdminStats>('/admin/stats');
export const listAdminUsers = () => apiFetch<AdminUser[]>('/admin/users');
export const listAdminCourses = () => apiFetch<AdminCourse[]>('/admin/courses');

export const updateUserRole = (userId: string, role: string) =>
	apiFetch<{ role: string }>(`/admin/users/${userId}/role`, {
		method: 'PATCH',
		body: JSON.stringify({ role })
	});

export const toggleUserActive = (userId: string) =>
	apiFetch<{ is_active: boolean }>(`/admin/users/${userId}/active`, {
		method: 'PATCH'
	});
