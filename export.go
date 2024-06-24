package gosnowflake

// This file just exports types, functions and methods as needed

import (
	"context"
	"database/sql/driver"
	"time"
)

// ExecResponse exports execResponse
type ExecResponse = execResponse

// ExecResponseRowType exports execResponseRowType
type ExecResponseRowType = execResponseRowType

// ExecResponseChunk exports execResponseChunk
type ExecResponseChunk = execResponseChunk

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
// Backwards compatible version. Deprecated.
func StringToValue(dest *driver.Value, srcColumnMeta execResponseRowType, srcValue *string, loc *time.Location) error {
	return stringToValue(context.Background(), dest, srcColumnMeta, srcValue, loc, nil)
}

// StringToValueCtx exports stringToValue
func StringToValueCtx(ctx context.Context, dest *driver.Value, srcColumnMeta execResponseRowType, srcValue *string, loc *time.Location, params map[string]*string) error {
	return stringToValue(ctx, dest, srcColumnMeta, srcValue, loc, params)
}
