package knowledge

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Chunk 表示用于检索的文档文本块。
type Chunk struct {
	ID         string
	DocumentID string
	ChapterID  string
	// HeadingPath 表示文本块来源小节的共同标题层级，用于补充检索上下文。
	HeadingPath []string
	// Position 表示文本块在文档中的序号，从 0 开始。
	Position int
	Content  string
	// TokenCount 表示 Content 按切分所用分词规则计算的词元数量。
	TokenCount int
	// SectionPaths 表示文本块覆盖的原始小节路径。
	SectionPaths []string
}

// splitChunks 解析文档章节并切分为不重叠的文本块。
func splitChunks(doc Document, config SplitConfig) ([]Chunk, error) {
	if config.ChapterMaxTokens == 0 {
		config.ChapterMaxTokens = 512
	}
	if config.ChunkMaxTokens == 0 {
		config.ChunkMaxTokens = 512
	}

	// 1. 规范化正文换行，使用 Markdown 解析库识别顶层标题，再用栈组织标题树。
	content := strings.ReplaceAll(doc.Content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")
	root := parseSections(lines)

	// 2. 确定各标题节点的正文范围，保留文档开头没有标题的正文。
	assignSectionEnds(root, len(lines))
	chapters := make([]splitChapter, 0)
	preambleEnd := len(lines)
	if len(root.children) > 0 {
		preambleEnd = root.children[0].start
	}
	if preamble := strings.TrimSpace(strings.Join(lines[:preambleEnd], "\n")); preamble != "" {
		chapters = append(chapters, splitChapter{path: "body", pieces: []sectionPiece{{path: "0", content: preamble}}})
	}

	// 3. 按章节预算自底向上组织章节：父章节及子章节整体能放下时合并。
	//    整体超出预算时，保留父章节自身正文，并在预算内合并相邻的小章节。
	//    单节自身超出预算时仍保持完整，交给后续文本块切分。
	for _, section := range root.children {
		chapters = append(chapters, organizeChapters(section, lines, config.ChapterMaxTokens)...)
	}

	// 4. 逐个章节按文本块上限切分：能放下时生成一个文本块，否则按固定词元窗口切分，不重叠。
	//    章节预算控制内容组织，文本块上限控制最终块大小，两者独立。
	// 5. 为文本块记录文档 ID、章节 ID、标题层级、文档内的块序号和来源小节路径。
	chunks := make([]Chunk, 0)
	for _, chapter := range chapters {
		text, sources := chapter.render()
		tokens := splitTokenSpans(text)
		for start := 0; start < len(tokens) || start == 0; start += config.ChunkMaxTokens {
			end := min(start+config.ChunkMaxTokens, len(tokens))
			from, to := 0, len(text)
			if start > 0 {
				from = tokens[start].start
			}
			if end < len(tokens) {
				to = tokens[end].start
			}
			paths := make([]string, 0)
			var headingPath []string
			for _, source := range sources {
				if source.start < to && source.end > from {
					if len(paths) == 0 {
						headingPath = append([]string(nil), source.headingPath...)
					} else {
						common := 0
						for common < len(headingPath) && common < len(source.headingPath) && headingPath[common] == source.headingPath[common] {
							common++
						}
						headingPath = headingPath[:common]
					}
					paths = append(paths, source.path)
				}
			}
			position := len(chunks)
			chunks = append(chunks, Chunk{
				ID:           fmt.Sprintf("%s:chunk:%d", doc.ID, position),
				DocumentID:   doc.ID,
				ChapterID:    doc.ID + ":" + chapter.path,
				HeadingPath:  headingPath,
				Position:     position,
				Content:      text[from:to],
				TokenCount:   end - start,
				SectionPaths: paths,
			})
			if end == len(tokens) {
				break
			}
		}
	}

	// 6. 按文档顺序汇总并返回文本块，章节仅作为内部处理结果，不单独返回。
	return chunks, nil
}

// sectionNode 记录标题树及其在原始正文中的行范围。
type sectionNode struct {
	headingPath []string
	level       int
	path        string
	start       int
	body        int
	end         int
	children    []*sectionNode
}

// parseSections 使用 Markdown 解析器识别标题，再以栈组织章节层级。
// 仅文档顶层标题划分章节，列表、引用、代码和 HTML 块中的内容保留在所属章节内。
func parseSections(lines []string) *sectionNode {
	source := []byte(strings.Join(lines, "\n"))
	document := goldmark.New().Parser().Parse(text.NewReader(source))
	lineStarts := make([]int, len(lines))
	for line := 1; line < len(lines); line++ {
		lineStarts[line] = lineStarts[line-1] + len(lines[line-1]) + 1
	}
	lineAt := func(offset int) int {
		return sort.Search(len(lineStarts), func(i int) bool { return lineStarts[i] > offset }) - 1
	}

	root := &sectionNode{}
	stack := []*sectionNode{root}
	for block := document.FirstChild(); block != nil; block = block.NextSibling() {
		heading, ok := block.(*ast.Heading)
		if !ok {
			continue
		}
		start := lineAt(heading.Pos())
		body := start + 1
		// 下划线式标题的源位置与首行正文相同，还需包含最后一行标题后的下划线。
		if heading.Lines().Len() > 0 && heading.Lines().At(0).Start == heading.Pos() {
			last := heading.Lines().At(heading.Lines().Len() - 1)
			body = lineAt(max(last.Start, last.Stop-1)) + 2
		}
		for len(stack) > 1 && stack[len(stack)-1].level >= heading.Level {
			stack = stack[:len(stack)-1]
		}
		parent := stack[len(stack)-1]
		path := fmt.Sprint(len(parent.children) + 1)
		if parent.path != "" {
			path = parent.path + "." + path
		}
		title := strings.Join(strings.Fields(string(heading.Text(source))), " ")
		headingPath := append(append([]string(nil), parent.headingPath...), title)
		node := &sectionNode{level: heading.Level, path: path, start: start, body: body, headingPath: headingPath}
		parent.children = append(parent.children, node)
		stack = append(stack, node)
	}
	return root
}

func assignSectionEnds(node *sectionNode, end int) {
	node.end = end
	for index, child := range node.children {
		boundary := end
		if index+1 < len(node.children) {
			boundary = node.children[index+1].start
		}
		assignSectionEnds(child, boundary)
	}
}

type sectionPiece struct {
	path        string
	content     string
	headingPath []string
}

type splitChapter struct {
	path   string
	pieces []sectionPiece
}

type sectionSpan struct {
	path        string
	start, end  int
	headingPath []string
}

// render 保留小节原文，并记录各小节在合并正文中的字节范围。
func (c splitChapter) render() (string, []sectionSpan) {
	var text strings.Builder
	spans := make([]sectionSpan, 0, len(c.pieces))
	for _, piece := range c.pieces {
		if text.Len() > 0 {
			text.WriteString("\n\n")
		}
		start := text.Len()
		text.WriteString(piece.content)
		spans = append(spans, sectionSpan{path: piece.path, start: start, end: text.Len(), headingPath: piece.headingPath})
	}
	return text.String(), spans
}

// organizeChapters 先组织子章节，再决定整体合并或在父节点内打包相邻章节。
func organizeChapters(node *sectionNode, lines []string, budget int) []splitChapter {
	children := make([]splitChapter, 0)
	for _, child := range node.children {
		children = append(children, organizeChapters(child, lines, budget)...)
	}
	ownEnd := node.end
	if len(node.children) > 0 {
		ownEnd = node.children[0].start
	}
	own := sectionPiece{path: node.path, content: strings.TrimSpace(strings.Join(lines[node.start:ownEnd], "\n")), headingPath: node.headingPath}
	merged := splitChapter{path: node.path, pieces: []sectionPiece{own}}
	for _, child := range children {
		merged.pieces = append(merged.pieces, child.pieces...)
	}
	text, _ := merged.render()
	if len(splitTokenSpans(text)) <= budget {
		return []splitChapter{merged}
	}
	if len(children) == 0 {
		return []splitChapter{merged}
	}
	out := make([]splitChapter, 0, len(children)+1)
	// 无独立正文的父标题附在首个子章节上，避免产生只有标题的碎片。
	if strings.TrimSpace(strings.Join(lines[node.body:ownEnd], "\n")) == "" {
		children[0].pieces = append([]sectionPiece{own}, children[0].pieces...)
	} else {
		out = append(out, splitChapter{path: node.path, pieces: []sectionPiece{own}})
	}
	var packed *splitChapter
	flush := func() {
		if packed != nil {
			out = append(out, *packed)
			packed = nil
		}
	}
	for _, child := range children {
		text, _ := child.render()
		if len(splitTokenSpans(text)) > budget {
			flush()
			out = append(out, child)
			continue
		}
		if packed != nil {
			candidate := splitChapter{pieces: append(append([]sectionPiece(nil), packed.pieces...), child.pieces...)}
			combined, _ := candidate.render()
			if len(splitTokenSpans(combined)) > budget {
				flush()
			}
		}
		if packed == nil {
			packed = &splitChapter{path: fmt.Sprintf("%s/pack-%d", node.path, len(out)+1)}
		}
		packed.pieces = append(packed.pieces, child.pieces...)
	}
	flush()
	return out
}

type splitTokenSpan struct {
	start int
	end   int
}

// splitTokenSpans 将常用汉字逐字计数，其余连续字母和数字计作一个词元。
// 空白和标点不计数，但切分时保留原始字符；此计数不代表模型的实际词元数。
func splitTokenSpans(text string) []splitTokenSpan {
	tokens := make([]splitTokenSpan, 0)
	inWord := false
	for index, char := range text {
		switch {
		case char >= 0x4E00 && char <= 0x9FFF:
			tokens = append(tokens, splitTokenSpan{start: index, end: index + utf8.RuneLen(char)})
			inWord = false
		case unicode.IsLetter(char) || unicode.IsDigit(char):
			if !inWord {
				tokens = append(tokens, splitTokenSpan{start: index, end: index + utf8.RuneLen(char)})
			} else {
				tokens[len(tokens)-1].end = index + utf8.RuneLen(char)
			}
			inWord = true
		default:
			inWord = false
		}
	}
	return tokens
}
