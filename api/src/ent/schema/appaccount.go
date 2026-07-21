package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

type AppAccount struct {
	ent.Schema
}

func (AppAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "app_accounts"},
	}
}

func (AppAccount) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").Unique().Immutable(),
		field.String("email").Default(""),
		field.String("name").Default(""),
		field.String("picture").Default(""),
	}
}

func (AppAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
	}
}

func (AppAccount) Mixin() []ent.Mixin {
	return []ent.Mixin{
		timeMixin{},
	}
}

type timeMixin struct {
	mixin.Schema
}

func (timeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}
