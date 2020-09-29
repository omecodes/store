package schema

import (
	"github.com/facebook/ent"
	"github.com/facebook/ent/schema/edge"
	"github.com/facebook/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().NotEmpty(),
		field.String("password").NotEmpty(),
		field.String("email").Unique().NotEmpty(),
		field.Bool("validated").Default(false),
		field.Int64("created_at"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", Group.Type).Ref("users").Unique(),
		edge.To("accesses", Access.Type),
		edge.To("permissions", Permission.Type),
	}
}
