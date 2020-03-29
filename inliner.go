package main

import (
	"bufio"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
)

var IncludeRegex = regexp.MustCompile("#include\\s+\"(?P<path>.*?)\"")
var IncludeRegexStd = regexp.MustCompile("#include\\s+<(.+)>")

type ReaderProvider interface {
	GetReader(string) (io.Reader, error)
}
type DefaultReaderProvider struct {
}

func (rp *DefaultReaderProvider) GetReader(path string) (io.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return bufio.NewReader(file), nil
}

type Inliner struct {
	includeDir string
	excludes   []*regexp.Regexp
}

type Link struct {
	from string
	to   string
}

func readContent(path string, provider ReaderProvider) string {
	reader, _ := provider.GetReader(path)
	scanner := bufio.NewScanner(reader)
	content := ""
	for scanner.Scan() {
		content += scanner.Text()
		content += "\n"
	}
	return content
}

// Extract links
// currentRoot should always be a directory.
func ExtractLinks(currentRoot string, content string, provider ReaderProvider) []Link {
	result := make([]Link, 0)
	set := NewSet()
	extractLinks(currentRoot, content, &set, &result, provider)
	return result
}

func extractLinks(currentRoot string, content string, set *Set, result *[]Link, provider ReaderProvider) {
	if set.Has(currentRoot) {
		return
	}
	set.Add(currentRoot)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if IncludeRegex.MatchString(line) {
			filePath := IncludeRegex.FindStringSubmatch(line)[1]
			newFilePath := path.Join(currentRoot, filePath)
			*result = append(*result, Link{currentRoot, newFilePath})
			newFileContent := readContent(newFilePath, provider)
			extractLinks(path.Dir(newFilePath), newFileContent, set, result, provider)
		}
	}
}

// Creates a new instance of the Inliner type.
// includeDir: the root directory of your cpp library.
// excludes: regular expressions denoting file paths to ignore from inlining
func NewInliner(includeDir string, excludes []string) (*Inliner, error) {
	n := len(excludes)
	excludesRegex := make([]*regexp.Regexp, n)
	for i, value := range excludes {
		compiledRegex, err := regexp.Compile(value)
		if err != nil {
			return nil, err
		}
		excludesRegex[i] = compiledRegex
	}
	return &Inliner{includeDir, excludesRegex}, nil
}

// Returns true if the current line starts with a #include
func isIncludeLine(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " "), "#include")
}

// Returns a string aggregating all the contents of the files given by the argument array.
func filesContent(order []string) string {
	return ""
}

func (inliner *Inliner) Inline(reader io.Reader) (string, error) {
	scanner := bufio.NewScanner(reader)
	lines := make([]string, 0)
	seenIncludes := NewSet()
	insertPosition, lineNumber := 0, 0
	for scanner.Scan() {
		currentLine := scanner.Text()
		if !IncludeRegex.MatchString(currentLine) {
			lines = append(lines, currentLine)
		}
		if isIncludeLine(currentLine) && IncludeRegexStd.MatchString(currentLine) {
			libraryName := IncludeRegexStd.FindStringSubmatch(currentLine)[1]
			seenIncludes.Add(libraryName)
			insertPosition = lineNumber
		}
		lineNumber++
	}
	provider := DefaultReaderProvider{}
	links := ExtractLinks(inliner.includeDir, strings.Join(lines, "\n"), &provider)
	graph := NewGraph(links, seenIncludes)
	fileOrder := graph.GetTopologicalOrder(inliner.includeDir)
	content := ""
	// a -> b
	// b -> c
	// a -> d
	// c -> d
	for i, _ := range lines {
		if i == insertPosition {
			content += filesContent(fileOrder)
		}
	}
	return content, nil
}
