// replace-domains scans all string-like columns in the metargb MySQL database
// and replaces legacy RGB hostnames with MetaRang domains.
//
// Prefer Makefile targets (uses Docker MySQL, no Go deps):
//
//	make replace-domains      # dry-run
//	make replace-domains-run  # apply
//
// Or run this tool directly (connects via DB_* env, default localhost:3306):
//
//	go run . [--dry-run]
//
// Environment (defaults match docker-compose):
//
//	DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_DATABASE
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var replacements = []struct {
	old, new string
}{
	{"api.rgb.irpsc.com", "api.metarang.com"},
	{"admin.rgb.irpsc.com", "admin.metarang.com"},
}

var textDataTypes = map[string]bool{
	"char":       true,
	"varchar":    true,
	"tinytext":   true,
	"text":       true,
	"mediumtext": true,
	"longtext":   true,
	"json":       true,
	"enum":       true,
	"set":        true,
}

type columnRef struct {
	table  string
	column string
}

func main() {
	dryRun := flag.Bool("dry-run", false, "report matches without updating rows")
	flag.Parse()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&multiStatements=false",
		getEnv("DB_USER", "metargb_user"),
		getEnv("DB_PASSWORD", "metargb_password"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_DATABASE", "metargb_db"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	schema := getEnv("DB_DATABASE", "metargb_db")
	columns, err := listTextColumns(db, schema)
	if err != nil {
		log.Fatalf("list columns: %v", err)
	}

	if len(columns) == 0 {
		log.Printf("no string-like columns found in schema %q", schema)
		return
	}

	log.Printf("schema=%q columns=%d dry_run=%v", schema, len(columns), *dryRun)
	log.Printf("replacements: api.rgb.irpsc.com -> api.metarang.com, admin.rgb.irpsc.com -> admin.metarang.com")

	var totalMatches, totalUpdated int64

	for _, col := range columns {
		for _, rep := range replacements {
			matches, updated, err := processColumn(db, col, rep.old, rep.new, *dryRun)
			if err != nil {
				log.Fatalf("%s.%s (%s -> %s): %v", col.table, col.column, rep.old, rep.new, err)
			}
			if matches == 0 {
				continue
			}
			totalMatches += matches
			totalUpdated += updated
			action := "would update"
			if !*dryRun {
				action = "updated"
			}
			log.Printf("  %s.%s: %d row(s) match %q; %s %d row(s)",
				col.table, col.column, matches, rep.old, action, updated)
		}
	}

	if totalMatches == 0 {
		log.Printf("done: no occurrences found")
		return
	}
	if *dryRun {
		log.Printf("done (dry-run): %d row(s) would be updated across all columns", totalMatches)
		return
	}
	log.Printf("done: %d row(s) updated", totalUpdated)
}

func listTextColumns(db *sql.DB, schema string) ([]columnRef, error) {
	const q = `
		SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME, ORDINAL_POSITION`

	rows, err := db.Query(q, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []columnRef
	for rows.Next() {
		var table, column, dataType string
		if err := rows.Scan(&table, &column, &dataType); err != nil {
			return nil, err
		}
		if !textDataTypes[strings.ToLower(dataType)] {
			continue
		}
		columns = append(columns, columnRef{table: table, column: column})
	}
	return columns, rows.Err()
}

func processColumn(db *sql.DB, col columnRef, old, new string, dryRun bool) (matches int64, updated int64, err error) {
	table := quoteIdent(col.table)
	column := quoteIdent(col.column)
	likePattern := "%" + old + "%"

	countQuery := fmt.Sprintf(
		"SELECT COUNT(*) FROM %s WHERE %s LIKE ?",
		table, column,
	)
	if err := db.QueryRow(countQuery, likePattern).Scan(&matches); err != nil {
		return 0, 0, err
	}
	if matches == 0 {
		return 0, 0, nil
	}
	if dryRun {
		return matches, matches, nil
	}

	updateQuery := fmt.Sprintf(
		"UPDATE %s SET %s = REPLACE(%s, ?, ?) WHERE %s LIKE ?",
		table, column, column, column,
	)
	res, err := db.Exec(updateQuery, old, new, likePattern)
	if err != nil {
		return matches, 0, err
	}
	updated, err = res.RowsAffected()
	if err != nil {
		return matches, 0, err
	}
	return matches, updated, nil
}

func quoteIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
