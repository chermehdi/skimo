package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
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
		return strings.NewReader(`#include "../io/tree_reader.h"`), nil
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
	links := ExtractLinks("include", `
		#include "strings.h"
		#include "lib/constants.h"
		#include "lib/tree.h"
		#include <iostream>
    `, provider)
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
	links := ExtractLinks("include", `
		#include "lib/tree.h"
		#include "lib/constants.h"
		#include <iostream>
    `, provider)
	assert.Equal(t, 3, len(links))
	assert.Equal(t, "include/lib/tree.h", links[0].to)
	assert.Equal(t, "include/io/tree_reader.h", links[1].to)
	assert.Equal(t, "include/lib/constants.h", links[2].to)
}

func TestExtractLinksNoLinks(t *testing.T) {

	provider := &MockProvider{}
	links := ExtractLinks("include", `
		#include <vector>
		#include <iostream>
    `, provider)

	assert.Equal(t, 0, len(links))
}
