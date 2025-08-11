package profiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golite.dev/mvp/internal/lexer"
	"golite.dev/mvp/internal/optimizer"
	"golite.dev/mvp/internal/parser"
)

// Metrics represents the collected performance data for a single run.
type Metrics struct {
	SourceFile       string  `json:"source_file"`
	BuildTimeMs      float64 `json:"build_time_ms"`
	BinarySizeBytes  int64   `json:"binary_size_bytes"`
	RunTimeMs        float64 `json:"run_time_ms"`
	MemoryUsageBytes int64   `json:"memory_usage_bytes"`
}

// Executor defines an interface for running external commands, allowing for mocking in tests.
type Executor interface {
	CombinedOutput(cmd *exec.Cmd) ([]byte, error)
}

// Profiler orchestrates the build, run, and measurement process.
type Profiler struct {
	exec    Executor
	workDir string // A temporary directory for intermediate files.
}

// New creates a new Profiler.
func New(executor Executor, workDir string) *Profiler {
	return &Profiler{
		exec:    executor,
		workDir: workDir,
	}
}

// Run executes the full profiling pipeline for a given GoLite source file
// using a specific optimizer configuration.
func (p *Profiler) Run(sourceFile string, optConfig optimizer.Config) (*Metrics, error) {
	metrics := &Metrics{SourceFile: filepath.Base(sourceFile)}

	goliteCompilerPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("could not find compiler executable path: %w", err)
	}

	input, err := os.ReadFile(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}

	l := lexer.New(string(input))
	program := parser.New(l).ParseProgram()
	optimizer.Optimize(program, optConfig)

	tempGoLiteFile := filepath.Join(p.workDir, "temp.golite")
	err = os.WriteFile(tempGoLiteFile, []byte(program.String()), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write temp golite file: %w", err)
	}

	cFile := filepath.Join(p.workDir, "output.c")
	binaryFile := filepath.Join(p.workDir, "program")

	buildStartTime := time.Now()
	buildCmd := exec.Command(goliteCompilerPath, "build", "-o", cFile, tempGoLiteFile)
	if output, err := p.exec.CombinedOutput(buildCmd); err != nil {
		return nil, fmt.Errorf("failed to compile GoLite to C: %s\n%s", err, string(output))
	}

	compileCmd := exec.Command("clang", cFile, "-o", binaryFile)
	if output, err := p.exec.CombinedOutput(compileCmd); err != nil {
		return nil, fmt.Errorf("failed to compile C to native: %s\n%s", err, string(output))
	}
	metrics.BuildTimeMs = float64(time.Since(buildStartTime).Microseconds()) / 1000.0

	fileInfo, err := os.Stat(binaryFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get binary size: %w", err)
	}
	metrics.BinarySizeBytes = fileInfo.Size()

	timeCmd, err := findTimeCommand()
	if err != nil {
		return nil, err
	}

	// THIS IS THE FIX: We wrap the command in `sh -c` to redirect its stdout to /dev/null,
	// so it doesn't interfere with the stderr output from the `time` command.
	runCmdString := fmt.Sprintf("%s %s %s > /dev/null", timeCmd.Path, timeCmd.Flag, binaryFile)
	runCmd := exec.Command("sh", "-c", runCmdString)

	runOutput, err := p.exec.CombinedOutput(runCmd)
	if err != nil {
		// `time` exits non-zero if the child process does. This is often okay.
	}

	if err := parseTimeOutput(string(runOutput), metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

type timeCommand struct {
	Path string
	Flag string
}

func findTimeCommand() (*timeCommand, error) {
	path := "/usr/bin/time"
	if _, err := os.Stat(path); err != nil {
		path = "/bin/time"
		if _, err2 := os.Stat(path); err2 != nil {
			return nil, fmt.Errorf("profiling tool '/usr/bin/time' not found")
		}
	}

	cmd := exec.Command(path, "--version")
	output, _ := cmd.CombinedOutput()
	if strings.Contains(string(output), "GNU time") {
		return &timeCommand{Path: path, Flag: "-v"}, nil
	}
	return &timeCommand{Path: path, Flag: "-l"}, nil
}

var (
	userTimeRegexBSD = regexp.MustCompile(`(\d+\.\d+)\s+user`)
	sysTimeRegexBSD  = regexp.MustCompile(`(\d+\.\d+)\s+sys`)
	memoryRegexBSD   = regexp.MustCompile(`(\d+)\s+maximum resident set size`)

	userTimeRegexGNU = regexp.MustCompile(`User time \(seconds\):\s+(\d+\.\d+)`)
	sysTimeRegexGNU  = regexp.MustCompile(`System time \(seconds\):\s+(\d+\.\d+)`)
	memoryRegexGNU   = regexp.MustCompile(`Maximum resident set size \(kbytes\):\s+(\d+)`)
)

func parseTimeOutput(output string, metrics *Metrics) error {
	var userTime, sysTime, memUsage float64
	var err error

	if memMatch := memoryRegexBSD.FindStringSubmatch(output); len(memMatch) > 1 {
		userMatch := userTimeRegexBSD.FindStringSubmatch(output)
		sysMatch := sysTimeRegexBSD.FindStringSubmatch(output)
		if len(userMatch) < 2 || len(sysMatch) < 2 {
			return fmt.Errorf("could not parse BSD time output: %s", output)
		}
		userTime, err = strconv.ParseFloat(userMatch[1], 64)
		if err != nil {
			return err
		}
		sysTime, err = strconv.ParseFloat(sysMatch[1], 64)
		if err != nil {
			return err
		}
		memUsage, err = strconv.ParseFloat(memMatch[1], 64)
		if err != nil {
			return err
		}
	} else if memMatch := memoryRegexGNU.FindStringSubmatch(output); len(memMatch) > 1 {
		userMatch := userTimeRegexGNU.FindStringSubmatch(output)
		sysMatch := sysTimeRegexGNU.FindStringSubmatch(output)
		if len(userMatch) < 2 || len(sysMatch) < 2 {
			return fmt.Errorf("could not parse GNU time output: %s", output)
		}
		userTime, err = strconv.ParseFloat(userMatch[1], 64)
		if err != nil {
			return err
		}
		sysTime, err = strconv.ParseFloat(sysMatch[1], 64)
		if err != nil {
			return err
		}
		memUsage, err = strconv.ParseFloat(memMatch[1], 64)
		if err != nil {
			return err
		}
		memUsage *= 1024
	} else {
		return fmt.Errorf("could not parse output from `time` command; profiling failed. Output:\n%s", output)
	}

	metrics.RunTimeMs = (userTime + sysTime) * 1000.0
	metrics.MemoryUsageBytes = int64(memUsage)
	return nil
}
