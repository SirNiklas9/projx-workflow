// Package workflow is ProjX's recipe engine — the optional layer that sits ON TOP
// of context + generation + store and runs named jobs. A recipe is DATA, not code:
// saved dials (gate/focus) + a prompt + an action + which store record it files.
// Adding your own workflow = adding a record, never writing code.
//
// A run is deterministic except for ONE step — the AI call — which is injected as
// the Model interface, so the AI code stays in generation/external and this repo
// stays pure, testable, and reproducible. Everything around the model (gather the
// packet, file the result, splice the edit) is plain code.
package workflow

import (
	gctx "github.com/SirNiklas9/projx-context"
	store "github.com/SirNiklas9/projx-store"
)

// Action is what a recipe DOES with the model's proposal.
type Action int

const (
	// Review is read-only: the proposal is returned as text (review, explain).
	Review Action = iota
	// FileRecord writes: the proposal is filed into the store in a fixed format
	// (commit/history/ADR). The model never edits the record directly.
	FileRecord
	// EditCode writes: the proposal replaces the focused symbol's span via the
	// deterministic splice (generation). The surgical edit as a saved job.
	EditCode
)

var actionNames = map[Action]string{
	Review: "review", FileRecord: "file-record", EditCode: "edit-code",
}

func (a Action) String() string {
	if n, ok := actionNames[a]; ok {
		return n
	}
	return "action?"
}

// Recipe is a named job — pure data, JSON-serializable, storable as a KRecipe
// record. It carries the dials it applies to context, the prompt it gives the
// model, and (for FileRecord) the store target it files into.
type Recipe struct {
	Name   string    `json:"name"`
	Action Action    `json:"action"`
	Prompt string    `json:"prompt"`
	Gate   gctx.Gate `json:"gate"`
	// Focus is a symbol's stable ID. Required for EditCode (the span to replace);
	// an optional deep-context anchor for Review/FileRecord.
	Focus string `json:"focus,omitempty"`
	// RecordKind/RecordScope are the store target for FileRecord recipes.
	RecordKind  store.Kind  `json:"recordKind,omitempty"`
	RecordScope store.Scope `json:"recordScope,omitempty"`
}

// Review builds a read-only recipe.
func NewReview(name, prompt string, gate gctx.Gate) Recipe {
	return Recipe{Name: name, Action: Review, Prompt: prompt, Gate: gate}
}

// NewCommit builds a write recipe that files the proposal as project history.
func NewCommit(name, prompt string, gate gctx.Gate) Recipe {
	return Recipe{
		Name: name, Action: FileRecord, Prompt: prompt, Gate: gate,
		RecordKind: store.KHistory, RecordScope: store.ScopeProject,
	}
}

// NewEdit builds a write recipe that splices the proposal into a focused symbol.
func NewEdit(name, prompt, focusID string, gate gctx.Gate) Recipe {
	return Recipe{Name: name, Action: EditCode, Prompt: prompt, Gate: gate, Focus: focusID}
}
