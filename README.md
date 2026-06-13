# projx-workflow

The **recipe engine** — the optional layer on top of context + generation + store
that runs named jobs. A **recipe is data, not code**: saved dials (gate/focus) + a
prompt + an action + a store target. Adding your own workflow = adding a record.

A run is **deterministic except for one step** — the AI call — which is injected as
the `Model` interface, so the AI code stays in `generation`/external and this repo
never imports a model SDK. Everything around the model (gather the packet, file the
result, splice the edit) is plain code.

## Shape

- **`Recipe`** — pure data, JSON-serializable, storable as a `KRecipe` record.
  Constructors: `NewReview`, `NewCommit`, `NewEdit`.
- **`Model.Propose(prompt, *context.Packet) (string, error)`** — the single
  non-deterministic step, injected.
- **`Run(r, *core.Project, store.Store, Model) (*Result, error)`** — gather (context)
  → propose (model) → apply the action.
- **Actions:** `Review` (return text) · `FileRecord` (file the proposal into the
  store in a fixed format — the model never edits the record) · `EditCode` (splice
  the proposal into the focused symbol via generation).
- **`Save`/`Load`/`List`** — recipes persist as records; "adding a recipe is adding
  a record."

## Status

P6. Pure-Go. Reads context/core/generation/store via local `replace`. No model SDK.

```sh
go test ./...
```
