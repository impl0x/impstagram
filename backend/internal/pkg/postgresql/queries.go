package postgresql

import (
	"strconv"
	"strings"
)

// atp could have just used a gorm ngl

type Query struct {
	query string
	args  []any
}

// returns the query string
func (q Query) Query() string {
	return q.query
}

// returns the query args
func (q Query) Args() []any {
	return q.args
}

func (q Query) WithReturning(colName string) Query {
	q.query += " RETURNING " + colName
	return q
}

func NewQuery(query string, args ...any) Query {
	return Query{
		query, args,
	}
}

var (
	// select * from table where x = y
	QuerySelectAllWhere = func(table string, x string, y any) Query {
		return Query{`SELECT * FROM $1 WHERE $2 = $3`, []any{x, y}}
	}
	// select * from table1 t1 join table2 t2 on t1.cond1 = t2.cond2 where t2.x = y
	//
	// where table1 is the table you are fetching the data from and table2 is the data used for the join and comparison for cond1(table1 col) and cond2(table2 col).
	// x and y is used for the where statement for table2
	QuerySelectAllInnerJoin = func(table1, table2 string, cond1, cond2 any, x string, y any) Query {
		return Query{`
		SELECT t1.*
		FROM $1 t1
		JOIN $2 t2 ON t1.$3 = t2.$4
		WHERE t2.$6 = $7
	`, []any{table1, table2, cond1, cond2, x, y}}
	}

	// Panics if len(fieldNames)!=len(values)
	QueryInsert = func(table string, fieldNames []string, values []any) Query {
		if len(fieldNames) != len(values) {
			panic("postgresql.Create: length of fieldNames not equal to length of values")
		}
		return Query{
			`INSERT INTO $1 (` + strings.Join(fieldNames, ", ") + `) VALUES (` + indexGen(2, len(values)) + `)`,
			append([]any{table}, values...),
		}
	}

	QueryUpdateOneWhere = func(table, field string, value any, x string, y any) Query {
		return Query{
			`UPDATE $1 SET $2 = $3 WHERE $4 = $5`,
			[]any{table, field, value, x, y},
		}
	}

	QueryDeleteWhere = func(table string, x string, y any) Query {
		return Query{
			`DELETE FROM $1 WHERE $2 = $3`,
			[]any{table, x, y},
		}
	}
)

func indexGen(start, cap int) string {
	var b strings.Builder
	for i := start; i < cap+start; i++ {
		b.WriteString("$" + strconv.Itoa(i) + ", ")
	}
	return b.String()
}
