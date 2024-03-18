package git

import (
	"testing"
)

func TestGetDomain(t *testing.T) {
	type test struct {
		err   string
		input string
		want  string
	}

	tests := []test{
		{
			err:   "",
			input: "https://github.com/example/example.git",
			want:  "github.com",
		},
		{
			err:   "",
			input: "git@gitlab.com:example/example.git",
			want:  "gitlab.com",
		},
		{
			err:   "git repo url is invalid",
			input: "qwerty",
			want:  "",
		},
	}

	for _, tc := range tests {
		ret, err := getDomain(tc.input)

		if err != nil {
			if tc.err != err.Error() {
				t.Errorf("Expected error %s received error:%s\n", tc.err, err)
			}
		}

		if err == nil {
			if ret != tc.want {
				t.Errorf("Wanted %s received %s\n", tc.want, ret)
			}

			if tc.err != "" {
				t.Error("Expected error but received none")
			}
		}
	}
}
