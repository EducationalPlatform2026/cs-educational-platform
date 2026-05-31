import { apiFetch } from './client';

export interface Exercise {
	id: string;
	course_id: string;
	created_by: string;
	title: string;
	description?: string;
	instructions: string;
	difficulty: 'easy' | 'medium' | 'hard';
	language: 'python' | 'go' | 'java' | 'c' | 'cpp' | 'javascript';
	template_code?: string;
	time_limit_ms: number;
	memory_limit_kb: number;
	is_published: boolean;
	created_at: string;
	updated_at: string;
}

export interface TestCase {
	id: string;
	exercise_id: string;
	input: string;
	expected_output: string;
	is_hidden: false;
	ordinal: number;
	created_at: string;
}

export interface HiddenTestCase {
	id: string;
	exercise_id: string;
	is_hidden: true;
	ordinal: number;
	created_at: string;
}

export type AnyTestCase = TestCase | HiddenTestCase;

export function listExercises(courseId: string): Promise<Exercise[]> {
	return apiFetch(`/courses/${courseId}/exercises`);
}

export function getExercise(id: string): Promise<Exercise> {
	return apiFetch(`/exercises/${id}`);
}

export function createExercise(
	courseId: string,
	data: {
		title: string;
		description?: string;
		instructions: string;
		difficulty: string;
		language: string;
		template_code?: string;
		time_limit_ms?: number;
		memory_limit_kb?: number;
		is_published?: boolean;
	}
): Promise<Exercise> {
	return apiFetch(`/courses/${courseId}/exercises`, {
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export function updateExercise(
	id: string,
	data: Partial<{
		title: string;
		description: string;
		instructions: string;
		difficulty: string;
		language: string;
		template_code: string;
		time_limit_ms: number;
		memory_limit_kb: number;
		is_published: boolean;
	}>
): Promise<Exercise> {
	return apiFetch(`/exercises/${id}`, { method: 'PUT', body: JSON.stringify(data) });
}

export function deleteExercise(id: string): Promise<void> {
	return apiFetch(`/exercises/${id}`, { method: 'DELETE' });
}

export function listTestCases(exerciseId: string): Promise<AnyTestCase[]> {
	return apiFetch(`/exercises/${exerciseId}/test-cases`);
}

export function createTestCase(
	exerciseId: string,
	data: { input: string; expected_output: string; is_hidden: boolean; ordinal: number }
): Promise<TestCase> {
	return apiFetch(`/exercises/${exerciseId}/test-cases`, {
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export function deleteTestCase(id: string): Promise<void> {
	return apiFetch(`/test-cases/${id}`, { method: 'DELETE' });
}
