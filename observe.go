package gosnowflake

import (
	"encoding/json"
	"fmt"
	"io"
)

const (
	// ResponseBodyLimit limits http response to be 100MB to avoid overwhelming the scheduler
	ResponseBodyLimit = 100 * 1024 * 1024
)

// ErrResponseTooLarge means the reponse is too large (thanks linter for these useful comments!)
var ErrResponseTooLarge = fmt.Errorf("response is too large")

type limitedJSONDecoder struct {
	decoder *json.Decoder
}

func (d *limitedJSONDecoder) Decode(v interface{}) error {
	err := d.decoder.Decode(v)
	if err == io.ErrUnexpectedEOF {
		return ErrResponseTooLarge
	}
	return err
}

func newLimitedJSONDecoder(buf io.ReadCloser) *limitedJSONDecoder {
	return &limitedJSONDecoder{
		decoder: json.NewDecoder(io.LimitReader(buf, ResponseBodyLimit)),
	}
}
