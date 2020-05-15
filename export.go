package gosnowflake

import (
	"context"
	"net/http"
)

// This file just exports types, functions and methods as needed

// Types
type ExecResponse = execResponse
type ExecResponseRowType = execResponseRowType
type ExecResponseChunk = execResponseChunk
type SnowflakeRows = snowflakeRows
type SnowflakeChunkDownloader = snowflakeChunkDownloader
type SnowflakeRestful = snowflakeRestful

// Methods

func (sr *snowflakeRows) GetExecResponse() *ExecResponse {
	return sr.execResp
}

func (sr *snowflakeResult) GetExecResponse() *ExecResponse {
	return sr.execResp
}

// Setter methods for unit testing

func (sr *snowflakeRows) SetExecResponse(er *ExecResponse) {
	sr.execResp = er
}

func (sr *snowflakeResult) SetExecResponse(er *ExecResponse) {
	sr.execResp = er
}

// Helpers

func NewSnowflakeRowsDownloader(ctx context.Context, rowType []ExecResponseRowType, rowSet [][]*string, totalRowCount int64, chunkMetas []ExecResponseChunk, chunkHeaders map[string]string) *SnowflakeRows {
	sc := &snowflakeConn{ // fake connection just to provide a .rest
		rest: &snowflakeRestful{
			RequestTimeout: defaultRequestTimeout,
			Client:         &http.Client{},
		},
	}
	scd := &snowflakeChunkDownloader{
		ctx:                ctx,
		sc:                 sc,
		CurrentChunk:       rowSet,
		ChunkMetas:         chunkMetas,
		Total:              totalRowCount,
		TotalRowIndex:      int64(-1),
		CellCount:          len(rowType),
		FuncDownload:       downloadChunk,
		FuncDownloadHelper: downloadChunkHelper,
		FuncGet:            getChunk,
		ChunkHeader:        chunkHeaders,
	}
	scd.start()
	return &SnowflakeRows{
		sc:              sc,
		RowType:         rowType,
		ChunkDownloader: scd,
	}
}
