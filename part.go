package sqlex

import (
	"errors"
	"fmt"
	"io"
	"strings"
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
	var sqlArray []string
	for _, part := range parts {
		sql, arg, err := part.ToSql()
		if err != nil {
			return nil, err
		}
		if sql == "" {
			continue
		}
		sqlArray = append(sqlArray, sql)
		args = append(args, arg...)
	}
	_, err := io.WriteString(w, strings.Join(sqlArray, sep))
	if err != nil {
		return nil, err
	}
	return args, nil
}
