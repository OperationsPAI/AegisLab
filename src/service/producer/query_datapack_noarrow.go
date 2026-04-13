//go:build !duckdb_arrow

package producer

import (
	"context"
	"fmt"
	"io"
)

// QueryDatapackFileContent requires the duckdb_arrow build tag because duckdb's Arrow API
// is compiled behind that tag in github.com/duckdb/duckdb-go/v2.
func QueryDatapackFileContent(ctx context.Context, datapackID int, filePath string) (string, int64, io.ReadCloser, error) {
	return "", 0, nil, fmt.Errorf("QueryDatapackFileContent requires building with -tags duckdb_arrow")
}
