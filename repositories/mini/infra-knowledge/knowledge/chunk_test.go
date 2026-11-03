package knowledge

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitChunksFixedWindows(t *testing.T) {
	content := "# Guide\n\n甲，乙！丙。丁？戊；己。"
	doc := Document{ID: "doc", Content: content}
	chunks, err := splitChunks(doc, SplitConfig{ChapterMaxTokens: 2, ChunkMaxTokens: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %+v", chunks)
	}
	var rebuilt strings.Builder
	for position, chunk := range chunks {
		if chunk.DocumentID != doc.ID || chunk.ChapterID != "doc:1" || chunk.Position != position {
			t.Fatalf("unexpected chunk identity: %+v", chunk)
		}
		if chunk.TokenCount > 3 || chunk.TokenCount != len(splitTokenSpans(chunk.Content)) {
			t.Fatalf("invalid token count: %+v", chunk)
		}
		if !utf8.ValidString(chunk.Content) || !reflect.DeepEqual(chunk.SectionPaths, []string{"1"}) {
			t.Fatalf("invalid content or source paths: %+v", chunk)
		}
		rebuilt.WriteString(chunk.Content)
	}
	if rebuilt.String() != content {
		t.Fatalf("content lost or duplicated: %q", rebuilt.String())
	}
	again, err := splitChunks(doc, SplitConfig{ChapterMaxTokens: 2, ChunkMaxTokens: 3})
	if err != nil || !reflect.DeepEqual(chunks, again) {
		t.Fatal("splitting is not deterministic")
	}
}

func TestSplitChunksChapterPacking(t *testing.T) {
	doc := Document{ID: "doc", Content: "# Root\nintro\n## A\none two\n## B\nthree four\n## C\nfive six"}
	chunks, err := splitChunks(doc, SplitConfig{ChapterMaxTokens: 6, ChunkMaxTokens: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected parent body and two sibling packs, got %+v", chunks)
	}
	if chunks[1].TokenCount != 6 || !strings.Contains(chunks[1].Content, "## A") || !strings.Contains(chunks[1].Content, "## B") {
		t.Fatalf("adjacent children were not packed within budget: %+v", chunks[1])
	}
	if !reflect.DeepEqual(chunks[1].SectionPaths, []string{"1.1", "1.2"}) {
		t.Fatalf("lost original section paths: %+v", chunks[1])
	}
	if !strings.Contains(chunks[2].Content, "## C") {
		t.Fatal("document order was not preserved")
	}
}

func TestSplitChunksIndependentLimits(t *testing.T) {
	doc := Document{ID: "doc", Content: "# Root\n## A\none two\n## B\nthree four"}
	chunks, err := splitChunks(doc, SplitConfig{ChapterMaxTokens: 7, ChunkMaxTokens: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected one merged chapter split into three chunks, got %+v", chunks)
	}
	for _, chunk := range chunks {
		if chunk.ChapterID != "doc:1" || chunk.TokenCount > 3 {
			t.Fatalf("unexpected chapter or chunk limit: %+v", chunk)
		}
	}
}

func TestSplitChunksPreambleAndHeadinglessContent(t *testing.T) {
	for _, tt := range []struct {
		name, content string
		chapters      []string
	}{
		{"plain", "plain text", []string{"doc:body"}},
		{"punctuation", "... !!!", []string{"doc:body"}},
		{"preamble", "intro\r\n\r\n# Topic\rbody", []string{"doc:body", "doc:1"}},
		{"setext", "Topic\n=====\nbody", []string{"doc:1"}},
		{"duplicates", "# Same\none\n# Same\ntwo", []string{"doc:1", "doc:2"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := splitChunks(Document{ID: "doc", Content: tt.content}, SplitConfig{})
			if err != nil {
				t.Fatal(err)
			}
			if len(chunks) != len(tt.chapters) {
				t.Fatalf("unexpected chunks: %+v", chunks)
			}
			ids := make(map[string]bool)
			for i, chunk := range chunks {
				if chunk.ChapterID != tt.chapters[i] || ids[chunk.ID] || strings.Contains(chunk.Content, "\r") {
					t.Fatalf("unexpected chunk: %+v", chunk)
				}
				ids[chunk.ID] = true
			}
		})
	}
}

func TestParseSectionsCodeAndHierarchy(t *testing.T) {
	lines := strings.Split("# Root\n````go\n# Fake\n```\n## StillFake\n````\n~~~\n# AlsoFake\n~~~\n    # Indented\n## Child\n### Deep\n## Sibling\n# Other", "\n")
	root := parseSections(lines)
	assignSectionEnds(root, len(lines))
	if len(root.children) != 2 || len(root.children[0].children) != 2 {
		t.Fatalf("incorrect hierarchy or code parsed as headings: %+v", root)
	}
	first := root.children[0]
	if first.end != 13 || first.children[0].end != 12 || len(first.children[0].children) != 1 {
		t.Fatalf("incorrect section boundaries: %+v", first)
	}
	if first.children[0].children[0].path != "1.1.1" || root.children[1].path != "2" {
		t.Fatal("incorrect hierarchical paths")
	}
}

func TestSplitChunksPreservesParentHeading(t *testing.T) {
	content := "# Root\n## A\none two\n## B\nthree four"
	chunks, err := splitChunks(Document{ID: "doc", Content: content}, SplitConfig{ChapterMaxTokens: 3, ChunkMaxTokens: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 2 || !strings.HasPrefix(chunks[0].Content, "# Root") || !strings.Contains(chunks[1].Content, "## B") {
		t.Fatalf("parent heading or children lost: %+v", chunks)
	}
}

func TestSplitChunksDefaultWindow(t *testing.T) {
	chunks, err := splitChunks(Document{ID: "doc", Content: strings.Repeat("word ", 513)}, SplitConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 2 || chunks[0].TokenCount != 512 || chunks[1].TokenCount != 1 {
		t.Fatalf("incorrect default window: %+v", chunks)
	}
}

func TestSplitTokenSpans(t *testing.T) {
	text := "Redis 配置 maxmemory 512，UTF8!"
	spans := splitTokenSpans(text)
	if len(spans) != 6 {
		t.Fatalf("expected 6 tokens, got %d", len(spans))
	}
}

func TestParseSectionsMarkdownBlocks(t *testing.T) {
	for _, tt := range []struct {
		name, content string
		start, body   int
	}{
		{"multiline setext", "First line\nsecond line\n===\nbody", 0, 3},
		{"empty heading", "#\nbody", 0, 1},
		{"empty closing heading", "## ###\nbody", 0, 1},
		{"html", "<div>\n# Not a heading\n</div>\n\n# Real\nbody", 4, 5},
		{"comment", "<!--\n# Not a heading\n-->\n# Real\nbody", 3, 4},
		{"quote", "> # Quoted\n> text\n\n# Real\nbody", 3, 4},
		{"list", "- item\n\n  # Nested\n\n# Real\nbody", 4, 5},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := parseSections(strings.Split(tt.content, "\n"))
			if len(root.children) != 1 {
				t.Fatalf("expected one top-level heading, got %+v", root.children)
			}
			heading := root.children[0]
			if heading.start != tt.start || heading.body != tt.body {
				t.Fatalf("unexpected source range: start=%d body=%d", heading.start, heading.body)
			}
			chunks, err := splitChunks(Document{ID: "doc", Content: tt.content}, SplitConfig{})
			if err != nil {
				t.Fatal(err)
			}
			var content []string
			for _, chunk := range chunks {
				content = append(content, chunk.Content)
			}
			lines := strings.Split(tt.content, "\n")
			var expected []string
			if tt.start > 0 {
				expected = append(expected, strings.TrimSpace(strings.Join(lines[:tt.start], "\n")))
			}
			expected = append(expected, strings.Join(lines[tt.start:], "\n"))
			if !reflect.DeepEqual(content, expected) {
				t.Fatalf("source changed: got %q, want %q", content, expected)
			}
		})
	}
}
