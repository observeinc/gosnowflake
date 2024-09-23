package gosnowflake

// This file just exports types, functions and methods as needed

import (
	"context"
	"database/sql/driver"
	"time"

	"github.com/snowflakedb/gosnowflake/v2/internal/query"
)

// ExecResponse exports execResponse
type ExecResponse = execResponse

// ExecResponseRowType exports execResponseRowType
type ExecResponseRowType = query.ExecResponseRowType

// ExecResponseChunk exports execResponseChunk
type ExecResponseChunk = query.ExecResponseChunk

// RawSnowflakeRows exports the "raw" underlying snowflakeRows
type RawSnowflakeRows = snowflakeRows

// SnowflakeRestful exports snowflakeRestful
type SnowflakeRestful = snowflakeRestful

// SnowflakeValue exports snowflakeValue
type SnowflakeValue = snowflakeValue

// ChunkRowType exports chunkRowType
type ChunkRowType = chunkRowType

// SimpleTokenAccessor exports simpleTokenAccessor
type SimpleTokenAccessor = simpleTokenAccessor

// ArrowToValue exports arrowToValue
var ArrowToValue = arrowToValue

// GetExecResponse returns the ExecResponse
func (sr *snowflakeRows) GetExecResponse() *ExecResponse {
	return sr.execResp
}

// GetExecResponse returns the ExecResponse
func (sr *snowflakeResult) GetExecResponse() *ExecResponse {
	return sr.execResp
}

// Setter method for unit testing
func (sr *snowflakeRows) SetExecResponse(er *ExecResponse) {
	sr.execResp = er
}

// Setter method for unit testing
func (sr *snowflakeResult) SetExecResponse(er *ExecResponse) {
	sr.execResp = er
}

// StringToValue exports stringToValue
// Deprecated: please use StringToValueWithContext instead
func StringToValue(dest *driver.Value, srcColumnMeta ExecResponseRowType, srcValue *string, loc *time.Location) error {
	sp := newSyncParams(nil)
	return stringToValue(context.Background(), dest, srcColumnMeta, srcValue, loc, &sp)
}

// StringToValueWithContext exports stringToValue
func StringToValueWithContext(ctx context.Context, dest *driver.Value, srcColumnMeta ExecResponseRowType, srcValue *string, loc *time.Location, params map[string]*string) error {
	sp := newSyncParams(params)
	return stringToValue(ctx, dest, srcColumnMeta, srcValue, loc, &sp)
}

// ExecResponseRowTypeEquals exports execResponseRowTypeEquals
func ExecResponseRowTypeEquals(a, b ExecResponseRowType) bool {
	// Skip fieldMetadata
	return a.Name == b.Name &&
		a.ByteLength == b.ByteLength &&
		a.Length == b.Length &&
		a.Type == b.Type &&
		a.Precision == b.Precision &&
		a.Scale == b.Scale &&
		a.Nullable == b.Nullable
}
