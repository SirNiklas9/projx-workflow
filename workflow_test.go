package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	gctx "github.com/SirNiklas9/projx-context"
	core "github.com/SirNiklas9/projx-core"
	store "github.com/SirNiklas9/projx-store"
)

// fakeModel is a deterministic stand-in for the one non-deterministic step. It
// records what it was handed and returns a fixed proposal.
type fakeModel struct {
	reply      string
	gotPrompt  string
	gotMapLen  int
	gotFocused bool
}

func (m *fakeModel) Propose(prompt string, pk *gctx.Packet) (string, error) {
	m.gotPrompt = prompt
	m.gotMapLen = len(pk.Map)
	m.gotFocused = pk.Focus != nil
	return m.reply, nil
}

func project(t *testing.T) *core.Project {
	t.Helper()
	root := t.TempDir()
	full := filepath.Join(root, "app", "main.go")
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	src := "package app\n\nfunc Add(a, b int) int { return a + b }\n"
	if err := os.WriteFile(full, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	p, _, err := core.ParseDir(root)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunReview(t *testing.T) {
	p := project(t)
	st := store.NewMem()
	m := &fakeModel{reply: "looks fine"}
	r := NewReview("review", "review this", gctx.Gate{})

	res, err := Run(r, p, st, m)
	if err != nil {
		t.Fatal(err)
	}
	if res.Output != "looks fine" {
		t.Fatalf("want proposal returned, got %q", res.Output)
	}
	if m.gotPrompt != "review this" || m.gotMapLen == 0 {
		t.Fatalf("model got wrong packet: prompt=%q mapLen=%d", m.gotPrompt, m.gotMapLen)
	}
	if res.Filed != "" || res.Edited != "" {
		t.Fatalf("review must not write: filed=%q edited=%q", res.Filed, res.Edited)
	}
}

func TestRunCommitFilesHistory(t *testing.T) {
	p := project(t)
	st := store.NewMem()
	m := &fakeModel{reply: "added Add()"}
	r := NewCommit("commit", "summarize the change", gctx.Gate{})

	res, err := Run(r, p, st, m)
	if err != nil {
		t.Fatal(err)
	}
	if res.Filed == "" {
		t.Fatal("commit must file a record")
	}
	rec, ok := st.Get(res.Filed)
	if !ok {
		t.Fatalf("filed record %q not in store", res.Filed)
	}
	if rec.Kind != store.KHistory || rec.Scope != store.ScopeProject {
		t.Fatalf("commit should file project history, got kind=%v scope=%v", rec.Kind, rec.Scope)
	}
	if !strings.Contains(rec.Body, "added Add()") || !strings.Contains(rec.Body, "## commit") {
		t.Fatalf("record body should wrap proposal in fixed format: %q", rec.Body)
	}
}

func TestRunEditCodeSplices(t *testing.T) {
	p := project(t)
	st := store.NewMem()
	add, ok := p.SymbolByID("app/main.go::Add")
	if !ok {
		t.Fatalf("Add not found; symbols=%v", p.Symbols())
	}
	m := &fakeModel{reply: "func Add(a, b int) int { return a + b + 0 }"}
	r := NewEdit("tweak", "no-op tweak", add.ID, gctx.Gate{})

	res, err := Run(r, p, st, m)
	if err != nil {
		t.Fatal(err)
	}
	if res.Edited != add.ID {
		t.Fatalf("want edited %q, got %q", add.ID, res.Edited)
	}
	if !m.gotFocused {
		t.Fatal("edit-code must hand the model a focused packet")
	}
	onDisk, err := os.ReadFile(filepath.Join(p.Root, "app", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(onDisk), "a + b + 0") {
		t.Fatalf("splice did not land on disk: %q", onDisk)
	}
	if !strings.HasPrefix(string(onDisk), "package app") {
		t.Fatalf("splice was not lossless outside the span: %q", onDisk)
	}
}

func TestRecipeSaveLoadRoundTrip(t *testing.T) {
	st := store.NewMem()
	r := NewCommit("commit", "summarize", gctx.Gate{Deny: []string{"secret/"}})
	if err := Save(st, r); err != nil {
		t.Fatal(err)
	}
	got, ok := Load(st, "commit")
	if !ok {
		t.Fatal("recipe should load back")
	}
	if got.Name != r.Name || got.Action != r.Action || got.RecordKind != r.RecordKind {
		t.Fatalf("round-trip lost data: %+v vs %+v", got, r)
	}
	if len(got.Gate.Deny) != 1 || got.Gate.Deny[0] != "secret/" {
		t.Fatalf("gate dials lost in round-trip: %+v", got.Gate)
	}
	if all := List(st); len(all) != 1 {
		t.Fatalf("List should see 1 recipe, got %d", len(all))
	}
}
