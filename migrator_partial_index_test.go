package postgres

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

type partialUniqueIndexModel struct {
	ID        uint   `gorm:"primaryKey"`
	SsoDomain string `gorm:"uniqueIndex:account_sso_domain_unique,where:\"SsoDomain\" <> ''"`
}

type plainUniqueIndexModel struct {
	ID        uint   `gorm:"primaryKey"`
	SsoDomain string `gorm:"uniqueIndex:account_sso_domain"`
}

func parseSchema(t *testing.T, model interface{}) *schema.Schema {
	t.Helper()
	s, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{NoLowerCase: true})
	if err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	// ParseIndexes has the side effect of setting field.Unique for single-column
	// unique indexes; AutoMigrate runs it before MigrateColumn.
	_ = s.ParseIndexes()
	return s
}

func TestFieldHasPartialUniqueIndex_TrueForWhereTag(t *testing.T) {
	s := parseSchema(t, &partialUniqueIndexModel{})
	// ParseIndexes sets field.Unique for the single-column unique index; the
	// guard must still recognize it as partial so MigrateColumn defers to
	// CreateIndex instead of emitting a predicate-less UNIQUE constraint.
	f := s.FieldsByDBName["SsoDomain"]
	if !f.Unique {
		t.Fatal("expected field.Unique=true after ParseIndexes")
	}
	if !fieldHasPartialUniqueIndex(s, f) {
		t.Fatal("expected fieldHasPartialUniqueIndex=true for uniqueIndex with where:")
	}
}

func TestFieldHasPartialUniqueIndex_FalseForPlainUniqueIndex(t *testing.T) {
	s := parseSchema(t, &plainUniqueIndexModel{})
	f := s.FieldsByDBName["SsoDomain"]
	if !f.Unique {
		t.Fatal("expected field.Unique=true after ParseIndexes")
	}
	if fieldHasPartialUniqueIndex(s, f) {
		t.Fatal("expected fieldHasPartialUniqueIndex=false for uniqueIndex without where:")
	}
}
