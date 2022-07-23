package jazz

import "fmt"

func ReportErr(line int, msg string) {
	fmt.Printf("[line %d] error: %s\n", line, msg)
}
