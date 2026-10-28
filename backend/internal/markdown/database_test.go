package markdown

import "testing"

func TestExtractDatabaseQueryTables(t *testing.T) {
	t.Parallel()
	content := "# Interviews\n\n" +
		"{database-query table=interviews sort=year}\n\n" +
		"{database-query filter=\"status=open\" table=\"open tickets\"}\n\n" +
		"{database-query table=interviews}\n\n" +
		"```\n{database-query table=inside_code}\n```\n" +
		"{database-row table=not_a_query}\n"
	got := ExtractDatabaseQueryTables(content)
	want := []string{"interviews", "open tickets"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("ExtractDatabaseQueryTables = %q, want %q", got, want)
	}
}
