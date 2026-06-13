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

// Load reads a recipe back by name. Adding a recipe is adding a record; there is no
// code path — Save/Load round-trips the same data Run consumes.
func Load(st store.Store, name string) (Recipe, bool) {
	rec, ok := st.Get("recipe/" + name)
	if !ok {
		return Recipe{}, false
	}
	var r Recipe
	if err := json.Unmarshal([]byte(rec.Body), &r); err != nil {
		return Recipe{}, false
	}
	return r, true
}

// List returns every saved recipe.
func List(st store.Store) []Recipe {
	var out []Recipe
	for _, rec := range st.List(store.OfKind(store.KRecipe)) {
		var r Recipe
		if err := json.Unmarshal([]byte(rec.Body), &r); err == nil {
			out = append(out, r)
		}
	}
	return out
}
