//go:build !duckdb_arrow

package injection

import (
	"context"
	"fmt"
	"io"
)

func (s *Service) queryDatapackFileContent(ctx context.Context, id int, filePath string) (string, int64, io.ReadCloser, error) {
	_ = ctx
	_ = id
	_ = filePath
	return "", 0, nil, fmt.Errorf("QueryDatapackFileContent requires building with -tags duckdb_arrow")
}
