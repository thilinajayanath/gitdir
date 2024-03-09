package path

import (
	"fmt"
	"strings"
)

func FulltPath(start, end string) string {
	splitPath := strings.Split(fmt.Sprintf("%s/%s", start, end), "/")

	path := []string{}

	for _, v := range splitPath {
		if v != "" {
			path = append(path, v)
		}
	}

	return fmt.Sprintf("/%s", strings.Join(path, "/"))
}
