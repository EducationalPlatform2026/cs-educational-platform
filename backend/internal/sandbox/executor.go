package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TestInput is one test case to run against.
type TestInput struct {
	ID       string
	Stdin    string
	Expected string
}

// TestOutput is the result for one test case.
type TestOutput struct {
	ID           string
	Status       string // matches submission_status enum
	ActualOutput string
	Stderr       string
	RuntimeMs    int
}

// RunAll compiles the code once (for compiled languages) and then runs it
// against every test case. If compilation fails every output has status
// "compile_error" and the compile stderr is attached.
// If tests is empty the caller should handle the no-test-case case itself.
func RunAll(ctx context.Context, language, code string, timeLimitMs int, tests []TestInput) []TestOutput {
	tmpDir, err := os.MkdirTemp("", "sandbox-*")
	if err != nil {
		return failAll(tests, "compile_error", fmt.Sprintf("create temp dir: %v", err))
	}
	defer os.RemoveAll(tmpDir)

	// --- prepare: write source + compile if needed ---
	runCmd, compileErr := prepare(ctx, tmpDir, language, code)
	if compileErr != "" {
		return failAll(tests, "compile_error", compileErr)
	}

	// --- run each test case ---
	if timeLimitMs <= 0 {
		timeLimitMs = 5000
	}
	timeout := time.Duration(timeLimitMs) * time.Millisecond

	out := make([]TestOutput, 0, len(tests))
	for _, tc := range tests {
		r := executeOne(ctx, runCmd, tmpDir, tc.Stdin, timeout)
		out = append(out, TestOutput{
			ID:           tc.ID,
			Status:       classify(r, tc.Expected),
			ActualOutput: r.stdout,
			Stderr:       r.stderr,
			RuntimeMs:    r.runtimeMs,
		})
	}
	return out
}

// ── internal helpers ──────────────────────────────────────────────────────────

// prepare writes the source file and compiles if the language requires it.
// Returns the command line to execute one test run, or a non-empty compile error string.
func prepare(ctx context.Context, dir, language, code string) ([]string, string) {
	switch language {
	case "python":
		if err := writeFile(dir, "script.py", code); err != nil {
			return nil, err.Error()
		}
		return pythonCmd("script.py"), ""

	case "javascript":
		if err := writeFile(dir, "script.js", code); err != nil {
			return nil, err.Error()
		}
		return []string{"node", filepath.Join(dir, "script.js")}, ""

	case "go":
		if err := writeFile(dir, "main.go", code); err != nil {
			return nil, err.Error()
		}
		// go run operates from the directory; pass the absolute path so the
		// working directory during execution does not matter.
		return []string{"go", "run", filepath.Join(dir, "main.go")}, ""

	case "java":
		if err := writeFile(dir, "Main.java", code); err != nil {
			return nil, err.Error()
		}
		stderr, ok := compile(ctx, dir, "javac", "Main.java")
		if !ok {
			return nil, stderr
		}
		return []string{"java", "-cp", dir, "Main"}, ""

	case "c":
		if err := writeFile(dir, "main.c", code); err != nil {
			return nil, err.Error()
		}
		bin := filepath.Join(dir, binaryName("main"))
		stderr, ok := compile(ctx, dir, "gcc", "main.c", "-o", bin)
		if !ok {
			return nil, stderr
		}
		return []string{bin}, ""

	case "cpp":
		if err := writeFile(dir, "main.cpp", code); err != nil {
			return nil, err.Error()
		}
		bin := filepath.Join(dir, binaryName("main"))
		stderr, ok := compile(ctx, dir, "g++", "main.cpp", "-o", bin)
		if !ok {
			return nil, stderr
		}
		return []string{bin}, ""

	default:
		return nil, "unsupported language: " + language
	}
}

type runResult struct {
	stdout    string
	stderr    string
	runtimeMs int
	timedOut  bool
	exitCode  int
}

func executeOne(ctx context.Context, cmdArgs []string, dir, stdin string, timeout time.Duration) runResult {
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(runCtx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := int(time.Since(start).Milliseconds())

	exitCode := 0
	if err != nil {
		if ex, ok := err.(*exec.ExitError); ok {
			exitCode = ex.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return runResult{
		stdout:    stdout.String(),
		stderr:    stderr.String(),
		runtimeMs: elapsed,
		timedOut:  runCtx.Err() == context.DeadlineExceeded,
		exitCode:  exitCode,
	}
}

// classify maps a run result + expected output to a submission_status string.
func classify(r runResult, expected string) string {
	if r.timedOut {
		return "time_limit"
	}
	if r.exitCode != 0 {
		return "runtime_error"
	}
	// Trim trailing whitespace / newlines to be lenient with output formatting.
	got := strings.TrimRight(r.stdout, " \t\r\n")
	want := strings.TrimRight(expected, " \t\r\n")
	if got == want {
		return "accepted"
	}
	return "wrong_answer"
}

// compile runs a compiler in dir and returns (stderr, success).
func compile(ctx context.Context, dir string, name string, args ...string) (string, bool) {
	compCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var stderr bytes.Buffer
	cmd := exec.CommandContext(compCtx, name, args...)
	cmd.Dir = dir
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stderr.String(), false
	}
	return "", true
}

func writeFile(dir, name, content string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
}

func pythonCmd(file string) []string {
	if runtime.GOOS == "windows" {
		return []string{"python", file}
	}
	return []string{"python3", file}
}

func binaryName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

func failAll(tests []TestInput, status, stderr string) []TestOutput {
	out := make([]TestOutput, len(tests))
	for i, tc := range tests {
		out[i] = TestOutput{ID: tc.ID, Status: status, Stderr: stderr}
	}
	return out
}
