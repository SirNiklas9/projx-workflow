package workflow

import (
	"fmt"
	"path/filepath"

	gctx "github.com/SirNiklas9/projx-context"
	core "github.com/SirNiklas9/projx-core"
	generation "github.com/SirNiklas9/projx-generation"
	store "github.com/SirNiklas9/projx-store"
)

// Model is the ONLY non-deterministic step in a run: given the prepared prompt and
// the gated packet, it produces a proposal. It is injected so the AI lives in
// generation/external and this repo never imports a model SDK.
type Model interface {
	Propose(prompt string, packet *gctx.Packet) (string, error)
}

// Result is what a run produced. Output is always the model's raw proposal; Filed
// and Edited record any write side effect.
type Result struct {
	Recipe string
	Action Action
	Output string
	Filed  string // store record ID, for FileRecord recipes
	Edited string // symbol ID spliced, for EditCode recipes
}

// Run executes a recipe deterministically around one model call: gather the gated
// packet (context) → propose (model) → apply the action (return / file / splice).
//
// CONTRACT — stale-span safety: p must reflect the current on-disk state of every
// file in the project. For EditCode recipes, Run locates the symbol's byte span in
// p and splices directly into the file at those offsets. If any file has been edited
// on disk since p was parsed, the spans are stale and the splice will corrupt it.
// Callers are responsible for re-parsing (core.ParseDir) when files may have changed
// before calling Run with an EditCode recipe.
func Run(r Recipe, p *core.Project, st store.Store, m Model) (*Result, error) {
	pk, err := gctx.Build(p, r.Focus, r.Gate)
	if err != nil {
		return nil, fmt.Errorf("workflow %q: gather: %w", r.Name, err)
	}
	out, err := m.Propose(r.Prompt, pk)
	if err != nil {
		return nil, fmt.Errorf("workflow %q: propose: %w", r.Name, err)
	}
	res := &Result{Recipe: r.Name, Action: r.Action, Output: out}

	switch r.Action {
	case Review:
		return res, nil

	case FileRecord:
		rec := store.Record{
			ID:    r.RecordScope.String() + "/" + r.RecordKind.String() + "/" + r.Name,
			Kind:  r.RecordKind,
			Scope: r.RecordScope,
			Key:   r.Name,
			Body:  fileFormat(r.Name, out),
		}
		if err := st.Put(rec); err != nil {
			return nil, fmt.Errorf("workflow %q: file: %w", r.Name, err)
		}
		res.Filed = rec.ID
		return res, nil

	case EditCode:
		if r.Focus == "" {
			return nil, fmt.Errorf("workflow %q: edit-code needs a focus symbol", r.Name)
		}
		sym, path, ok := locate(p, r.Focus)
		if !ok {
			return nil, fmt.Errorf("workflow %q: focus %q not found", r.Name, r.Focus)
		}
		full := filepath.Join(p.Root, filepath.FromSlash(path))
		if err := generation.ApplyFile(full, sym.Span, out); err != nil {
			return nil, fmt.Errorf("workflow %q: edit: %w", r.Name, err)
		}
		res.Edited = sym.ID
		return res, nil
	}
	return nil, fmt.Errorf("workflow %q: unknown action %d", r.Name, r.Action)
}

// fileFormat wraps a model proposal in a fixed, deterministic record body. The
// model proposes raw content; the recipe — not the model — controls the format.
func fileFormat(name, body string) string {
	return "## " + name + "\n\n" + body + "\n"
}

// locate finds the symbol with the given ID and the file path that holds it.
func locate(p *core.Project, id string) (core.Symbol, string, bool) {
	f, ok := p.SymbolFile(id)
	if !ok {
		return core.Symbol{}, "", false
	}
	s, _ := f.SymbolByID(id)
	return s, f.Path, true
}
