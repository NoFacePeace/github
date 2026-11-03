package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/coder/hnsw"
)

func TestBuildFulltextIndex(t *testing.T) {
	chunks := []Chunk{
		{ID: "a", DocumentID: "doc", ChapterID: "doc:1", Position: 0, Content: "Redis redis 配置 max_memory=512", HeadingPath: []string{"Guide"}, SectionPaths: []string{"1"}},
		{ID: "b", DocumentID: "doc", Position: 1, Content: "REDIS cache"},
		{ID: "c", DocumentID: "doc", Position: 2, Content: "... !!!"},
	}
	index, err := (&indexer{}).buildFulltextIndex(context.Background(), chunks)
	if err != nil {
		t.Fatal(err)
	}
	wantPostings := map[string]map[string]int{
		"redis": {"a": 2, "b": 1}, "配": {"a": 1}, "置": {"a": 1},
		"max": {"a": 1}, "memory": {"a": 1}, "512": {"a": 1}, "cache": {"b": 1},
	}
	if !reflect.DeepEqual(index.Postings, wantPostings) {
		t.Fatalf("unexpected postings: %+v", index.Postings)
	}
	wantDocFreq := map[string]int{"redis": 2, "配": 1, "置": 1, "max": 1, "memory": 1, "512": 1, "cache": 1}
	if !reflect.DeepEqual(index.DocFreq, wantDocFreq) || index.AvgLen != 3 {
		t.Fatalf("unexpected statistics: %+v, %v", index.DocFreq, index.AvgLen)
	}
	for position, count := range []int{7, 2, 0} {
		if index.Chunks[position].Content != chunks[position].Content || index.Chunks[position].Position != position || index.Chunks[position].TokenCount != count {
			t.Fatalf("chunk content, order, or length changed: %+v", index.Chunks[position])
		}
	}
	if chunks[0].TokenCount != 0 || index.Chunks[0].ChapterID != "doc:1" {
		t.Fatal("input was mutated or chapter metadata was lost")
	}
	chunks[0].HeadingPath[0] = "Changed"
	chunks[0].SectionPaths[0] = "changed"
	chunks[0].Content = "changed"
	if index.Chunks[0].HeadingPath[0] != "Guide" || index.Chunks[0].SectionPaths[0] != "1" || index.Chunks[0].Content == "changed" {
		t.Fatal("index shares mutable data with the input")
	}
	encoded, err := json.Marshal(index)
	if err != nil {
		t.Fatal(err)
	}
	var restored FulltextIndex
	if err := json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(index, restored) {
		t.Fatal("index did not survive JSON round trip")
	}
}

func TestBuildFulltextIndexUnicodeTerms(t *testing.T) {
	index, err := (&indexer{}).buildFulltextIndex(context.Background(), []Chunk{{ID: "a", Content: "CAFÉ café 中文 foo_bar 123"}})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]map[string]int{"café": {"a": 2}, "中": {"a": 1}, "文": {"a": 1}, "foo": {"a": 1}, "bar": {"a": 1}, "123": {"a": 1}}
	if !reflect.DeepEqual(index.Postings, want) || index.Chunks[0].TokenCount != 7 {
		t.Fatalf("unexpected Unicode terms: %+v", index)
	}
}

func TestBuildFulltextIndexEmpty(t *testing.T) {
	index, err := (&indexer{}).buildFulltextIndex(context.Background(), nil)
	if err != nil || len(index.Chunks) != 0 || len(index.Postings) != 0 || len(index.DocFreq) != 0 || index.AvgLen != 0 {
		t.Fatalf("expected empty index, got %+v, %v", index, err)
	}
}

func TestBuildFulltextIndexCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (&indexer{}).buildFulltextIndex(ctx, []Chunk{{ID: "a", Content: "body"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

func TestBuildFulltextIndexInvalidChunkIDs(t *testing.T) {
	for _, chunks := range [][]Chunk{
		{{Content: "body"}},
		{{ID: "a", Content: "first"}, {ID: "a", Content: "second"}},
	} {
		index, err := (&indexer{}).buildFulltextIndex(context.Background(), chunks)
		if err == nil || len(index.Chunks) != 0 || len(index.Postings) != 0 {
			t.Fatalf("expected error and no partial index, got %+v, %v", index, err)
		}
	}
}

type recordingEmbedder struct {
	input   string
	err     error
	vectors [][]float32
	calls   int
}

func (e *recordingEmbedder) Embed(_ context.Context, content string) ([]float32, error) {
	e.input = content
	if e.err != nil {
		return nil, e.err
	}
	vector := e.vectors[e.calls]
	e.calls++
	return vector, nil
}

func TestBuildVectorIndexHeadingInput(t *testing.T) {
	for _, tt := range []struct {
		name     string
		headings []string
		want     string
	}{
		{"with headings", []string{"Root", "Child"}, "Root > Child\nbody"},
		{"without headings", nil, "body"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			failure := errors.New("embedding unavailable")
			embedder := &recordingEmbedder{err: failure}
			builder := &indexer{embedder: embedder}
			_, err := builder.buildVectorIndex(context.Background(), []Chunk{{ID: "chunk-1", Content: "body", HeadingPath: tt.headings}})
			if !errors.Is(err, failure) {
				t.Fatalf("expected embedding error, got %v", err)
			}
			if embedder.input != tt.want {
				t.Fatalf("embedding input: got %q, want %q", embedder.input, tt.want)
			}
		})
	}
}

func TestBuildVectorIndexRoundTrip(t *testing.T) {
	embedder := &recordingEmbedder{vectors: [][]float32{{1, 0}, {0, 1}, {-1, 0}}}
	builder := &indexer{embedder: embedder}
	index, err := builder.buildVectorIndex(context.Background(), []Chunk{
		{ID: "a", Content: "first"},
		{ID: "b", Content: "second"},
		{ID: "c", Content: "third"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if index.Dimensions != 2 || embedder.calls != 3 {
		t.Fatalf("unexpected dimensions or calls: %d, %d", index.Dimensions, embedder.calls)
	}
	graph := hnsw.NewGraph[string]()
	if err := graph.Import(bytes.NewReader(index.Graph)); err != nil {
		t.Fatal(err)
	}
	defaults := hnsw.NewGraph[string]()
	if graph.Len() != 3 || graph.M != defaults.M || graph.Ml != defaults.Ml || graph.EfSearch != defaults.EfSearch {
		t.Fatal("graph contents or library defaults did not survive export")
	}
	hits := graph.Search([]float32{0, 1}, 1)
	if len(hits) != 1 || hits[0].Key != "b" {
		t.Fatalf("unexpected nearest neighbor: %+v", hits)
	}
}

func TestBuildVectorIndexInvalidVectors(t *testing.T) {
	for _, tt := range []struct {
		name    string
		vectors [][]float32
	}{
		{"empty", [][]float32{nil}},
		{"zero", [][]float32{{0, 0}}},
		{"nan", [][]float32{{float32(math.NaN()), 1}}},
		{"infinite", [][]float32{{1, float32(math.Inf(1))}}},
		{"dimensions", [][]float32{{1, 0}, {1, 0, 0}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			embedder := &recordingEmbedder{vectors: tt.vectors}
			chunks := []Chunk{{ID: "a", Content: "first"}}
			if len(tt.vectors) == 2 {
				chunks = append(chunks, Chunk{ID: "b", Content: "second"})
			}
			index, err := (&indexer{embedder: embedder}).buildVectorIndex(context.Background(), chunks)
			if err == nil || len(index.Graph) != 0 {
				t.Fatalf("expected error and no index, got %+v, %v", index, err)
			}
		})
	}
}

func TestBuildVectorIndexCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	embedder := &recordingEmbedder{}
	_, err := (&indexer{embedder: embedder}).buildVectorIndex(ctx, []Chunk{{ID: "a"}})
	if !errors.Is(err, context.Canceled) || embedder.calls != 0 {
		t.Fatalf("canceled build called embedder or returned wrong error: %v", err)
	}
}

func TestBuildVectorIndexEmpty(t *testing.T) {
	index, err := (&indexer{}).buildVectorIndex(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	graph := hnsw.NewGraph[string]()
	if err := graph.Import(bytes.NewReader(index.Graph)); err != nil {
		t.Fatal(err)
	}
	if graph.Len() != 0 || index.Dimensions != 0 {
		t.Fatal("expected empty graph")
	}
}
