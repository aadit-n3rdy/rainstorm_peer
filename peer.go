package main

import (
	"fmt"
	"strings"
)

func main() {
	done := false
	var s string;
	for !done {
		fmt.Scanln(&s);
		tokens := strings.Fields(s)
		if len(tokens) == 0 {
			continue;
		}
	}
	return;
}
