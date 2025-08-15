package types

import (
	"encoding"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

/* func Text(v string) pgtype.Text {
	var t pgtype.Text
	_ = t.Scan(v)
	return t
}

func Numeric(v string) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(v) // v può essere int, float64, string o nil
	return n
} */

// Wrapper per pgtype.Numeric
type Numeric struct {
	pgtype.Numeric
}

var _ encoding.TextUnmarshaler = (*Numeric)(nil) // Verifica interfaccia

func (n *Numeric) UnmarshalText(text []byte) error {
	return n.Scan(string(text)) // Usa Scan per convertire la stringa in Numeric
}

func (n Numeric) String() string {
	if !n.Valid {
		return "" // O "NULL" se preferisci per valori non validi
	}

	// Usa big.Int.String() per la parte intera
	intStr := n.Int.String()
	if n.Exp == 0 {
		return intStr // Nessuna parte frazionaria
	}

	// Gestisci la parte frazionaria (Exp indica la posizione del punto decimale)
	if n.Exp > 0 {
		// Espandi con zeri (es. Exp=2 -> moltiplica per 10^2)
		return intStr + strings.Repeat("0", int(n.Exp))
	}

	// Exp < 0: inserisci il punto decimale
	exp := int(-n.Exp) // Numero di cifre decimali
	intStrLen := len(intStr)
	if intStr[0] == '-' {
		intStrLen-- // Escludi il segno negativo dalla lunghezza
	}

	// Padding con zeri a sinistra se necessario
	if intStrLen <= exp {
		prefix := "0."
		if intStr[0] == '-' {
			prefix = "-0."
			intStr = intStr[1:] // Rimuovi il segno negativo
		}
		return prefix + strings.Repeat("0", exp-intStrLen) + intStr
	}

	// Inserisci il punto decimale nella posizione corretta
	pos := intStrLen - exp
	if intStr[0] == '-' {
		return intStr[:pos+1] + "." + intStr[pos+1:] // +1 per il segno negativo
	}
	return intStr[:pos] + "." + intStr[pos:]
}

func (n Numeric) ToFloat64() (float64, error) {
	if !n.Valid {
		return 0, fmt.Errorf("cannot convert invalid (NULL) numeric to float64")
	}

	str := n.String()
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert numeric to float64: %w", err)
	}
	return f, nil
}

// Wrapper per pgtype.Text
type Text struct {
	pgtype.Text
}

var _ encoding.TextUnmarshaler = (*Text)(nil)
var _ fmt.Stringer = (*Text)(nil)

func (t *Text) UnmarshalText(text []byte) error {
	return t.Scan(string(text))
}

func (t *Text) String() string {
	v, _ := t.TextValue()
	return v.String
}
