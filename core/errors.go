package core

import "fmt"

type InvalidFrequencyError string

func (e InvalidFrequencyError) Error() string {
	return fmt.Sprintf("invalid frequency %q", string(e))
}
