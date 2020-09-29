package schema

import (
	"github.com/facebook/ent"
	"github.com/facebook/ent/schema/edge"
	"github.com/facebook/ent/schema/field"
)

// Access holds the schema definition for the Access entity.
type Access struct {
	ent.Schema
}

func (Access) Config() ent.Config {
	return ent.Config{
		Table: "accesses",
	}
}

// Fields of the Access.
func (Access) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().Unique(),
		field.String("creator").NotEmpty(),
		field.Int64("created_at"),
	}
}

// Edges of the Access.
func (Access) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).Ref("accesses").Unique(),
	}
}
