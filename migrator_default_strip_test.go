package postgres

import (
	"regexp"
	"testing"
)

// stripDefaultCast mirrors the migrator.go cast-stripping line. gorm trims
// quotes off string defaults, so this must return the unquoted inner value.
func stripDefaultCast(re *regexp.Regexp, s string) string {
	return re.ReplaceAllString(s, "$1")
}

// TestStripDefaultCast guards the default-value cast stripping in migrator.go.
func TestStripDefaultCast(t *testing.T) {
	// input (postgres column default) -> expected (unquoted inner value)
	cases := map[string]string{
		`'ACTIVE'::text`:     `ACTIVE`,
		`''::text`:           ``,
		`'default'::text`:    `default`,
		`'{}'::jsonb`:        `{}`,
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
