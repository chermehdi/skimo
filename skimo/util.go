package skimo

import (
	"bufio"
)

type PathPredicate func(string) bool

func readContent(path string, provider ReaderProvider) string {
	truePredicate := func(string) bool {
		return true
	}
	return readContentWithPredicate(path, provider, truePredicate)
}

func readContentExcludeIncludes(path string, provider ReaderProvider) string {
	predicate := func(currentLine string) bool {
		if IncludeRegexStd.MatchString(currentLine) || IncludeRegex.MatchString(currentLine) {
			return false
		}
		return true
	}
	return readContentWithPredicate(path, provider, predicate)
}

func readContentWithPredicate(path string, provider ReaderProvider, predicate PathPredicate) string {
	reader, _ := provider.GetReader(path)
	scanner := bufio.NewScanner(reader)
	content := ""
	for scanner.Scan() {
		currentLine := scanner.Text()
		if predicate(currentLine) {
			content += currentLine
			content += "\n"
		}
		if IncludeRegexStd.MatchString(currentLine) || IncludeRegex.MatchString(currentLine) {
			continue
		}
	}
	return content
}
