package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

// runMainEnv marks the re-executed test binary that should run main().
const runMainEnv = "PRODUCTPLAN_TEST_RUN_MAIN"

// TestMain lets a test re-execute this binary as the real program, so the
// process exit code and stderr are observed exactly as a user sees them.
func TestMain(m *testing.M) {
	if os.Getenv(runMainEnv) == "1" {
		os.Args = append([]string{"productplan"}, strings.Fields(os.Getenv(runMainEnv+"_ARGS"))...)
		main()
		return
	}
	os.Exit(m.Run())
}

// runProgram runs main() in a child process with the given env overrides.
func runProgram(t *testing.T, args string, env ...string) (exitCode int, stdout, stderr string) {
	t.Helper()
	cmd := exec.Command(os.Args[0]) // #nosec G204 G702 -- re-executes this test binary
	cmd.Env = append(filteredEnv("PRODUCTPLAN_API_TOKEN"), append(env, runMainEnv+"=1", runMainEnv+"_ARGS="+args)...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return 0, out.String(), errOut.String()
	case errors.As(err, &exitErr):
		return exitErr.ExitCode(), out.String(), errOut.String()
	}
	t.Fatalf("running program: %v", err)
	return 0, "", ""
}

// filteredEnv is the current environment without the named variables.
func filteredEnv(drop ...string) []string {
	var env []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if !slices.Contains(drop, name) {
			env = append(env, kv)
		}
	}
	return env
}

// Article VI: a missing token is a startup failure, not a failure on the
// first tool call. The process exits 1 and names the variable.
func TestMissingTokenExitsOneAndNamesVariable(t *testing.T) {
	for _, mode := range []string{"", "serve", "roadmaps"} {
		t.Run("mode="+mode, func(t *testing.T) {
			code, _, stderr := runProgram(t, mode)
			if code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
			if !strings.Contains(stderr, "PRODUCTPLAN_API_TOKEN") {
				t.Errorf("stderr does not name PRODUCTPLAN_API_TOKEN: %q", stderr)
			}
		})
	}
}

// An empty token is treated the same as an unset one.
func TestEmptyTokenExitsOne(t *testing.T) {
	code, _, stderr := runProgram(t, "", "PRODUCTPLAN_API_TOKEN=")
	if code != 1 || !strings.Contains(stderr, "PRODUCTPLAN_API_TOKEN") {
		t.Errorf("exit %d, stderr %q; want 1 naming the variable", code, stderr)
	}
}

// Help needs no token.
func TestHelpNeedsNoToken(t *testing.T) {
	code, stdout, _ := runProgram(t, "--help")
	if code != 0 || !strings.Contains(stdout, "Usage:") {
		t.Errorf("exit %d, stdout %q; want 0 with usage", code, stdout)
	}
}
