package lats

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
)

const (
	compileWeight = 2
)

var (
	failedTestRegex *regexp.Regexp
	codeBlockRegex  *regexp.Regexp
	packageRegex	*regexp.Regexp
)

func init() {
	failedTestPattern := "^(?:.+):(\\d+): (.+)$"
	failedTestRegex = regexp.MustCompile(failedTestPattern)

	codeBlockPattern := "(?s)```go(.*?)```"
	codeBlockRegex = regexp.MustCompile(codeBlockPattern)

	packagePattern := "(package \\w+)"
	packageRegex = regexp.MustCompile(packagePattern)
}

type goExecutor struct {
}

func NewGoExecutor() models.Executor {
	return &goExecutor{}
}

func createTempProject(targetDir string) error {
	cmd := exec.Command("go", "mod", "init", "go-lats")
	cmd.Dir = targetDir
	_, err := cmd.Output()

	if err != nil {
		return err
	}

	return nil
}

func writeToFile(path, code string) error {
	prelude := "package lats\n\n"
	code = prelude + code

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(code))
	if err != nil {
		return err
	}

	return nil
}

func formatFile(path string) error {
	cmd := exec.Command("go", "fmt", path)
	cmd.Dir = filepath.Dir(path)
	_, err := cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.Command("goimports", "-w", path)
	cmd.Dir = filepath.Dir(path)
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	return nil
}

func downloadImports(path string) error {
	cmd := exec.Command("go", "get", "-d", "./...")
	cmd.Dir = path
	_, err := cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = path
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	return nil
}

func buildProject(path string, filename string) ([]string, error) {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	compileErrors := grabCompileErrs(string(output), filename)

	return compileErrors, nil
}

func grabCompileErrs(output string, filename string) []string {
	compileErrors := make([]string, 0)
	compileErr := ""
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, fmt.Sprintf( ".\\%s", filename)) {
			if compileErr != "" {
				compileErrors = append(compileErrors, compileErr)
			}
			compileErr = strings.Trim(line, "") + "\n"
		}
		if strings.HasPrefix(line, "        ") {
			compileErr += strings.Trim(line, "") + "\n"
		}
	}

	if compileErr != "" {
		compileErrors = append(compileErrors, compileErr)
	}
	return compileErrors
}

func runTests(path string, filename string) ([]string, error) {
	testErrors := make([]string, 0)
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = path
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		compileErrors := grabCompileErrs(stderr.String(), filename)
		testErrors = append(testErrors, compileErrors...)
	}
	testErrors = append(testErrors, grabTestErrors(string(output))...)
	return testErrors, nil
}

func grabTestErrors(output string) []string {
	failedAsserts := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "        lats_test.go") {
			matches := failedTestRegex.FindStringSubmatch(strings.Trim(line, ""))
			var lineNo, panicReason string
			if len(matches) == 3 {
				lineNo = matches[1]
				panicReason = matches[2]
			}
			failedAsserts = append(failedAsserts, fmt.Sprintf("[Line] %s, [Reason] %s", lineNo, panicReason))
		}
	}
	return failedAsserts
}

func cleanUp(tempDir string) {
	os.RemoveAll(tempDir)
}

func (e *goExecutor) Execute(code string, tests []string) (*models.ExecutionResult, error) {
	rand := uuid.New().String()
	tempPath := os.TempDir()
	tempDir := filepath.Join(tempPath, fmt.Sprintf("go-lats-%s", rand))
	os.Mkdir(tempDir, 0755)

	err := createTempProject(tempDir)
	if err != nil {
		return nil, err
	}
	defer cleanUp(tempDir)

	code = codeBlockRegex.FindStringSubmatch(code)[1]
	code = packageRegex.ReplaceAllString(code, "")

	filename := "lats.go"
	codePath := filepath.Join(tempDir, filename)
	err = writeToFile(codePath, code)
	if err != nil {
		return nil, err
	}

	err = formatFile(codePath)
	if err != nil {
		return nil, err
	}

	err = downloadImports(tempDir)
	if err != nil {
		return nil, err
	}

	compileErrors, err := buildProject(tempDir, filename)
	if err != nil {
		return nil, err
	}

	if len(compileErrors) > 0 {
		return &models.ExecutionResult{
			IsPassing: false,
			Feedback:  strings.Join(compileErrors, "\n"),
			Score:     calculateScore(false, false, len(tests), 0),
		}, nil
	}

	isPassing := true
	passingTests := 0
	passedFeedback := "Tested passed:\n"
	failedFeedback := "Tested failed:\n"

	for _, test := range tests {
		testFilename := "lats_test.go"
		testPath := filepath.Join(tempDir, testFilename)
		err = writeToFile(testPath, test)
		if err != nil {
			return nil, err
		}

		err = formatFile(testPath)
		if err != nil {
			return nil, err
		}

		err = downloadImports(tempDir)
		if err != nil {
			return nil, err
		}

		testErrors, err := runTests(tempDir, testFilename)
		if err != nil {
			return nil, err
		}

		if len(testErrors) > 0 {
			isPassing = false
			failedFeedback += fmt.Sprintf("%s\n%s\n", test, strings.Join(testErrors, "\n"))
		} else {
			passedFeedback += fmt.Sprintf("%s\n", test)
			passingTests++
		}
	}

	return &models.ExecutionResult{
		IsPassing: isPassing,
		Feedback:  fmt.Sprintf("%s\n%s", passedFeedback, failedFeedback),
		Score:     calculateScore(isPassing, len(compileErrors) == 0, len(tests), passingTests),
	}, nil
}

func calculateScore(isPassing bool, compiles bool, totalTests int, passingTests int) float32 {
	if isPassing {
		return 1
	}

	maxPoints := compileWeight + totalTests
	compilePoints := 0
	if compiles {
		compilePoints = compileWeight
	}
	totalPoints := compilePoints + passingTests

	return float32(maxPoints) / float32(totalPoints)
}
