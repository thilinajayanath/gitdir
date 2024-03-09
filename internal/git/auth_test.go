package git

import "testing"

func TestParseAuth(t *testing.T) {
	testInput := []string{
		"test:test@github.com",
	}

	ret := parseAuth(testInput, "github.com")

	if ret[0].username != "test" || ret[0].password != "test" {
		t.Errorf("error %s %s", ret[0], ret[1])
	}
}
