package types

import (
	"encoding"

	"github.com/jackc/pgx/v5/pgtype"
)

// Wrapper per pgtype.Text
type Text struct {
	pgtype.Text
}

var _ encoding.TextUnmarshaler = (*Text)(nil)

func (t *Text) UnmarshalText(text []byte) error {
	nt := string(text)

	if nt == "" {
		return t.Scan(nil)
	}
	return t.Scan(nt)
}

func NewText(v string) Text {
	var t Text
	_ = t.Scan(v)
	return t
}
