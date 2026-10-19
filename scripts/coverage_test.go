package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCoverageGateUsesWeightedStatementsAndDomainThreshold(t *testing.T) {
	for _, tc := range []struct {
		name, profile string
		pass          bool
	}{
		{"passes-boundaries-excludes-cmd", "mode: atomic\nx/internal/domain/d.go:1.1,2.1 90 1\nx/internal/domain/d.go:3.1,4.1 10 0\nx/internal/api/a.go:1.1,2.1 10 0\nx/internal/api/zero.go:1.1,2.1 0 1\nx/cmd/server/m.go:1.1,2.1 5000 0\n", true},
		{"fails-domain", "mode: atomic\nx/internal/domain/d.go:1.1,2.1 80 1\nx/internal/domain/d.go:3.1,4.1 20 0\n", false},
		{"fails-weighted-not-package-average", "mode: atomic\nx/internal/domain/d.go:1.1,2.1 90 1\nx/internal/domain/d.go:3.1,4.1 10 0\nx/internal/api/a.go:1.1,2.1 100 0\n", false},
		{"empty", "mode: atomic\n", false},
		{"missing-domain", "mode: atomic\nx/internal/api/a.go:1.1,2.1 100 1\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "coverage.out")
			if err := os.WriteFile(path, []byte(tc.profile), 0600); err != nil {
				t.Fatal(err)
			}
			output, err := exec.Command("go", "run", "./coveragegate", path).CombinedOutput()
			if (err == nil) != tc.pass {
				t.Fatalf("pass=%v error=%v output=%s", tc.pass, err, output)
			}
			if tc.pass && !strings.Contains(string(output), "90/110 = 81.82%") {
				t.Fatalf("wrong aggregation: %s", output)
			}
		})
	}
}

func TestWorkflowParsesWithRequiredLocalGateCommands(t *testing.T) {
	data, err := os.ReadFile("../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	var workflow struct {
		Jobs map[string]struct {
			Steps []struct {
				Run string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		t.Fatal(err)
	}
	var runs []string
	for _, job := range workflow.Jobs {
		for _, step := range job.Steps {
			runs = append(runs, step.Run)
		}
	}
	all := strings.Join(runs, "\n")
	for _, required := range []string{"gofmt -l .", "go build ./...", "go vet ./...", "go test -p 2 ./... -race -count=1 -coverprofile=coverage.out", "go tool cover -func=coverage.out", "go run ./scripts/coveragegate coverage.out", "git diff --exit-code -- go.mod go.sum"} {
		if !strings.Contains(all, required) {
			t.Errorf("missing workflow gate %q", required)
		}
	}
}
