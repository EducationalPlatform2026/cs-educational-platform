import { apiFetch } from './client';

export interface StaffMember {
	user_id: string;
	first_name: string;
	last_name: string;
	email: string;
}

export interface StudentStat {
	user_id: string;
	first_name: string;
	last_name: string;
	email: string;
	enrolled_at: string;
	exercises_attempted: number;
	exercises_solved: number;
	total_submissions: number;
	accepted_submissions: number;
	acceptance_rate: number;
	last_submission_at?: string;
}

export interface ExerciseStat {
	id: string;
	title: string;
	difficulty: 'easy' | 'medium' | 'hard';
	students_attempted: number;
	students_solved: number;
	solve_rate: number;
	total_submissions: number;
}

export interface CourseSummary {
	total_students: number;
	total_teaching_assistants: number;
	total_exercises: number;
	total_submissions: number;
	accepted_submissions: number;
	course_acceptance_rate: number;
	students_with_submission: number;
}

export interface CourseStats {
	professor: StaffMember;
	teaching_assistants: StaffMember[];
	summary: CourseSummary;
	students: StudentStat[];
	exercises: ExerciseStat[];
}

export function getCourseStats(courseId: string): Promise<CourseStats> {
	return apiFetch(`/courses/${courseId}/stats`);
}
