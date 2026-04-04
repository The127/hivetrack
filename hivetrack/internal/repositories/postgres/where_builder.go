package postgres

import "fmt"

// whereBuilder accumulates WHERE clauses and positional arguments for a SQL query.
// It handles the common patterns: equality checks, IS NULL, EXISTS subqueries, and
// arbitrary parameterized clauses.
//
// Important: all values are passed pre-dereferenced. The caller checks for nil pointers
// and passes (ok bool, val any) to avoid the typed-nil-in-interface pitfall.
type whereBuilder struct {
	query  string
	args   []any
	argIdx int
}

func newWhereBuilder(baseQuery string) *whereBuilder {
	return &whereBuilder{
		query:  baseQuery + ` WHERE 1=1`,
		argIdx: 1,
	}
}

// eq appends "AND <column> = $N" if ok is true. The column should be a column
// reference like "i.project_id".
func (w *whereBuilder) eq(column string, ok bool, val any) {
	if !ok {
		return
	}
	w.query += fmt.Sprintf(` AND %s=$%d`, column, w.argIdx)
	w.args = append(w.args, val)
	w.argIdx++
}

// clause appends an arbitrary "AND <sqlFragment>" containing exactly one "$%d"
// placeholder, bound to val. Skipped if ok is false.
func (w *whereBuilder) clause(sqlFragment string, ok bool, val any) {
	if !ok {
		return
	}
	w.query += fmt.Sprintf(` AND `+sqlFragment, w.argIdx)
	w.args = append(w.args, val)
	w.argIdx++
}

// raw appends a static "AND <sqlFragment>" with no parameters. Skipped if cond is false.
func (w *whereBuilder) raw(cond bool, sqlFragment string) {
	if !cond {
		return
	}
	w.query += ` AND ` + sqlFragment
}
