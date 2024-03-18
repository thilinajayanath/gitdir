package git

import "testing"

func TestParseAuth(t *testing.T) {
	type test struct {
		content []string
		domain  string
		want    [2]string
	}

	tests := []test{
		{
			content: []string{"user-name:password123456!@#$%^@github.com"},
			domain:  "github.com",
			want:    [2]string{"user-name", "password123456!@#$%^"},
		},
		{
			content: []string{
				"https://test123:test123@gitlab.com",
				"https://asd:asd@github.com",
			},
			domain: "gitlab.com",
			want:   [2]string{"test123", "test123"},
		},
	}

	for _, tc := range tests {
		ret := parseAuth(tc.content, tc.domain)

		if ret[0].username != tc.want[0] || ret[0].password != tc.want[1] {
			t.Errorf("Wanted: %v received: %v\n", tc.want, ret)
		}
	}
}
