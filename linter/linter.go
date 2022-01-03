package main

import (
    "strings"
)

func lintSourceCode(contentToLint string) string {
	return strings.ReplaceAll(contentToLint, " ", "_")
}
