[![Project Status: Inactive – The project has reached a stable, usable state but is no longer being actively developed; support/maintenance will be provided as time allows.](https://www.repostatus.org/badges/latest/inactive.svg)](https://www.repostatus.org/#inactive)

sqlex is forks from [Squirrel](github.com/Masterminds/squirrel)


# sqlex - fluent SQL generator for Go

```go
import "github.com/unrotten/sqlex"
```


[![GoDoc](https://godoc.org/github.com/unrotten/sqlex?status.png)](https://godoc.org/github.com/unrotten/sqlex)
[![Build Status](https://api.travis-ci.org/unrotten/sqlex.svg?branch=master)](https://travis-ci.org/unrotten/sqlex)

**sqlex is not an ORM.** For an application of sqlex, check out
[structable, a table-struct mapper](https://github.com/Masterminds/structable)


sqlex helps you build SQL queries from composable parts:

```go
import sq "github.com/unrotten/sqlex"

users := sq.Select("*").From("users").Join("emails USING (email_id)")

active := users.Where(sq.Eq{"deleted_at": nil})

sql, args, err := active.ToSql()

sql == "SELECT * FROM users JOIN emails USING (email_id) WHERE deleted_at IS NULL"
```

```go
sql, args, err := sq.
    Insert("users").Columns("name", "age").
    Values("moe", 13).Values("larry", sq.Expr("? + 5", 12)).
    ToSql()

sql == "INSERT INTO users (name,age) VALUES (?,?),(?,? + 5)"
```

sqlex can also execute queries directly:

```go
stooges := users.Where(sq.Eq{"username": []string{"moe", "larry", "curly", "shemp"}})
three_stooges := stooges.Limit(3)
rows, err := three_stooges.RunWith(db).Query()

// Behaves like:
rows, err := db.Query("SELECT * FROM users WHERE username IN (?,?,?,?) LIMIT 3",
                      "moe", "larry", "curly", "shemp")
```

sqlex makes conditional query building a breeze:

```go
if len(q) > 0 {
    users = users.Where("name LIKE ?", fmt.Sprint("%", q, "%"))
}
```

sqlex wants to make your life easier:

```go
// StmtCache caches Prepared Stmts for you
dbCache := sq.NewStmtCacher(db)

// StatementBuilder keeps your syntax neat
mydb := sq.StatementBuilder.RunWith(dbCache)
select_users := mydb.Select("*").From("users")
```

sqlex loves PostgreSQL:

```go
psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// You use question marks for placeholders...
sql, _, _ := psql.Select("*").From("elephants").Where("name IN (?,?)", "Dumbo", "Verna").ToSql()

/// ...sqlex replaces them using PlaceholderFormat.
sql == "SELECT * FROM elephants WHERE name IN ($1,$2)"


/// You can retrieve id ...
query := sq.Insert("nodes").
    Columns("uuid", "type", "data").
    Values(node.Uuid, node.Type, node.Data).
    Suffix("RETURNING \"id\"").
    RunWith(m.db).
    PlaceholderFormat(sq.Dollar)

query.QueryRow().Scan(&node.id)
```

You can escape question marks by inserting two question marks:

```sql
SELECT * FROM nodes WHERE meta->'format' ??| array[?,?]
```

will generate with the Dollar Placeholder:

```sql
SELECT * FROM nodes WHERE meta->'format' ?| array[$1,$2]
```

## FAQ

* **How can I build an IN query on composite keys / tuples, e.g. `WHERE (col1, col2) IN ((1,2),(3,4))`? ([#104](https://github.com/unrotten/sqlex/issues/104))**

    sqlex does not explicitly support tuples, but you can get the same effect with e.g.:

    ```go
    sq.Or{
      sq.Eq{"col1": 1, "col2": 2},
      sq.Eq{"col1": 3, "col2": 4}}
    ```

    ```sql
    WHERE (col1 = 1 AND col2 = 2) OR (col1 = 3 AND col2 = 4)
    ```

    and I add IF condition into this application,so we can use like with e.g:
    
    ```go
    sq.IF{"a"!="",sq.Eq{"col1":1}}.IF{1==0,sq.Eq{"col2":2}}
    ```
  
    ```sql
    WHERE col1 = 1
    ```
    
    (which should produce the same query plan as the tuple version)

* **Why doesn't `Eq{"mynumber": []uint8{1,2,3}}` turn into an `IN` query? ([#114](https://github.com/unrotten/sqlex/issues/114))**

    Values of type `[]byte` are handled specially by `database/sql`. In Go, [`byte` is just an alias of `uint8`](https://golang.org/pkg/builtin/#byte), so there is no way to distinguish `[]uint8` from `[]byte`.

* **Some features are poorly documented!**

    This isn't a frequent complaints section!

* **Some features are poorly documented?**

    Yes. The tests should be considered a part of the documentation; take a look at those for ideas on how to express more complex queries.

## License

sqlex is released under the
[MIT License](http://www.opensource.org/licenses/MIT).
