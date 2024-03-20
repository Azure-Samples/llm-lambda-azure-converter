package lats

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
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
	codeBlockRegex *regexp.Regexp
)

func init() {
	codeBlockPattern := "(?s)```go(.*?)```"
	codeBlockRegex = regexp.MustCompile(codeBlockPattern)
}

type goExecutor struct {
}

func NewGoExecutor() models.Executor {
	return &goExecutor{}
}

func createTempProject(targetDir string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("go", "mod", "init", "github.com/devsquad/lats-temp-project")
	cmd.Dir = targetDir
	cmd.Stderr = &stderr
	_, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("could not create temp project:\n%s", stderr.String())
	}

	return nil
}

func writeToFile(path, code string) error {
	if strings.Contains(code, "```") {
		code = codeBlockRegex.FindStringSubmatch(code)[1]
	}
	if !strings.Contains(code, "package") {
		code = "package main\n\n" + code
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

func formatFile(projectPath string, path string, module string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("goimports", "-w", path)
	cmd.Dir = projectPath
	cmd.Stderr = &stderr
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("could not run 'goimports -w %s':\n%s", path, stderr.String())
	}

	cmd = exec.Command("go", "get", "-d", filepath.Join(module, filepath.Dir(path)))
	cmd.Dir = projectPath
	cmd.Stderr = &stderr
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("could not run 'go get -d':\n%s", stderr.String())
	}

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	cmd.Stderr = &stderr
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("could not run 'go mod tidy':\n%s", stderr.String())
	}

	cmd = exec.Command("go", "fmt", path)
	cmd.Dir = projectPath
	cmd.Stderr = &stderr
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("could not run go fmt:\n%s", stderr.String())
	}

	return nil
}

func buildProject(projectPath string, targetPackage string) ([]string, error) {
	var stderr bytes.Buffer
	cmd := exec.Command("go", "build", targetPackage)
	cmd.Dir = projectPath
	cmd.Stderr = &stderr
	_, err := cmd.Output()
	compileErrors := make([]string, 0)
	if err != nil {
		compileErrors = grabCompileErrs(stderr.String(), targetPackage)
	}

	return compileErrors, nil
}

func grabCompileErrs(output string, targetPackage string) []string {
	targetPackage = filepath.ToSlash(targetPackage)
	compileErrors := make([]string, 0)
	compileErr := ""
	found := false
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		} else if strings.HasPrefix(line, fmt.Sprintf("# %s", targetPackage)) {
			found = true
			if compileErr != "" {
				compileErrors = append(compileErrors, compileErr)
			}
			compileErr = strings.Trim(line, "") + "\n"
		} else if strings.HasPrefix(line, "#") {
			found = false
			compileErrors = append(compileErrors, compileErr)
			compileErr = ""
		} else if found {
			compileErr += strings.Trim(line, "") + "\n"
		}
	}

	if compileErr != "" {
		compileErrors = append(compileErrors, compileErr)
	}
	return compileErrors
}

func getModule(projectPath string) (string, error) {
	file, err := os.Open(path.Join(projectPath, "go.mod"))
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	module := strings.Replace(scanner.Text(), "module ", "", 1)

	return module, nil
}

func runTests(projectPath string, filePath string, module string) ([]string, error) {
	testErrors := make([]string, 0)
	targetPackage := filepath.Join(module, filepath.Dir(filePath))
	cmd := exec.Command("go", "test", targetPackage)
	cmd.Dir = projectPath
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		compileErrors := grabCompileErrs(stderr.String(), targetPackage)
		testErrors = append(testErrors, compileErrors...)
	}
	testErrors = append(testErrors, grabTestErrors(string(output), filepath.Base(filePath))...)
	return testErrors, nil
}

func grabTestErrors(output string, testFilename string) []string {
	failedAsserts := make([]string, 0)
	failedAssert := ""
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, fmt.Sprintf("        %s", testFilename)) {
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

func (e *goExecutor) Execute(code string, tests []string, converterOptions ...models.ExecutorOption) (*models.ExecutionResult, error) {

	options := &models.ExecutorOptions{
		ProjectPath:   filepath.Join(os.TempDir(), fmt.Sprintf("go-lats-%s", uuid.New().String())),
		TargetPath:    ".",
		Filename:      "lats",
		CreateProject: false,
	}
	for _, converterOption := range converterOptions {
		converterOption(options)
	}

	targetPath := filepath.Join(options.ProjectPath, options.TargetPath)
	err := os.RemoveAll(targetPath)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(targetPath, 0755)
	if err != nil {
		return nil, err
	}

	if options.CreateProject {
		os.Mkdir(options.ProjectPath, 0755)
		err := createTempProject(options.ProjectPath)
		if err != nil {
			return nil, err
		}
		defer cleanUp(options.ProjectPath)
	}

	module, err := getModule(options.ProjectPath)
	if err != nil {
		return nil, err
	}

	filePath := path.Join(options.TargetPath, fmt.Sprintf("%s.go", options.Filename))
	codePath := filepath.Join(options.ProjectPath, filePath)
	err = writeToFile(codePath, code)
	if err != nil {
		return nil, err
	}

	err = formatFile(options.ProjectPath, filePath, module)
	if err != nil {
		return nil, err
	}

	compileErrors, err := buildProject(options.ProjectPath, filepath.Join(module, options.TargetPath))
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
	passedFeedback := "Tests passed:\n"
	failedFeedback := "Tests failed:\n"

	for _, test := range tests {
		testFilename := fmt.Sprintf("%s_test.go", options.Filename)
		relativePath := path.Join(options.TargetPath, testFilename)
		testPath := path.Join(options.ProjectPath, relativePath)
		err = writeToFile(testPath, test)
		if err != nil {
			return nil, err
		}

		err = formatFile(options.ProjectPath, relativePath, module)
		if err != nil {
			return nil, err
		}

		testErrors, err := runTests(options.ProjectPath, relativePath, module)
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
