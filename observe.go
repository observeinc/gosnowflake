package gosnowflake

import (
	"errors"
	"fmt"
	"io"

	"github.com/mailru/easyjson"
	_ "github.com/mailru/easyjson/gen" // This is required to have go mod vendor not remove the package
	"github.com/mailru/easyjson/jlexer"
)

const (
	// ResponseBodyLimit limits http response to be 100MB to avoid overwhelming the scheduler
	ResponseBodyLimit = 100 * 1024 * 1024
)

// ErrResponseTooLarge means the reponse is too large (thanks linter for these useful comments!)
var ErrResponseTooLarge = fmt.Errorf("response is too large")

func decodeResponse(body io.ReadCloser, resp interface{}) error {
	lr := &io.LimitedReader{R: body, N: ResponseBodyLimit}
	var err error
	if v, is := resp.(easyjson.Unmarshaler); is {
		err = easyjson.UnmarshalFromReader(lr, v)
		// jlexer reports a value cut short as io.EOF; encoding/json, which the
		// truncated-response retry in postRestfulQueryHelper expects, says io.ErrUnexpectedEOF.
		if err == io.EOF && lr.N < ResponseBodyLimit {
			err = io.ErrUnexpectedEOF
		}
	} else {
		err = decodeJSONBody(lr, resp)
	}
	var lexerErr *jlexer.LexerError
	// With budget left, io.ErrUnexpectedEOF is a truncated body, which
	// postRestfulQueryHelper retries; only a decode that hit the limit is too large.
	if err != nil && lr.N <= 0 &&
		(errors.Is(err, io.ErrUnexpectedEOF) || errors.As(err, &lexerErr)) {
		return ErrResponseTooLarge
	}
	return err
}
