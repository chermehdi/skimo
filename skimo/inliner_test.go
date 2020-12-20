package skimo

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strings"
	"testing"
)

type MockProvider struct{}
type MockProviderMulti struct{}

func (pr *MockProvider) GetReader(path string) (io.Reader, error) {
	if path == "include/strings.h" {
		return strings.NewReader(""), nil
	}
	if path == "include/lib/constants.h" {
		return strings.NewReader(""), nil
	}
	if path == "include/lib/tree.h" {
		return strings.NewReader(""), nil
	}
	return nil, errors.New("Failed to find given path")
}

func (pr *MockProviderMulti) GetReader(path string) (io.Reader, error) {
	if path == "include/lib/tree.h" {
		return strings.NewReader(`#include "../io/tree_reader.h"
		const int INF = 1 << 30;`), nil
	}
	if path == "include/lib/constants.h" {
		return strings.NewReader(""), nil
	}
	if path == "include/tree_reader.h" {
		return strings.NewReader(""), nil
	}
	if path == "include/io/tree_reader.h" {
		return strings.NewReader(""), nil
	}
	return nil, errors.New("Failed to find given path")
}

func TestExtractLinks(t *testing.T) {
	// #include "something.(h|hpp|cc|cpp)"
	// #include "something
	// #include "%something.hpp"
	// #include "(../)/something*/something.(h|hpp|cc|cpp)"
	provider := &MockProvider{}
	links, _ := ExtractLinks("include", `
		#include "strings.h"
		#include "lib/constants.h"
		#include "lib/tree.h"
		#include <iostream>
    `, provider, false)
	assert.Equal(t, 3, len(links))
	assert.Equal(t, "include/strings.h", links[0].to)
	assert.Equal(t, "include/lib/constants.h", links[1].to)
	assert.Equal(t, "include/lib/tree.h", links[2].to)
}

func TestExtractLinksMultiLevel(t *testing.T) {
	// #include "something.(h|hpp|cc|cpp)"
	// #include "something
	// #include "%something.hpp"
	// #include "(../)/something*/something.(h|hpp|cc|cpp)"
	provider := &MockProviderMulti{}
	links, _ := ExtractLinks("include", `
		#include "lib/tree.h"
		#include "lib/constants.h"
		#include <iostream>
    `, provider, false)
	assert.Equal(t, 3, len(links))
	assert.Equal(t, "include/lib/tree.h", links[0].to)
	assert.Equal(t, "include/io/tree_reader.h", links[1].to)
	assert.Equal(t, "include/lib/constants.h", links[2].to)
}

func TestExtractLinksNoLinks(t *testing.T) {

	provider := &MockProvider{}
	links, _ := ExtractLinks("include", `
		#include <vector>
		#include <iostream>
    `, provider, false)

	assert.Equal(t, 0, len(links))
}

func TestInliner_NothingToInline(t *testing.T) {
	provider := &MockProvider{}
	inliner, err := NewInlinerWithProvider("include", []string{}, false, provider)
	assert.NoError(t, err)
	cppFile := `#include <iostream>
		#include <vector>

		int main() {
		}
	`
	content, err := inliner.Inline(strings.NewReader(cppFile))
	assert.NoError(t, err)
	assert.Equal(t, cppFile, content)
}

func TestInliner_OneDependencyLevel(t *testing.T) {
	provider := &MockProviderMulti{}
	inliner, err := NewInlinerWithProvider("include", []string{}, false, provider)
	assert.NoError(t, err)
	cppFile := `#include <iostream>
		#include <vector>
		#include "lib/tree.h" 
		int main() {
		}
	`
	content, err := inliner.Inline(strings.NewReader(cppFile))
	assert.NoError(t, err)
	assert.True(t, strings.Contains(content, "#include <vector>"))
	assert.False(t, strings.Contains(content, "#include \"lib/tree.h\""))
	assert.True(t, strings.Contains(content, "const int INF = 1 << 30;"))
}

func TestInliner_MultipleDependencyLevels(t *testing.T) {
	provider := &MockProviderMulti{}
	inliner, err := NewInlinerWithProvider("include", []string{}, false, provider)
	assert.NoError(t, err)
	cppFile := `#include <iostream>
		#include <vector>
		#include "tree_reader.h"
		#include "lib/constants.h"
		#include "lib/tree.h" 
		int main() {
		}
	`
	content, err := inliner.Inline(strings.NewReader(cppFile))
	assert.NoError(t, err)
	assert.True(t, strings.Contains(content, "#include <vector>"))
	assert.False(t, strings.Contains(content, "#include \"lib/tree.h\""))
	assert.False(t, strings.Contains(content, "#include \"io/tree_reader.h\""))
	assert.False(t, strings.Contains(content, "#include \"tree_reader.h\""))
	assert.True(t, strings.Contains(content, "const int INF = 1 << 30;"))
}

func TestInliner_StripFirstN(t *testing.T) {
	// Happy path
	value, err := stringFirstNParts([]string{"hello", "world", "test"}, 1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(strings.Split(value, string(os.PathSeparator))))
	// Don't remove anything
	value, err = stringFirstNParts([]string{"hello", "world", "test"}, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(strings.Split(value, string(os.PathSeparator))))
	// Panic
	_, err = stringFirstNParts([]string{"hello", "world", "test"}, 4)
	assert.Error(t, err)
}

func TestInliner_LeavesDocumentationComments(t *testing.T) {
	provider := &MockProviderMulti{}
	inliner, err := NewInlinerWithProvider("include", []string{}, false, provider)
	assert.NoError(t, err)
	cppFile := "//\n" + "// @author MaxHeap\n" + "// Some other comment\n" + "#include <iostream>\n" + "#include <vector>\n" + "#include \"lib/tree.h\"" + "int main() {\n" + "}"
	content, err := inliner.Inline(strings.NewReader(cppFile))
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(content, "//\n// @author MaxHeap\n// Some other comment"))
	assert.True(t, strings.Contains(content, "#include <vector>"))
	assert.False(t, strings.Contains(content, "#include \"lib/tree.h\""))
	assert.True(t, strings.Contains(content, "const int INF = 1 << 30;"))
}
