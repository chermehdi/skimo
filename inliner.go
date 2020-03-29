package main

import (
	"bufio"
	"fmt"
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
	includeDir     string
	excludes       []*regexp.Regexp
	readerProvider ReaderProvider
	verbose        bool
}

type Link struct {
	from string
	to   string
}

type IncludeSet struct {
	includes []string
}

func NewIncludeSet() *IncludeSet {
	return &IncludeSet{
		includes: make([]string, 0),
	}
}

func (is *IncludeSet) Add(include string) {
	is.includes = append(is.includes, include)
}

// Get the set of unique includes found in the include set, by their insertion order.
// This is done by starting with the base includes, then moving to the next ones.
func (is *IncludeSet) GetUniqueOrdered(base []string) []string {
	result := make([]string, 0)
	seen := NewSet()
	for _, include := range base {
		if !seen.Has(include) {
			seen.Add(include)
			result = append(result, include)
		}
	}
	for _, include := range is.includes {
		if !seen.Has(include) {
			seen.Add(include)
			result = append(result, include)
		}
	}
	return result
}

// Extract links
// currentRoot should always be a directory.
func ExtractLinks(currentRoot string, content string, provider ReaderProvider, verbose bool) ([]Link, *IncludeSet) {
	result := make([]Link, 0)
	set := NewSet()
	includeSet := NewIncludeSet()
	// TODO(chermehdi): cleanup, too many arguments
	extractLinks(currentRoot, "main.cpp", content, &set, &result, provider, includeSet, &verbose)
	return result, includeSet
}

func extractLinks(currentRoot string, lastFile string, content string, set *Set, result *[]Link, provider ReaderProvider, includeSet *IncludeSet, verbose *bool) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if isIncludeLine(line) {
			includeSet.Add(line)
		}
		if IncludeRegex.MatchString(line) {
			filePath := IncludeRegex.FindStringSubmatch(line)[1]
			newFilePath := path.Join(currentRoot, filePath)
			if set.Has(newFilePath) {
				continue
			}
			set.Add(newFilePath)
			*result = append(*result, Link{lastFile, newFilePath})
			if *verbose {
				_, _ = fmt.Fprintf(os.Stderr, "Reading the file %s\n", newFilePath)
			}
			newFileContent := readContent(newFilePath, provider)
			extractLinks(path.Dir(newFilePath), newFilePath, newFileContent, set, result, provider, includeSet, verbose)
		}
	}
}

func NewInlinerWithProvider(includeDir string, excludes []string, verbose bool, provider ReaderProvider) (*Inliner, error) {
	n := len(excludes)
	excludesRegex := make([]*regexp.Regexp, n)
	for i, value := range excludes {
		compiledRegex, err := regexp.Compile(value)
		if err != nil {
			return nil, err
		}
		excludesRegex[i] = compiledRegex
	}
	return &Inliner{includeDir, excludesRegex, provider, verbose}, nil
}

// Creates a new instance of the Inliner type.
// includeDir: the root directory of your cpp library.
// excludes: regular expressions denoting file paths to ignore from inlining
func NewInliner(includeDir string, verbose bool, excludes []string) (*Inliner, error) {
	return NewInlinerWithProvider(includeDir, excludes, verbose, &DefaultReaderProvider{})
}

// Returns true if the current line starts with a #include
func isIncludeLine(line string) bool {
	return IncludeRegexStd.MatchString(line)
}

// Returns a string aggregating all the contents of the files given by the argument array.
func filesContent(order []string, provider ReaderProvider) string {
	content := ""
	// process in reverse order
	n := len(order)
	for i := 0; i < n; i++ {
		content += fmt.Sprintf("// BEGIN %s\n", order[i])
		content += readContentExcludeIncludes(order[i], provider)
		content += fmt.Sprintf("// END %s\n", order[i])
	}
	return content
}

func (inliner *Inliner) Inline(reader io.Reader) (string, error) {
	scanner := bufio.NewScanner(reader)
	lines := make([]string, 0)
	seenIncludes := NewSet()
	insertPosition, lineNumber := 0, 0
	firstLevelIncludes := make([]string, 0)
	currentFileContent := ""
	for scanner.Scan() {
		currentLine := scanner.Text()
		if !IncludeRegex.MatchString(currentLine) {
			lines = append(lines, currentLine)
		}
		if isIncludeLine(currentLine) /* && IncludeRegexStd.MatchString(currentLine) */ {
			firstLevelIncludes = append(firstLevelIncludes, currentLine)
			insertPosition = lineNumber
		}
		currentFileContent += currentLine + "\n"
		lineNumber++
	}
	links, includeSet := ExtractLinks(inliner.includeDir, currentFileContent, inliner.readerProvider, inliner.verbose)
	graph := NewGraph(links, seenIncludes)
	fileOrder := graph.GetTopologicalOrder(inliner.includeDir)
	content := ""
	for _, include := range includeSet.GetUniqueOrdered(firstLevelIncludes) {
		content += include + "\n"
	}
	// a -> b
	// b -> c
	// a -> d
	// c -> d
	content += filesContent(fileOrder, inliner.readerProvider)
	for i := insertPosition + 1; i < len(lines); i++ {
		// we don't want a trailing new line
		if i > insertPosition+1 {
			content += "\n"
		}
		content += lines[i]
	}
	return content, nil
}
