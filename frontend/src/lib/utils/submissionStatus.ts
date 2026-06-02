import type { SubmissionStatus } from '$lib/api/submissions';

export interface StatusMeta {
	label: string;
	cls: string;
}

export const statusMeta: Record<SubmissionStatus, StatusMeta> = {
	pending:       { label: 'Pending',       cls: 'status-pending'  },
	running:       { label: 'Running…',      cls: 'status-running'  },
	accepted:      { label: 'Accepted',      cls: 'status-accepted' },
	wrong_answer:  { label: 'Wrong Answer',  cls: 'status-wrong'    },
	runtime_error: { label: 'Runtime Error', cls: 'status-error'    },
	time_limit:    { label: 'Time Limit',    cls: 'status-error'    },
	memory_limit:  { label: 'Memory Limit',  cls: 'status-error'    },
	compile_error: { label: 'Compile Error', cls: 'status-error'    }
};

export function getStatusMeta(status: string): StatusMeta {
	return statusMeta[status as SubmissionStatus] ?? { label: status, cls: 'status-pending' };
}

/**
 * Shared CSS for status badges. Import these global styles once in +layout.svelte
 * or include them locally in each component that uses status badges.
 */
export const statusBadgeCSS = `
.status-accepted { background: #dcfce7; color: #166534; }
.status-pending  { background: #f3f4f6; color: #6b7280; }
.status-running  { background: #fef9c3; color: #a16207; }
.status-wrong    { background: #fef2f2; color: #b91c1c; }
.status-error    { background: #fef2f2; color: #b91c1c; }
`;
