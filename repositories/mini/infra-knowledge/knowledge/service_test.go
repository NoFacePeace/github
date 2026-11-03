package knowledge

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestBuildRetrieveHits(t *testing.T) {
	documents := []Document{
		{ID: "a", ProjectID: "project", ProjectName: "Project", FileName: "a.md", Title: "First", Content: "# A\none two three four\n# B\nfive six"},
		{ID: "b", ProjectID: "project", FileName: "b.md", Title: "Second", Content: "# Other\nbody"},
	}
	first, err := splitChunks(documents[0], SplitConfig{ChapterMaxTokens: 1, ChunkMaxTokens: 3})
	if err != nil {
		t.Fatal(err)
	}
	second, err := splitChunks(documents[1], SplitConfig{ChapterMaxTokens: 1, ChunkMaxTokens: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 3 || len(second) != 1 {
		t.Fatalf("unexpected test chunks: %+v, %+v", first, second)
	}
	// 故意打乱索引顺序，结果仍须按候选排名输出、按文档序号寻找上下文。
	chunks := []Chunk{first[2], second[0], first[1], first[0]}
	candidates := []Candidate{{ChunkID: second[0].ID, Score: 0.9}, {ChunkID: first[1].ID, Score: 0.8}, {ChunkID: first[2].ID, Score: 0.7}}
	hits, err := buildRetrieveHits(context.Background(), documents, chunks, candidates)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != len(candidates) {
		t.Fatalf("unexpected hit count: %d", len(hits))
	}
	for index, hit := range hits {
		if hit.Chunk.ID != candidates[index].ChunkID || hit.Score != candidates[index].Score {
			t.Fatalf("candidate rank or score changed: %+v", hit)
		}
	}
	if hits[0].DocumentID != "b" || hits[0].DocumentTitle != "Second" || hits[0].PreviousChunk != nil || hits[0].NextChunk != nil || hits[0].PreviousChapter != nil || hits[0].NextChapter != nil {
		t.Fatalf("context crossed document boundary: %+v", hits[0])
	}
	middle := hits[1]
	if middle.DocumentFileName != "a.md" || middle.ProjectName != "Project" || middle.ProjectID != "project" || middle.Chunk.Content != first[1].Content {
		t.Fatalf("missing document metadata or exact evidence: %+v", middle)
	}
	if middle.Chapter.ID != "a:1" || middle.Chapter.Content != "# A\none two three four" || middle.Chapter.TokenCount != 5 || !reflect.DeepEqual(middle.Chapter.HeadingPath, []string{"A"}) {
		t.Fatalf("incorrect reconstructed chapter: %+v", middle.Chapter)
	}
	if middle.PreviousChunk == nil || middle.PreviousChunk.ID != first[0].ID || middle.NextChunk == nil || middle.NextChunk.ID != first[2].ID {
		t.Fatalf("incorrect adjacent chunks: %+v", middle)
	}
	if middle.PreviousChapter != nil || middle.NextChapter == nil || middle.NextChapter.Content != "# B\nfive six" {
		t.Fatalf("incorrect adjacent chapters: %+v", middle)
	}
	last := hits[2]
	if last.NextChunk != nil || last.NextChapter != nil || last.PreviousChapter == nil || last.PreviousChapter.ID != "a:1" {
		t.Fatalf("incorrect end-of-document context: %+v", last)
	}
	// 修改结果不能改变保存于输入索引中的切片。
	hits[1].Chunk.HeadingPath[0] = "Changed"
	if first[1].HeadingPath[0] != "A" || chunks[0].ID != first[2].ID {
		t.Fatal("input metadata or chunk ordering was modified")
	}
}

func TestBuildRetrieveHitsInvalidReferences(t *testing.T) {
	for _, tt := range []struct {
		name   string
		chunks []Chunk
		id     string
	}{
		{"missing chunk", []Chunk{{ID: "a", DocumentID: "doc", ChapterID: "doc:1"}}, "missing"},
		{"out of scope", []Chunk{{ID: "a", DocumentID: "other", ChapterID: "other:1"}}, "a"},
		{"duplicate id", []Chunk{{ID: "a", DocumentID: "doc"}, {ID: "a", DocumentID: "doc", Position: 1}}, "a"},
		{"duplicate position", []Chunk{{ID: "a", DocumentID: "doc", ChapterID: "doc:1"}, {ID: "b", DocumentID: "doc", ChapterID: "doc:2"}}, "a"},
		{"noncontiguous chapter", []Chunk{{ID: "a", DocumentID: "doc", ChapterID: "doc:1"}, {ID: "b", DocumentID: "doc", ChapterID: "doc:2", Position: 1}, {ID: "c", DocumentID: "doc", ChapterID: "doc:1", Position: 2}}, "a"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			hits, err := buildRetrieveHits(context.Background(), []Document{{ID: "doc"}}, tt.chunks, []Candidate{{ChunkID: tt.id}})
			if err == nil || len(hits) != 0 {
				t.Fatalf("expected an error and no results: %+v, %v", hits, err)
			}
		})
	}
}

func TestBuildRetrieveHitsEmptyAndCanceled(t *testing.T) {
	hits, err := buildRetrieveHits(context.Background(), nil, nil, nil)
	if err != nil || len(hits) != 0 {
		t.Fatalf("expected empty result: %+v, %v", hits, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = buildRetrieveHits(ctx, nil, nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
