package types

import (
	"database/sql/driver"
	"fmt"
	"os"

	"github.com/AlphaByte02/FairSplit/internal/aes"
)

var cipher_suite = aes.New(os.Getenv("CRYPT_KEY"))

type EncryptedString string

// dbValue: current field's value in database
func (es *EncryptedString) Scan(dbValue any) error {
	if dbValue == nil {
		*es = ""
		return nil
	}

	switch value := dbValue.(type) {
	case []byte:
		decryptText, _ := cipher_suite.DecryptBase64(string(value))
		*es = EncryptedString(decryptText)
	case string:
		decryptText, _ := cipher_suite.DecryptBase64(value)
		*es = EncryptedString(decryptText)
	default:
		return fmt.Errorf("unsupported data %#v", dbValue)
	}
	return nil
}

func (es EncryptedString) Value() (driver.Value, error) {
	if es == "" {
		return nil, nil
	}

	encodeText, err := cipher_suite.EncryptBase64([]byte(es))
	if err != nil {
		return nil, err
	}

	return encodeText, nil
}
