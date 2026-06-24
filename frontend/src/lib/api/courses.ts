import { apiFetch } from './client';

export interface Course {
	id: string;
	title: string;
	description?: string;
	created_by: string;
	is_published: boolean;
	is_enrolled: boolean;
	created_at: string;
	updated_at: string;
}

export interface Member {
	user_id: string;
	first_name: string;
	last_name: string;
	role: string;
	enrolled_at: string;
}

export interface LeaderboardEntry {
	rank: number;
	user_id: string;
	first_name: string;
	last_name: string;
	solved_count: number;
}

export function listCourses(): Promise<Course[]> {
	return apiFetch('/courses');
}

export function getCourse(id: string): Promise<Course> {
	return apiFetch(`/courses/${id}`);
}

export function createCourse(data: {
	title: string;
	description?: string;
	is_published?: boolean;
}): Promise<Course> {
	return apiFetch('/courses', { method: 'POST', body: JSON.stringify(data) });
}

export function updateCourse(
	id: string,
	data: { title?: string; description?: string; is_published?: boolean }
): Promise<Course> {
	return apiFetch(`/courses/${id}`, { method: 'PUT', body: JSON.stringify(data) });
}

export function deleteCourse(id: string): Promise<void> {
	return apiFetch(`/courses/${id}`, { method: 'DELETE' });
}

export function enrollCourse(id: string, role?: string): Promise<void> {
	return apiFetch(`/courses/${id}/enroll`, {
		method: 'POST',
		body: role ? JSON.stringify({ role }) : '{}'
	});
}

export function getCourseMembers(id: string): Promise<Member[]> {
	return apiFetch(`/courses/${id}/members`);
}

export function getCourseLeaderboard(id: string): Promise<LeaderboardEntry[]> {
	return apiFetch(`/courses/${id}/leaderboard`);
}

export function updateMemberRole(
	courseId: string,
	userId: string,
	role: 'student' | 'teaching_assistant'
): Promise<{ role: string }> {
	return apiFetch(`/courses/${courseId}/members/${userId}/role`, {
		method: 'PATCH',
		body: JSON.stringify({ role })
	});
}
