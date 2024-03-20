package lats

import (
	"reflect"
	"testing"

	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
)

func Test_goExecutor_Execute(t *testing.T) {
	type args struct {
		code    string
		tests   []string
		options []models.ExecutorOption
	}
	tests := []struct {
		name    string
		e       *goExecutor
		args    args
		want    *models.ExecutionResult
		wantErr bool
	}{
		{
			name: "Successful execution",
			e:    &goExecutor{},
			args: args{
				code: "```go" + `
package main

func salute(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
				` + "```",
				tests: []string{
					"```go" + `
package main
func TestSalute(t *testing.T) {
	assert.Equal(t, "Hello, World!", salute("World"))
}
` + "```",
					"```go" + `
package main
func TestSalute(t *testing.T) {
	assert.Equal(t, "Hello, Ana!", salute("Ana"))
}
` + "```",
					"```go" + `
package main
func TestSalute(t *testing.T) {
	assert.NotEqual(t, "Hello, Ana!", salute("Ada"))
}
` + "```",
				},
				options: []models.ExecutorOption{
					models.WithExCreateProject(true),
					models.WithExFilename("salute"),
				},
			},
			want: &models.ExecutionResult{
				IsPassing: true,
				Feedback:  "Tests passed:\n```go\npackage main\nfunc TestSalute(t *testing.T) {\n\tassert.Equal(t, \"Hello, World!\", salute(\"World\"))\n}\n```\n```go\npackage main\nfunc TestSalute(t *testing.T) {\n\tassert.Equal(t, \"Hello, Ana!\", salute(\"Ana\"))\n}\n```\n```go\npackage main\nfunc TestSalute(t *testing.T) {\n\tassert.NotEqual(t, \"Hello, Ana!\", salute(\"Ada\"))\n}\n```\n\nTests failed:\n",
				Score:     1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.Execute(tt.args.code, tt.args.tests, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("goExecutor.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("goExecutor.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_grabCompileErrs(t *testing.T) {
	type args struct {
		output        string
		targetPackage string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Basic error",
			args: args{
				output: `
# go-lats-35116-6739b2903daabf6d
.\lats.go:10:7: undefined: math
.\lats.go:11:18: too many return values
        have (bool, bool)
        want (bool)
.\lats.go:15:16: too many return values
        have (bool, bool)
        want (bool)`,
				targetPackage: "go-lats-35116-6739b2903daabf6d",
			},
			want: []string{
				`# go-lats-35116-6739b2903daabf6d
.\lats.go:10:7: undefined: math
.\lats.go:11:18: too many return values
        have (bool, bool)
        want (bool)
.\lats.go:15:16: too many return values
        have (bool, bool)
        want (bool)
`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := grabCompileErrs(tt.args.output, tt.args.targetPackage); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("grabCompileErrs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_grabTestErrors(t *testing.T) {
	type args struct {
		output   string
		filename string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Basic error",
			args: args{
				output: `
--- FAIL: TestHasCloseElements (0.00s)
    --- FAIL: TestHasCloseElements/all_elements_equal (0.00s)
        lats_test.go:53: HasCloseElements() = false, want true
    --- FAIL: TestHasCloseElements/negative_threshold (0.00s)
        lats_test.go:53: HasCloseElements() = false, want true
FAIL
FAIL    go-lats-35116-6739b2903daabf6d  2.672s
FAIL`,
				filename: "lats_test.go",
			},
			want: []string{
				"        lats_test.go:53: HasCloseElements() = false, want true\n",
				"        lats_test.go:53: HasCloseElements() = false, want true\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := grabTestErrors(tt.args.output, tt.args.filename); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("grabTestErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateScore(t *testing.T) {
	type args struct {
		isPassing    bool
		compiles     bool
		totalTests   int
		passingTests int
	}
	tests := []struct {
		name string
		args args
		want float32
	}{
		{
			name: "Passing",
			args: args{
				isPassing:    true,
				compiles:     true,
				totalTests:   8,
				passingTests: 8,
			},
			want: 1,
		},
		{
			name: "Not compiling",
			args: args{
				isPassing:    false,
				compiles:     false,
				totalTests:   8,
				passingTests: 0,
			},
			want: 0,
		},
		{
			name: "Compiles but no test pass",
			args: args{
				isPassing:    false,
				compiles:     true,
				totalTests:   8,
				passingTests: 0,
			},
			want: 0.2,
		},
		{
			name: "Compiles and 2 tests pass",
			args: args{
				isPassing:    false,
				compiles:     true,
				totalTests:   8,
				passingTests: 2,
			},
			want: 0.4,
		},
		{
			name: "Compiles and 6 tests pass",
			args: args{
				isPassing:    false,
				compiles:     true,
				totalTests:   8,
				passingTests: 6,
			},
			want: 0.8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateScore(tt.args.isPassing, tt.args.compiles, tt.args.totalTests, tt.args.passingTests); got != tt.want {
				t.Errorf("calculateScore() = %v, want %v", got, tt.want)
			}
		})
	}
}
