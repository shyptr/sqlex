package sqlex

import (
	"errors"
	"fmt"
	"io"
)

type part struct {
	pred interface{}
	args []interface{}
}

func newPart(pred interface{}, args ...interface{}) Sqlex {
	return &part{pred, args}
}

func (p part) ToSql() (sql string, args []interface{}, err error) {
	switch pred := p.pred.(type) {
	case nil:
		// no-op
	case Sqlex:
		sql, args, err = pred.ToSql()
	case string:
		sql = pred
		args = p.args
	default:
		err = fmt.Errorf("expected string or Sqlex, not %T", pred)
	}
	return
}

var noSql = errors.New("there is non sqlStr from toSql()")

func appendToSql(parts []Sqlex, w io.Writer, sep string, args []interface{}) ([]interface{}, error) {
	build := func(b Sqlex) (err error) {
		baseSql, baseArgs, err := b.ToSql()
		if err != nil {
			return err
		}
		if baseSql == "" {
			return noSql
		}
		_, err = io.WriteString(w, baseSql)
		if err != nil {
			return
		}
		args = append(args, baseArgs...)
		return
	}
	var skip bool
	if err := build(parts[0]); err != nil {
		if err == noSql {
			skip = true
		} else {
			return nil, err
		}
	}
	for _, part := range parts[1:] {
		if !skip {
			if _, err := io.WriteString(w, sep); err != nil {
				return nil, err
			}
		} else {
			skip = false
		}
		if err := build(part); err != nil {
			if err == noSql {
				skip = true
			} else {
				return nil, err
			}
		}
	}
	return args, nil
}
