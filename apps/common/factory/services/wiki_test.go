package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWiki_InitTemplate(t *testing.T) {
	dir := t.TempDir()
	wikiPath := filepath.Join(dir, wikiDir, wikiFile)

	// Not exists yet.
	content := readWiki(wikiPath)
	assert.Empty(t, content)

	// After init write.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, wikiDir), 0o755))
	initial := "# Project Wiki (auto-maintained by Factory)\n<!-- Last updated: 2026-01-01, task: t1 -->\n\n## Architecture\n\n## Conventions\n\n## Gotchas\n\n## Dependencies\n"
	require.NoError(t, os.WriteFile(wikiPath, []byte(initial), 0o644))

	content = readWiki(wikiPath)
	assert.Contains(t, content, "# Project Wiki")
	assert.Contains(t, content, "## Architecture")
	assert.Contains(t, content, "## Conventions")
	assert.Contains(t, content, "## Gotchas")
	assert.Contains(t, content, "## Dependencies")
}

func TestWiki_ReadProjectWiki(t *testing.T) {
	dir := t.TempDir()

	// Missing project path.
	assert.Empty(t, readProjectWiki(""))

	// Non-existent dir.
	assert.Empty(t, readProjectWiki(dir))

	// With wiki.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, wikiDir), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, wikiDir, wikiFile),
		[]byte("# wiki content"),
		0o644,
	))
	assert.Equal(t, "# wiki content", readProjectWiki(dir))
}

func TestWiki_LineCount(t *testing.T) {
	assert.Equal(t, 0, lineCount(""))
	assert.Equal(t, 1, lineCount("one line"))
	assert.Equal(t, 3, lineCount("a\nb\nc"))
}

func TestWiki_CompactWiki_UnderLimit(t *testing.T) {
	content := "# Title\n## Section\nline1\nline2"
	result := compactWiki(content, 200)
	assert.Equal(t, content, result)
}

func TestWiki_CompactWiki_OverLimit(t *testing.T) {
	// Build a wiki that exceeds maxLines.
	var lines []string
	lines = append(lines, "# Project Wiki")
	lines = append(lines, "<!-- Last updated: 2026-01-01, task: t1 -->")
	lines = append(lines, "## Architecture")
	for i := range 250 {
		lines = append(lines, strings.Repeat("x", 10)+string(rune('a'+i%26)))
	}
	content := strings.Join(lines, "\n")

	result := compactWiki(content, 200)
	assert.LessOrEqual(t, lineCount(result), 200)
	// Headers preserved.
	assert.Contains(t, result, "# Project Wiki")
	assert.Contains(t, result, "## Architecture")
}

func TestWiki_CompactWiki_PreservesHeaders(t *testing.T) {
	lines := []string{
		"# Project Wiki",
		"<!-- comment -->",
		"## Architecture",
		"detail line",
		"## Conventions",
		"convention line",
		"## Gotchas",
		"## Dependencies",
	}
	content := strings.Join(lines, "\n")
	result := compactWiki(content, 200)
	assert.Contains(t, result, "## Architecture")
	assert.Contains(t, result, "## Conventions")
	assert.Contains(t, result, "## Gotchas")
	assert.Contains(t, result, "## Dependencies")
}
