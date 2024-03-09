package path

import "testing"

func TestFulltPath(t *testing.T) {
	type test struct {
		input []string
		want  string
	}

	tests := []test{
		{input: []string{"/var/tmp/", "/asd/asd/"}, want: "/var/tmp/asd/asd"},
		{input: []string{"var/tmp/", "asd/aa/"}, want: "/var/tmp/asd/aa"},
		{input: []string{"/var/tmp", "/sad/cc//"}, want: "/var/tmp/sad/cc"},
		{input: []string{"var/tmp", "//asd/s/d//"}, want: "/var/tmp/asd/s/d"},
		{input: []string{"/var/tmp/", "//asd//das//asd//"}, want: "/var/tmp/asd/das/asd"},
	}

	for _, tc := range tests {
		got := FulltPath(tc.input[0], tc.input[1])
		if got != tc.want {
			t.Errorf("wanted %s, got %s\n", tc.want, got)
		}
	}
}
