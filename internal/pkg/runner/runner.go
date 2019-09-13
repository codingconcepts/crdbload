package runner

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"text/template"
	"time"

	"github.com/codingconcepts/datagen/internal/pkg/random"

	"github.com/google/uuid"

	"github.com/codingconcepts/datagen/internal/pkg/parse"
	"github.com/pkg/errors"
)

// Runner holds the configuration that will be used at runtime.
type Runner struct {
	db      *sql.DB
	funcs   template.FuncMap
	helpers map[string]interface{}
	store   *store
	debug   bool

	dateFormat      string
	stringFdefaults random.StringFDefaults
}

// New returns a pointer to a newly configured Runner.  Optionally
// taking a variable number of configuration options.
func New(db *sql.DB, opts ...Option) *Runner {
	r := Runner{
		db:    db,
		store: newStore(),
		debug: false,
		stringFdefaults: random.StringFDefaults{
			StringMinDefault: 10,
			StringMaxDefault: 10,
			IntMinDefault:    10000,
			IntMaxDefault:    99999,
		},
	}

	for _, opt := range opts {
		opt(&r)
	}

	r.funcs = template.FuncMap{
		"string":  random.String,
		"stringf": random.StringF(r.stringFdefaults),
		"int":     random.Int,
		"date":    random.Date(r.dateFormat),
		"float":   random.Float,
		"uuid":    func() string { return uuid.New().String() },
		"set":     random.Set,
		"ref":     r.store.reference,
		"row":     r.store.row,
		"each":    r.store.each,
		"ntimes": func(min int64, extra ...int64) []struct{} {
			max := min
			if len(extra) > 0 {
				max = extra[0]
			}
			return make([]struct{}, random.Int(min, max))
		},
	}

	return &r
}

// Run executes a given block, returning any errors encountered.
func (r *Runner) Run(b parse.Block) error {
	tmpl, err := template.New("block").Funcs(r.funcs).Parse(b.Body)
	if err != nil {
		return errors.Wrap(err, "parsing template")
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, r.helpers); err != nil {
		return errors.Wrap(err, "executing template")
	}

	if r.debug {
		fmt.Println(buf.String())
		return nil
	}

	rows, err := r.db.Query(buf.String())
	if err != nil {
		return errors.Wrap(err, "executing query")
	}

	return r.scan(b, rows)
}

func (r *Runner) ResetEach(name string) {
	r.store.eachRow = 0
	r.store.currentGroup = 0
	r.store.eachContext = name
}

func (r *Runner) scan(b parse.Block, rows *sql.Rows) error {
	for rows.Next() {
		columnTypes, err := rows.ColumnTypes()
		if err != nil {
			return errors.Wrap(err, "getting columns types from result")
		}

		values := make([]interface{}, len(columnTypes))
		for i, ct := range columnTypes {
			switch ct.DatabaseTypeName() {
			case "UUID":
				values[i] = reflect.New(reflect.TypeOf("")).Interface()
			default:
				values[i] = reflect.New(ct.ScanType()).Interface()
			}
		}

		if err = rows.Scan(values...); err != nil {
			return errors.Wrap(err, "scanning columns")
		}

		curr := map[string]interface{}{}
		for i, ct := range columnTypes {
			values[i] = r.prepareValue(reflect.ValueOf(values[i]).Elem())
			curr[ct.Name()] = values[i]
		}
		r.store.set(b.Name, curr)
	}

	return nil
}

// prepareValue ensures that data being read out of the database following
// a scan is in the correct format for being re-inserted into the database
// during follow-up queries.
func (r *Runner) prepareValue(v reflect.Value) interface{} {
	switch v.Type() {
	case reflect.TypeOf(time.Time{}):
		t := v.Interface().(time.Time)
		return t.Format(r.dateFormat)
	default:
		return v
	}
}
