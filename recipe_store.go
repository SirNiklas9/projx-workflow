package workflow

import (
	"encoding/json"
	"fmt"

	store "github.com/SirNiklas9/projx-store"
)

// Save persists a recipe as a KRecipe record (global scope — recipes are how YOU
// work, portable across projects). The recipe is its own data: JSON in the body.
func Save(st store.Store, r Recipe) error {
	body, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("workflow: marshal recipe %q: %w", r.Name, err)
	}
	return st.Put(store.Record{
		ID:    "recipe/" + r.Name,
		Kind:  store.KRecipe,
		Scope: store.ScopeGlobal,
		Key:   r.Name,
		Body:  string(body),
	})
}

// Load reads a recipe back by name. Returns (zero, false, nil) when not found and
// (zero, false, err) when the record exists but cannot be unmarshalled — callers
// can distinguish corruption from absence via the error.
func Load(st store.Store, name string) (Recipe, bool, error) {
	rec, ok := st.Get("recipe/" + name)
	if !ok {
		return Recipe{}, false, nil
	}
	var r Recipe
	if err := json.Unmarshal([]byte(rec.Body), &r); err != nil {
		return Recipe{}, false, fmt.Errorf("workflow: unmarshal recipe %q: %w", name, err)
	}
	return r, true, nil
}

// List returns every saved recipe. It returns an error on the first corrupt record
// so callers are not silently handed a truncated list.
func List(st store.Store) ([]Recipe, error) {
	var out []Recipe
	for _, rec := range st.List(store.OfKind(store.KRecipe)) {
		var r Recipe
		if err := json.Unmarshal([]byte(rec.Body), &r); err != nil {
			return out, fmt.Errorf("workflow: unmarshal recipe %q: %w", rec.Key, err)
		}
		out = append(out, r)
	}
	return out, nil
}
