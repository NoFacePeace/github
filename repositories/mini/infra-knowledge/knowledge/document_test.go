package knowledge

import "testing"

func TestNewDocument(t *testing.T) {
	content := "\n# Guide\n\n    example\n"
	doc, err := NewDocument(" doc-1 ", " guide.md ", " Guide ", content, " project-1 ", " Project ")
	if err != nil {
		t.Fatal(err)
	}
	if doc.ID != "doc-1" || doc.FileName != "guide.md" || doc.Title != "Guide" || doc.ProjectID != "project-1" || doc.ProjectName != "Project" {
		t.Fatalf("unexpected document fields: %+v", doc)
	}
	if doc.Content != content {
		t.Fatal("document content was modified")
	}
	if len(doc.Fingerprint) != 64 {
		t.Fatalf("unexpected fingerprint: %q", doc.Fingerprint)
	}
}

func TestNewDocumentRequiresFields(t *testing.T) {
	for i, name := range []string{"id", "fileName", "title", "content", "projectID", "projectName"} {
		t.Run(name, func(t *testing.T) {
			fields := []string{"doc-1", "guide.md", "Guide", "body", "project-1", "Project"}
			fields[i] = " \n\t"
			doc, err := NewDocument(fields[0], fields[1], fields[2], fields[3], fields[4], fields[5])
			if err == nil || doc != (Document{}) {
				t.Fatalf("expected an error and empty document, got %+v, %v", doc, err)
			}
		})
	}
}

func TestDocumentFingerprint(t *testing.T) {
	base, err := NewDocument("doc-1", "guide.md", "Guide", "body", "project-1", "Project")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name   string
		fields [6]string
		same   bool
	}{
		{"new identity", [6]string{"doc-2", "guide.md", "Guide", "body", "project-1", "Renamed"}, true},
		{"metadata whitespace", [6]string{"doc-1", " guide.md ", " Guide ", "body", " project-1 ", "Project"}, true},
		{"file name", [6]string{"doc-1", "other.md", "Guide", "body", "project-1", "Project"}, false},
		{"title", [6]string{"doc-1", "guide.md", "Other", "body", "project-1", "Project"}, false},
		{"content", [6]string{"doc-1", "guide.md", "Guide", "body\n", "project-1", "Project"}, false},
		{"project", [6]string{"doc-1", "guide.md", "Guide", "body", "project-2", "Project"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.fields
			doc, err := NewDocument(f[0], f[1], f[2], f[3], f[4], f[5])
			if err != nil {
				t.Fatal(err)
			}
			if (doc.Fingerprint == base.Fingerprint) != tt.same {
				t.Fatalf("unexpected fingerprint equality: want %v", tt.same)
			}
		})
	}
}
