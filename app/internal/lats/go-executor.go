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
	codeBlockRegex  *regexp.Regexp
	packageRegex    *regexp.Regexp
)

func init() {
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
	if strings.Contains(code, "```") {
		code = codeBlockRegex.FindStringSubmatch(code)[1]
	}
	if strings.Contains(code, "package") {
		code = packageRegex.ReplaceAllString(code, "package lats")
	} else {
		code = "package lats\n\n" + code
	}

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
	var stderr bytes.Buffer
	cmd := exec.Command("goimports", "-w", path)
	cmd.Dir = filepath.Dir(path)
	cmd.Stderr = &stderr
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("could not format the code:\n%s", stderr.String())
	}

	cmd = exec.Command("go", "get", "-d", "./...")
	cmd.Dir = filepath.Dir(path)
	cmd.Stderr = &stderr
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("could not get the dependencies:\n%s", stderr.String())
	}

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = filepath.Dir(path)
	cmd.Stderr = &stderr
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("could not run go mod tidy:\n%s", stderr.String())
	}

	cmd = exec.Command("go", "fmt", path)
	cmd.Dir = filepath.Dir(path)
	cmd.Stderr = &stderr
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("could not run go fmt:\n%s", stderr.String())
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
		if strings.HasPrefix(line, fmt.Sprintf(".\\%s", filename)) {
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
	failedAssert := ""
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "        lats_test.go") {
			if failedAssert != "" {
				failedAsserts = append(failedAsserts, failedAssert)
			}
			failedAssert = strings.Trim(line, "") + "\n"
		} else if strings.HasPrefix(line, "        ") {
			failedAssert += strings.Trim(line, "") + "\n"
		}
	}
	if failedAssert != "" {
		failedAsserts = append(failedAsserts, failedAssert)
	}
	return failedAsserts
}

func cleanUp(tempDir string) {
	err := os.RemoveAll(tempDir)
	if err != nil {
		fmt.Println(err)
	}
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

	return float32(totalPoints) / float32(maxPoints)
}
