package knowledge

import (
	"bytes"
	"context"
	"errors"
	"math"
	"testing"

	"github.com/coder/hnsw"
)

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
