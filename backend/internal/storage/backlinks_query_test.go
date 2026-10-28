package storage

import (
	"slices"
	"testing"
)

const rowPage = "{database-row table=interviews}\n\n| Field | Value |\n| --- | --- |\n| name | Adam |\n"

// A row-bound page is linked to, at render time, by every page that
// lists its table with {database-query}. Its backlinks include them.
func TestGetBacklinks_RowPageIncludesQueryPages(t *testing.T) {
	t.Parallel()
	s := newTestFileStore(t)
	mustPut := func(p, md string) {
		if _, err := s.Put(p, md, "alice"); err != nil {
			t.Fatalf("Put %s: %v", p, err)
		}
	}
	mustPut("/hr/interviews/adam-2026", rowPage)
	mustPut("/hr/dashboard", "# Dashboard\n\n{database-query table=interviews}\n")
	mustPut("/hr/other", "# Other\n\n{database-query table=tickets}\n")
	mustPut("/hr/linker", "See [Adam](/hr/interviews/adam-2026).\n")

	got := s.GetBacklinks("/hr/interviews/adam-2026")
	want := []string{"/hr/dashboard", "/hr/linker"}
	if !slices.Equal(got, want) {
		t.Errorf("backlinks = %v, want %v", got, want)
	}

	// Removing the query drops the backlink.
	mustPut("/hr/dashboard", "# Dashboard\n\nNo query any more.\n")
	if got := s.GetBacklinks("/hr/interviews/adam-2026"); !slices.Equal(got, []string{"/hr/linker"}) {
		t.Errorf("after removing the query, backlinks = %v, want [/hr/linker]", got)
	}
}

// A page that is not row-bound gets only hyperlink backlinks, even if
// some page queries a table.
func TestGetBacklinks_PlainPageIgnoresQueries(t *testing.T) {
	t.Parallel()
	s := newTestFileStore(t)
	if _, err := s.Put("/plain", "# Plain\n", "alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Put("/dash", "{database-query table=interviews}\n", "alice"); err != nil {
		t.Fatal(err)
	}
	if got := s.GetBacklinks("/plain"); len(got) != 0 {
		t.Errorf("backlinks = %v, want none", got)
	}
}

// The startup rebuild repopulates the query map.
func TestGetBacklinks_QueryPagesSurviveRebuild(t *testing.T) {
	t.Parallel()
	s := newTestFileStore(t)
	_, _ = s.Put("/hr/interviews/adam-2026", rowPage, "alice")
	_, _ = s.Put("/hr/dashboard", "{database-query table=interviews}\n", "alice")
	if err := s.RebuildIndexes(); err != nil {
		t.Fatalf("RebuildIndexes: %v", err)
	}
	if got := s.GetBacklinks("/hr/interviews/adam-2026"); !slices.Equal(got, []string{"/hr/dashboard"}) {
		t.Errorf("after rebuild, backlinks = %v, want [/hr/dashboard]", got)
	}
}
