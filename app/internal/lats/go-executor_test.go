package lats

import (
	"reflect"
	"testing"

	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
)

func Test_goExecutor_Execute(t *testing.T) {
	type args struct {
		code  string
		tests []string
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
				code: "```go" +`
				package main

				func salute(name string) string {
					return fmt.Sprintf("Hello, %s!", name)
				}
				` + "```",
				tests: []string{
					`
					func TestSalute(t *testing.T) {
						assert.Equal(t, "Hello, World!", salute("World"))
					}
					`,
					`
					func TestSalute(t *testing.T) {
						assert.Equal(t, "Hello, Ana!", salute("Ana"))
					}
					`,
					`
					func TestSalute(t *testing.T) {
						assert.NotEqual(t, "Hello, Ana!", salute("Ada"))
					}
					`,
				},
			},
			want: &models.ExecutionResult{
				IsPassing: true,
				Feedback:  "Tested passed:\n\n\t\t\t\t\tfunc TestSalute(t *testing.T) {\n\t\t\t\t\t\tassert.Equal(t, \"Hello, World!\", salute(\"World\"))\n\t\t\t\t\t}\n\t\t\t\t\t\n\n\t\t\t\t\tfunc TestSalute(t *testing.T) {\n\t\t\t\t\t\tassert.Equal(t, \"Hello, Ana!\", salute(\"Ana\"))\n\t\t\t\t\t}\n\t\t\t\t\t\n\n\t\t\t\t\tfunc TestSalute(t *testing.T) {\n\t\t\t\t\t\tassert.NotEqual(t, \"Hello, Ana!\", salute(\"Ada\"))\n\t\t\t\t\t}\n\t\t\t\t\t\n\nTested failed:\n",
				Score:     1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.Execute(tt.args.code, tt.args.tests)
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
		output string
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
			},
			want: []string{
				".\\lats.go:10:7: undefined: math\n",
				".\\lats.go:11:18: too many return values\n        have (bool, bool)\n        want (bool)\n",
				".\\lats.go:15:16: too many return values\n        have (bool, bool)\n        want (bool)\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := grabCompileErrs(tt.args.output, "lats.go"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("grabCompileErrs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_grabTestErrors(t *testing.T) {
	type args struct {
		output string
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
			},
			want: []string{
				"[Line] 53, [Reason] HasCloseElements() = false, want true",
				"[Line] 53, [Reason] HasCloseElements() = false, want true",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := grabTestErrors(tt.args.output); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("grabTestErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}
