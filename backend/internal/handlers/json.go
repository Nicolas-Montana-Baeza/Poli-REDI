package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func decodeStrictJSON(body []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("el cuerpo debe contener un unico objeto JSON")
	}

	return nil
}
