//go:build duckdb_arrow

package injection

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/duckdb/duckdb-go/v2"
	"github.com/sirupsen/logrus"
)

func (s *Service) queryDatapackFileContent(ctx context.Context, id int, filePath string) (string, int64, io.ReadCloser, error) {
	injection, err := s.getReadyDatapack(id)
	if err != nil {
		return "", 0, nil, err
	}

	fullPath, err := s.store.ResolveFilePath(injection.Name, filePath)
	if err != nil {
		return "", 0, nil, fmt.Errorf("invalid file path: %w", err)
	}
	if filepath.Ext(fullPath) != ".parquet" {
		return "", 0, nil, fmt.Errorf("file is not a parquet file: %s", filePath)
	}

	connector, err := duckdb.NewConnector("", nil)
	if err != nil {
		return "", 0, nil, err
	}

	countConn, err := connector.Connect(ctx)
	if err != nil {
		logrus.Errorf("connect failed: %v", err)
		return "", 0, nil, err
	}
	defer func() { _ = countConn.Close() }()

	var totalRows int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM read_parquet('%s')", fullPath)

	db := sql.OpenDB(connector)
	if err := db.QueryRowContext(ctx, countQuery).Scan(&totalRows); err != nil {
		return "", 0, nil, err
	}

	safeSQL, err := buildSafeParquetSQL(ctx, db, fullPath)
	if err != nil {
		return "", 0, nil, fmt.Errorf("failed to build safe parquet SQL: %w", err)
	}

	pr, pw := io.Pipe()
	go func() {
		defer func() { _ = pw.Close() }()

		conn, err := connector.Connect(ctx)
		if err != nil {
			logrus.Errorf("connect failed: %v", err)
			return
		}
		defer func() { _ = conn.Close() }()

		arrow, err := duckdb.NewArrowFromConn(conn)
		if err != nil {
			_ = pw.CloseWithError(fmt.Errorf("failed to get arrow interface: %w", err))
			return
		}

		rdr, err := arrow.QueryContext(ctx, safeSQL)
		if err != nil {
			_ = pw.CloseWithError(fmt.Errorf("query context failed: %w", err))
			return
		}
		defer rdr.Release()

		writer := ipc.NewWriter(pw, ipc.WithSchema(rdr.Schema()), ipc.WithZstd())
		defer func() { _ = writer.Close() }()

		for rdr.Next() {
			record := rdr.RecordBatch()
			if err := writer.Write(record); err != nil {
				record.Release()
				_ = pw.CloseWithError(fmt.Errorf("failed to write arrow record: %w", err))
				return
			}
			record.Release()
		}

		if err := rdr.Err(); err != nil {
			_ = pw.CloseWithError(fmt.Errorf("reader error: %w", err))
		}
	}()

	return filepath.Base(fullPath), totalRows, pr, nil
}

func buildSafeParquetSQL(ctx context.Context, db *sql.DB, filePath string) (string, error) {
	fallbackSQL := fmt.Sprintf("SELECT * FROM read_parquet('%s')", filePath)
	describeQuery := fmt.Sprintf("DESCRIBE SELECT * FROM read_parquet('%s')", filePath)
	rows, err := db.QueryContext(ctx, describeQuery)
	if err != nil {
		logrus.Warnf("failed to describe parquet file, falling back to SELECT *: %v", err)
		return fallbackSQL, nil
	}
	defer func() { _ = rows.Close() }()

	var columns []string
	for rows.Next() {
		var colName, colType, null, key, def, extra string
		if err := rows.Scan(&colName, &colType, &null, &key, &def, &extra); err != nil {
			return "", err
		}

		quotedName := fmt.Sprintf("\"%s\"", strings.ReplaceAll(colName, "\"", "\"\""))
		normalized := strings.ToUpper(strings.TrimSpace(colType))

		switch normalized {
		case "UINT64", "UHUGEINT":
			logrus.Infof("parquet column %q: casting %s -> BIGINT", colName, colType)
			columns = append(columns, fmt.Sprintf("CAST(%s AS BIGINT) AS %s", quotedName, quotedName))
		default:
			columns = append(columns, quotedName)
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if len(columns) == 0 {
		return fallbackSQL, nil
	}

	return fmt.Sprintf("SELECT %s FROM read_parquet('%s')", strings.Join(columns, ", "), filePath), nil
}
