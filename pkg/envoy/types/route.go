package types

import (
	"fmt"
	"strings"
)

func GenerateRouteName(path, method string) string {
	return fmt.Sprintf("%s-%s", path, strings.ToUpper(method))
}
