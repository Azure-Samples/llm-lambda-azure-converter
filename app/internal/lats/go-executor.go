package lats

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
)

var (
	failedTestRegex *regexp.Regexp
)

func init() {
	failedTestPattern := "^(?:.+):(\\d+): (.+)$"
	failedTestRegex = regexp.MustCompile(failedTestPattern)
}

type goExecutor struct{
}

func NewGoExecutor() models.Executor {
	return &goExecutor{}
}

func createTempProject() (string, error) {
	rand := uuid.New().String()

	tempPath := os.TempDir()
	tempDir := filepath.Join(tempPath, fmt.Sprintf("go-lats-%s", rand))
	os.Mkdir(tempDir, 0755)

	cmd := exec.Command(fmt.Sprintf("go mod init go-lats-%s", rand))
	cmd.Dir = tempDir
	_, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return tempDir, nil
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
	cmd := exec.Command(fmt.Sprintf("go fmt %s", path))
	cmd.Dir = filepath.Base(path)
	_, err := cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.Command(fmt.Sprintf("goimports -w %s", path))
	cmd.Dir = filepath.Base(path)
	_, err = cmd.Output()
	if err != nil {
		return err
	}

	return nil
}

func downloadImports(path string) error {
	cmd := exec.Command("go get -d ./... && go mod tidy")
	cmd.Dir = filepath.Base(path)
	_, err := cmd.Output()
	if err != nil {
		return err
	}

	return nil
}

func buildProject(path string) ([]string, error) {
	cmd := exec.Command("go build ./...")
	cmd.Dir = filepath.Base(path)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	compileErrors := grabCompileErrs(string(output))

	return compileErrors, nil
}

func grabCompileErrs(output string) []string {
	objs := make([]string, 0)
	compileErr := ""
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, ".\\lats.go") {
			if compileErr != "" {
				objs = append(objs, compileErr)
			}
			compileErr = strings.Trim(line, "") + "\n"
		}
		if strings.HasPrefix(line, "        ") {
			compileErr += strings.Trim(line, "") + "\n"
		}
	}

	if compileErr != "" {
		objs = append(objs, compileErr)
	}
	return objs
}

func runTests(path string) ([]string, error) {
	cmd := exec.Command("go test ./...")
	cmd.Dir = filepath.Base(path)
	output, err := cmd.Output()
	if err != nil {
		compileErrors := grabCompileErrs(err.Error())
		return compileErrors, nil
	}
	return grabTestErrors(string(output)), nil
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

func (e goExecutor) Execute(code string, tests []string) (*models.ExecutionResult, error) {
	tempDir, err := createTempProject()
	if err != nil {
		return nil, err
	}
	defer cleanUp(tempDir)

	codePath := filepath.Join(tempDir, "lats.go")
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

	compileErrors, err := buildProject(tempDir)
	if err != nil {
		return nil, err
	}

	if len(compileErrors) > 0 {
		return &models.ExecutionResult{
			IsPassing: false,
			Feedback:  strings.Join(compileErrors, "\n"),
			State:     make([]bool, len(tests)),
		}, nil
	}

	isPassing := true
	passedFeedback := "Tested passed:\n"
	failedFeedback := "Tested failed:\n"
	state := make([]bool, len(tests))

	for _, test := range tests {
		testPath := filepath.Join(tempDir, "lats_test.go")
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

		testErrors, err := runTests(tempDir)
		if err != nil {
			return nil, err
		}

		if len(testErrors) > 0 {
			isPassing = false
			failedFeedback += fmt.Sprintf("%s\n%s\n", test, strings.Join(testErrors, "\n"))
			state = append(state, false)
		} else {
			passedFeedback += fmt.Sprintf("%s\n", test)
			state = append(state, true)
		}

	}

	return &models.ExecutionResult{
		IsPassing: isPassing,
		Feedback:  fmt.Sprintf("%s\n%s", passedFeedback, failedFeedback),
		State:     state,
	}, nil
}

func (e goExecutor) Evaluate(code string, tests []string) bool {
	result, err := e.Execute(code, tests)
	if err != nil {
		return false
	}

	return result.IsPassing	
}