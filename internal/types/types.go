package types

import (
	"encoding"
	"fmt"
	"math"
	"math/big"
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

var _ encoding.TextUnmarshaler = (*Numeric)(nil)

func (n *Numeric) UnmarshalText(text []byte) error {
	return n.Scan(string(text))
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

func (n Numeric) Float64() (float64, error) {
	f8, err := n.Float64Value()
	if err != nil {
		return 0, fmt.Errorf("failed to convert numeric to float64: %w", err)
	}
	return f8.Float64, nil
}

func (n Numeric) Add(other Numeric) (Numeric, error) {
	if !n.Valid || !other.Valid {
		return Numeric{}, fmt.Errorf("cannot add invalid (NULL) numeric values")
	}

	aInt := new(big.Int).Set(n.Int)
	bInt := new(big.Int).Set(other.Int)
	aExp := n.Exp
	bExp := other.Exp

	resultExp := min(bExp, aExp)

	if aExp > resultExp {
		diff := aExp - resultExp
		aInt.Mul(aInt, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(diff)), nil))
	}
	if bExp > resultExp {
		diff := bExp - resultExp
		bInt.Mul(bInt, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(diff)), nil))
	}

	resultInt := new(big.Int).Add(aInt, bInt)

	return Numeric{
		Numeric: pgtype.Numeric{
			Int:   resultInt,
			Exp:   resultExp,
			Valid: true,
		},
	}, nil
}

func (n Numeric) Sub(other Numeric) (Numeric, error) {
	if !n.Valid || !other.Valid {
		return Numeric{}, fmt.Errorf("cannot subtract invalid (NULL) numeric values")
	}

	aInt := new(big.Int).Set(n.Int)
	bInt := new(big.Int).Set(other.Int)
	aExp := n.Exp
	bExp := other.Exp

	resultExp := min(bExp, aExp)

	if aExp > resultExp {
		diff := aExp - resultExp
		aInt.Mul(aInt, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(diff)), nil))
	}
	if bExp > resultExp {
		diff := bExp - resultExp
		bInt.Mul(bInt, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(diff)), nil))
	}
	resultInt := new(big.Int).Sub(aInt, bInt)
	return Numeric{
		Numeric: pgtype.Numeric{
			Int:   resultInt,
			Exp:   resultExp,
			Valid: true,
		},
	}, nil
}

func (n Numeric) Mul(other Numeric) (Numeric, error) {
	if !n.Valid || !other.Valid {
		return Numeric{}, fmt.Errorf("cannot multiply invalid (NULL) numeric values")
	}

	resultInt := new(big.Int).Mul(n.Int, other.Int)
	resultExp := n.Exp + other.Exp

	return Numeric{
		Numeric: pgtype.Numeric{
			Int:   resultInt,
			Exp:   resultExp,
			Valid: true,
		},
	}, nil
}

func (n Numeric) Div(other Numeric, precision int32) (Numeric, error) {
	if !n.Valid || !other.Valid {
		return Numeric{}, fmt.Errorf("cannot divide invalid (NULL) numeric values")
	}
	if other.Int.Cmp(big.NewInt(0)) == 0 {
		return Numeric{}, fmt.Errorf("division by zero")
	}

	aRat := new(big.Rat).SetInt(n.Int)
	bRat := new(big.Rat).SetInt(other.Int)

	aExp := n.Exp
	bExp := other.Exp
	expDiff := aExp - bExp
	if expDiff != 0 {
		scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(math.Abs(float64(expDiff)))), nil)
		if expDiff > 0 {
			aRat.Mul(aRat, new(big.Rat).SetInt(scale))
		} else {
			bRat.Mul(bRat, new(big.Rat).SetInt(scale))
		}
	}

	resultRat := new(big.Rat).Quo(aRat, bRat)

	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision)), nil)
	resultRat.Mul(resultRat, new(big.Rat).SetInt(scale))
	resultInt := new(big.Int).Quo(resultRat.Num(), resultRat.Denom())

	return Numeric{
		Numeric: pgtype.Numeric{
			Int:   resultInt,
			Exp:   -precision,
			Valid: true,
		},
	}, nil
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
