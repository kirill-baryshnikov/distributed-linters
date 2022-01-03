package main

import (
	"fmt"
	"log"
    "strings"
)

func lintSourceCode(codeToLint string, language Language) string {
	lintedLines := []string {}
	for index, line := range strings.Split(strings.TrimSuffix(codeToLint, "\n"), "\n") {
     	lintedLines = append(lintedLines, lint(line, language))
     	log.Println(fmt.Sprintf("Linted line %d", index))
	}

	return strings.Join(lintedLines[:], "\n")
}

func lint(contentToLint string, language Language) string {
	lintedContent := []string {}

	for index, character := range contentToLint {
		if character == '=' {
			if index == 0 || contentToLint[index - 1] != ' ' {
				lintedContent = append(lintedContent, " ")
			}

			lintedContent = append(lintedContent, string(character))

			if index != len(contentToLint) - 1 && contentToLint[index + 1] != ' ' && !(language == Python && contentToLint[index + 1] == '\\') {
				lintedContent = append(lintedContent, " ")
			}
		} else {
			lintedContent = append(lintedContent, string(character))
		}	
	}

	return strings.Join(lintedContent[:], "")
}
