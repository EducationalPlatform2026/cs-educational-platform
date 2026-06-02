package sandbox

import (
	"context"
	"fmt"
	"log"
	"time"

	"cs-educational-platform/backend/internal/db"
)

// StartWorker launches a background goroutine that polls the DB every 3 seconds
// for pending submissions, evaluates them, and writes results back.
// It stops cleanly when ctx is cancelled.
func StartWorker(ctx context.Context) {
	go func() {
		log.Println("sandbox: worker started")
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("sandbox: worker stopped")
				return
			case <-ticker.C:
				if err := processBatch(ctx); err != nil {
					log.Printf("sandbox: batch error: %v", err)
				}
			}
		}
	}()
}

// processBatch picks up to 5 oldest pending submissions and evaluates them.
func processBatch(ctx context.Context) error {
	rows, err := db.Pool.Query(ctx,
		`SELECT s.id, s.exercise_id, s.code, s.language,
		        e.time_limit_ms, e.memory_limit_kb
		 FROM   submissions s
		 JOIN   exercises   e ON e.id = s.exercise_id
		 WHERE  s.status = 'pending'
		 ORDER  BY s.submitted_at
		 LIMIT  5`,
	)
	if err != nil {
		return fmt.Errorf("query pending: %w", err)
	}
	defer rows.Close()

	type job struct {
		id, exerciseID, code, language string
		timeLimitMs, memoryLimitKb     int
	}
	var jobs []job
	for rows.Next() {
		var j job
		if err := rows.Scan(&j.id, &j.exerciseID, &j.code, &j.language,
			&j.timeLimitMs, &j.memoryLimitKb); err != nil {
			return fmt.Errorf("scan job: %w", err)
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, j := range jobs {
		evaluate(ctx, j.id, j.exerciseID, j.code, j.language, j.timeLimitMs)
	}
	return nil
}

// evaluate runs one submission end-to-end and writes all results to the DB.
func evaluate(ctx context.Context, submissionID, exerciseID, code, language string, timeLimitMs int) {
	log.Printf("sandbox: evaluating submission %s", submissionID)

	// Mark as running so no other worker picks it up.
	if _, err := db.Pool.Exec(ctx,
		`UPDATE submissions SET status = 'running' WHERE id = $1`, submissionID,
	); err != nil {
		log.Printf("sandbox: mark running %s: %v", submissionID, err)
		return
	}

	// Fetch test cases ordered by ordinal.
	tcRows, err := db.Pool.Query(ctx,
		`SELECT id, input, expected_output
		 FROM   test_cases
		 WHERE  exercise_id = $1
		 ORDER  BY ordinal, id`,
		exerciseID,
	)
	if err != nil {
		finishError(ctx, submissionID, fmt.Sprintf("fetch test cases: %v", err))
		return
	}

	var tests []TestInput
	for tcRows.Next() {
		var t TestInput
		if err := tcRows.Scan(&t.ID, &t.Stdin, &t.Expected); err != nil {
			tcRows.Close()
			finishError(ctx, submissionID, fmt.Sprintf("scan test case: %v", err))
			return
		}
		tests = append(tests, t)
	}
	tcRows.Close()
	if err := tcRows.Err(); err != nil {
		finishError(ctx, submissionID, err.Error())
		return
	}

	// No test cases → auto-accept with full score.
	if len(tests) == 0 {
		db.Pool.Exec(ctx, //nolint:errcheck
			`UPDATE submissions SET status = 'accepted'::submission_status, score = 100 WHERE id = $1`,
			submissionID)
		log.Printf("sandbox: submission %s → accepted (no test cases)", submissionID)
		return
	}

	// Run all test cases.
	outputs := RunAll(ctx, language, code, timeLimitMs, tests)

	// Tally results and determine overall status.
	passed := 0
	finalStatus := "accepted"
	var lastStderr string

	for _, o := range outputs {
		if o.Status != "accepted" && (finalStatus == "accepted" || finalStatus == "") {
			finalStatus = o.Status
		}
		if o.Status == "accepted" {
			passed++
		}
		if o.Stderr != "" {
			lastStderr = o.Stderr
		}

		// Persist per-test-case result.
		actualOut := &o.ActualOutput
		if o.ActualOutput == "" {
			actualOut = nil
		}
		runtimeMs := &o.RuntimeMs
		if o.RuntimeMs == 0 {
			runtimeMs = nil
		}
		if _, err := db.Pool.Exec(ctx,
			`INSERT INTO submission_results
			   (submission_id, test_case_id, status, actual_output, runtime_ms)
			 VALUES ($1, $2, $3::submission_status, $4, $5)`,
			submissionID, o.ID, o.Status, actualOut, runtimeMs,
		); err != nil {
			log.Printf("sandbox: insert result for tc %s: %v", o.ID, err)
		}
	}

	score := float64(passed) / float64(len(tests)) * 100

	var stderrPtr *string
	if lastStderr != "" {
		stderrPtr = &lastStderr
	}

	if _, err := db.Pool.Exec(ctx,
		`UPDATE submissions
		 SET status = $1::submission_status, score = $2, stderr = $3
		 WHERE id = $4`,
		finalStatus, score, stderrPtr, submissionID,
	); err != nil {
		log.Printf("sandbox: update submission %s: %v", submissionID, err)
		return
	}

	log.Printf("sandbox: submission %s → %s (%.0f%% — %d/%d passed)",
		submissionID, finalStatus, score, passed, len(tests))
}

// finishError marks the submission as runtime_error with an internal message.
func finishError(ctx context.Context, submissionID, reason string) {
	log.Printf("sandbox: internal error for %s: %s", submissionID, reason)
	db.Pool.Exec(ctx, //nolint:errcheck
		`UPDATE submissions
		 SET status = 'runtime_error'::submission_status, stderr = $1
		 WHERE id = $2`,
		reason, submissionID,
	)
}
