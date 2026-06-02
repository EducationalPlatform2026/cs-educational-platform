import { apiFetch } from './client';

export interface Submission {
	id: string;
	exercise_id: string;
	user_id: string;
	code: string;
	language: string;
	status: SubmissionStatus;
	score: number;
	stderr?: string;
	submitted_at: string;
}

export interface SubmissionResult {
	id: string;
	submission_id: string;
	test_case_id: string;
	status: SubmissionStatus;
	actual_output?: string;
	runtime_ms?: number;
	memory_kb?: number;
}

export interface SubmissionDetail extends Submission {
	results: SubmissionResult[];
}

export type SubmissionStatus =
	| 'pending'
	| 'running'
	| 'accepted'
	| 'wrong_answer'
	| 'runtime_error'
	| 'time_limit'
	| 'memory_limit'
	| 'compile_error';

export function submitCode(exerciseId: string, code: string, language: string): Promise<Submission> {
	return apiFetch(`/exercises/${exerciseId}/submit`, {
		method: 'POST',
		body: JSON.stringify({ code, language })
	});
}

export function listSubmissions(exerciseId: string): Promise<Submission[]> {
	return apiFetch(`/exercises/${exerciseId}/submissions`);
}

export function getSubmission(id: string): Promise<SubmissionDetail> {
	return apiFetch(`/submissions/${id}`);
}
