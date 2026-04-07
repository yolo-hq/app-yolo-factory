package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuestionResolver_AutoResolve(t *testing.T) {
	claudeMD := `# yolo-factory

## Rules
- Integration tests only — no mocks, real DB
- actx.Resolve("Entity", id) + action.OK() — never return entity data directly
- fields.{key}= for field selection — bare fields= is rejected
- Framework first — if YOLO lacks a pattern, build in yolo/ before app code
`

	answer := autoResolve(claudeMD, "Should I use integration tests or unit tests with mocks?")
	assert.NotEmpty(t, answer)
	assert.Contains(t, answer, "Integration tests only")
}

func TestQuestionResolver_NoAutoResolve(t *testing.T) {
	claudeMD := `# yolo-factory

## Rules
- Integration tests only
- No mocks
`

	answer := autoResolve(claudeMD, "What database driver should I use for MySQL?")
	assert.Empty(t, answer)
}

func TestQuestionResolver_EmptyCLAUDEMD(t *testing.T) {
	answer := autoResolve("", "Any question here?")
	assert.Empty(t, answer)
}

func TestQuestionResolver_EmptyQuestion(t *testing.T) {
	answer := autoResolve("some content", "")
	assert.Empty(t, answer)
}

func TestExtractKeywords(t *testing.T) {
	kw := extractKeywords("Should I use integration tests or unit tests with mocks?")
	assert.Contains(t, kw, "integration")
	assert.Contains(t, kw, "tests")
	assert.Contains(t, kw, "unit")
	assert.Contains(t, kw, "mocks")
	// "should" is a stop word
	assert.NotContains(t, kw, "should")
	// "or", "I", "use" are too short
	assert.NotContains(t, kw, "or")
}

func TestExtractKeywords_NoDuplicates(t *testing.T) {
	kw := extractKeywords("test test test again test")
	count := 0
	for _, k := range kw {
		if k == "test" {
			count++
		}
	}
	assert.Equal(t, 1, count)
}
