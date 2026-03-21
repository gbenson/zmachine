package util

import (
	"fmt"
	"strings"
)

func ComponentName(comp any) string {
	if name, ok := comp.(string); ok {
		return name
	}
	name := fmt.Sprintf("%T", comp)    // *package.Type
	name = strings.TrimLeft(name, "*") // package.Type
	return name
}
