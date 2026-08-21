package postgres

import (
	"regexp"
	"testing"
)

// stripDefaultCast mirrors the migrator.go line that strips a `::type` cast
// (and surrounding quotes) from a postgres column default so it can be
// compared to the gorm field default. gorm trims quotes off string defaults
// (gorm/schema/field.go), so the strip must return the *unquoted* inner value.
func stripDefaultCast(re *regexp.Regexp, s string) string {
	return re.ReplaceAllString(s, "$1")
}

// TestStripDefaultCast guards the default-value cast stripping in
// migrator.go. The previous regex used a word boundary before the optional
// closing quote, which only matched when the captured value ended in a word
// character. Defaults whose inner value did not end in a word char (an
// empty-string text default, or a jsonb default of {}) kept the cast and
// never compared equal to the gorm field default, so AutoMigrate re-emitted
// the same no-op ALTER on every run.
func TestStripDefaultCast(t *testing.T) {
	// input (postgres column default) -> expected (unquoted inner value)
	cases := map[string]string{
		`'ACTIVE'::text`:     `ACTIVE`,
		`''::text`:           ``, // regression: empty-string default
		`'default'::text`:    `default`,
		`'{}'::jsonb`:        `{}`, // regression: non-word-ending inner value
		`5::int`:             `5`,
		`5`:                  `5`, // no cast -> unchanged
		`false`:              `false`,
		`0`:                  `0`,
		`true::bool`:         `true`,
		`'2020-01-01'::date`: `2020-01-01`,
	}

	re := regexp.MustCompile(`'?(.*?)'?:+[\w\s]+$`)
	for in, want := range cases {
		if got := stripDefaultCast(re, in); got != want {
			t.Errorf("strip(%q) = %q, want %q", in, got, want)
		}
	}
}
