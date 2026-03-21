package util

import "fmt"

type NotImplementedError string

func (e NotImplementedError) Error() string {
	return fmt.Sprintf("%s: not implemented", string(e))
}
