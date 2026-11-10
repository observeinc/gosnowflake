package gosnowflake

import (
	"bytes"
	"context"
	"database/sql/driver"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/snowflakedb/gosnowflake/v2/internal/query"
)

type repeatByteReader byte

func (b repeatByteReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(b)
	}
	return len(p), nil
}

// oversizedBody is a JSON document one value longer than ResponseBodyLimit, streamed so the
// test does not hold it in memory twice.
func oversizedBody() io.ReadCloser {
	return io.NopCloser(io.MultiReader(
		strings.NewReader(`{"data":{"rowset":[["`),
		io.LimitReader(repeatByteReader('x'), ResponseBodyLimit),
		strings.NewReader(`"]]},"code":"0","success":true}`),
	))
}

func body(s string) io.ReadCloser {
	return io.NopCloser(bytes.NewReader([]byte(s)))
}

func strPtr(s string) *string {
	return &s
}

func TestDecodeResponse(t *testing.T) {
	// execResponse decodes through easyjson, cancelQueryResponse through encoding/json.
	targets := map[string]func() any{
		"easyjson":      func() any { return &execResponse{} },
		"encoding/json": func() any { return &cancelQueryResponse{} },
	}
	for name, target := range targets {
		t.Run(name, func(t *testing.T) {
			assertErrIsE(t, decodeResponse(oversizedBody(), target()), ErrResponseTooLarge)
			// A body cut short within the limit must stay io.ErrUnexpectedEOF, which
			// postRestfulQueryHelper answers by retrying the query.
			assertErrIsE(t, decodeResponse(body(`{"data":{},"code":"0"`), target()), io.ErrUnexpectedEOF)
			assertNilE(t, decodeResponse(body(`{"data":{},"code":"0","success":true}`), target()))
		})
	}
}

func queryReturning(resp *execResponse) *snowflakeConn {
	return &snowflakeConn{
		cfg: &Config{},
		rest: &snowflakeRestful{
			FuncPostQuery: func(context.Context, *snowflakeRestful, *url.Values, map[string]string, []byte, time.Duration, UUID, *Config) (*execResponse, error) {
				return resp, nil
			},
		},
	}
}

func TestNoDownloadQueryLeavesChunksToTheCaller(t *testing.T) {
	resp := &execResponse{Success: true, Data: execResponseData{
		QueryID:           "01-nodownload",
		QueryResultFormat: "json",
		RowType:           []query.ExecResponseRowType{{Name: "C1", Type: "text"}},
		// Downloading this chunk would fail; the query must not try.
		Chunks: []query.ExecResponseChunk{{URL: "https://chunk.invalid/0", RowCount: 1}},
	}}
	rows, err := queryReturning(resp).QueryContext(context.Background(), "/*NODOWNLOAD*/ SELECT 1", nil)
	assertNilF(t, err)
	sfRows, ok := rows.(*snowflakeRows)
	assertTrueF(t, ok)
	assertNilE(t, sfRows.ChunkDownloader, "a NODOWNLOAD query must not start a chunk downloader")
	assertEqualE(t, sfRows.GetExecResponse(), resp)
	assertEqualE(t, sfRows.GetQueryID(), "01-nodownload")
	assertNilE(t, rows.Close())
}

func TestExecResponseIsExported(t *testing.T) {
	resp := &execResponse{Success: true, Data: execResponseData{
		QueryID:         "01-dml",
		StatementTypeID: statementTypeIDDml,
		RowType:         []query.ExecResponseRowType{{Name: "number of rows inserted", Type: "fixed"}},
		RowSet:          [][]*string{{strPtr("3")}},
	}}
	res, err := queryReturning(resp).ExecContext(context.Background(), "INSERT INTO t VALUES (1), (2), (3)", nil)
	assertNilF(t, err)
	sfRes, ok := res.(*snowflakeResult)
	assertTrueF(t, ok)
	assertEqualE(t, sfRes.GetExecResponse(), resp)
	n, err := res.RowsAffected()
	assertNilF(t, err)
	assertEqualE(t, n, int64(3))
}

func TestStringToValueExports(t *testing.T) {
	// A DATE arrives as days since the epoch.
	col := ExecResponseRowType{Name: "C1", Type: "date"}
	var dest driver.Value
	assertNilF(t, StringToValue(&dest, col, strPtr("19000"), time.UTC))
	assertEqualE(t, dest, time.Date(2022, time.January, 8, 0, 0, 0, 0, time.UTC))

	dest = nil
	assertNilF(t, StringToValueWithContext(context.Background(), &dest, ExecResponseRowType{Name: "C2", Type: "text"}, strPtr("abc"), time.UTC, nil))
	assertEqualE(t, dest, "abc")
}

func TestExecResponseRowTypeEqualsIgnoresFields(t *testing.T) {
	a := ExecResponseRowType{Name: "C1", Type: "object", Nullable: true, Fields: []query.FieldMetadata{{Name: "x", Type: "text"}}}
	b := a
	b.Fields = nil
	assertTrueE(t, ExecResponseRowTypeEquals(a, b))
	b.Name = "C2"
	assertFalseE(t, ExecResponseRowTypeEquals(a, b))
}
