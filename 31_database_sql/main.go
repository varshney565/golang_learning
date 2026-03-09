// Package main demonstrates database/sql in Go.
// Topics: connection pool, queries, transactions, prepared statements, patterns.
// Uses SQLite via mattn/go-sqlite3 (CGo) conceptually — shown with comments.
//
// To run with a real DB:
//   go get github.com/mattn/go-sqlite3
//   go get github.com/lib/pq          (PostgreSQL)
//   go get github.com/go-sql-driver/mysql (MySQL)
//
// This file shows the patterns with a mock to avoid CGo dependency.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: database/sql Architecture
// -----------------------------------------------------------------------
// database/sql is a generic interface — NOT tied to any database.
// You register a DRIVER (postgres, mysql, sqlite3) via blank import:
//   import _ "github.com/lib/pq"
//
// The package manages a CONNECTION POOL automatically:
//   - Opens connections on demand
//   - Reuses idle connections
//   - Closes connections that have been idle too long
//
// Key types:
//   *sql.DB     — the pool (safe for concurrent use, keep one per app)
//   *sql.Tx     — a transaction
//   *sql.Stmt   — a prepared statement
//   *sql.Rows   — result set from a query
//   *sql.Row    — single row result

// -----------------------------------------------------------------------
// SECTION 2: Opening and Configuring the Pool
// -----------------------------------------------------------------------

func openDatabase(dsn string) (*sql.DB, error) {
	// sql.Open does NOT actually connect — it just validates the DSN.
	// The first real connection happens on the first query.
	db, err := sql.Open("postgres", dsn) // driver name + DSN
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// ── Connection Pool Settings ──────────────────────────────
	// MaxOpenConns: max simultaneous open connections (0 = unlimited)
	// Set this to match your DB server's connection limit.
	db.SetMaxOpenConns(25)

	// MaxIdleConns: max connections to keep idle
	// Should be <= MaxOpenConns
	db.SetMaxIdleConns(5)

	// ConnMaxLifetime: close connections older than this
	// Prevents using stale connections after a DB restart/failover
	db.SetConnMaxLifetime(5 * time.Minute)

	// ConnMaxIdleTime: close connections that have been idle too long
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Verify the connection actually works
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	return db, nil
}

// -----------------------------------------------------------------------
// SECTION 3: Query Patterns
// -----------------------------------------------------------------------

type User struct {
	ID        int
	Name      string
	Email     string
	CreatedAt time.Time
}

// QueryRow — single row (use when you expect exactly one result)
func getUserByID(ctx context.Context, db *sql.DB, id int) (*User, error) {
	query := `SELECT id, name, email, created_at FROM users WHERE id = $1`

	var u User
	// QueryRowContext is preferred over QueryRow — respects context (timeout/cancel)
	err := db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user %d: not found", id)
		}
		return nil, fmt.Errorf("getUserByID: %w", err)
	}
	return &u, nil
}

// QueryContext — multiple rows (MUST close rows when done)
func listUsers(ctx context.Context, db *sql.DB) ([]*User, error) {
	query := `SELECT id, name, email, created_at FROM users ORDER BY id`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listUsers: %w", err)
	}
	defer rows.Close() // ALWAYS defer rows.Close() — releases the connection

	var users []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("listUsers scan: %w", err)
		}
		users = append(users, &u)
	}

	// Check for errors from iteration (network failure mid-query, etc.)
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listUsers rows: %w", err)
	}

	return users, nil
}

// ExecContext — INSERT/UPDATE/DELETE (no rows returned)
func createUser(ctx context.Context, db *sql.DB, name, email string) (int64, error) {
	query := `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`

	// For PostgreSQL RETURNING clause, use QueryRow
	var id int64
	err := db.QueryRowContext(ctx, query, name, email).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("createUser: %w", err)
	}
	return id, nil

	// For MySQL/SQLite (no RETURNING):
	// result, err := db.ExecContext(ctx, query, name, email)
	// id, _ = result.LastInsertId()
}

// -----------------------------------------------------------------------
// SECTION 4: Transactions
// -----------------------------------------------------------------------
// A transaction ensures multiple operations succeed or fail together (ACID).
//
// Pattern: always use defer for rollback, commit explicitly at the end.

func transferMoney(ctx context.Context, db *sql.DB, fromID, toID int, amount float64) error {
	// Begin transaction
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable, // isolation level
		ReadOnly:  false,
	})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() // runs if we return before Commit(), is a no-op after Commit()

	// Debit from sender
	result, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1`,
		amount, fromID,
	)
	if err != nil {
		return fmt.Errorf("debit: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("insufficient funds or account %d not found", fromID)
	}

	// Credit to receiver
	if _, err = tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance + $1 WHERE id = $2`,
		amount, toID,
	); err != nil {
		return fmt.Errorf("credit: %w", err)
	}

	// Commit only if everything succeeded
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// -----------------------------------------------------------------------
// SECTION 5: Prepared Statements
// -----------------------------------------------------------------------
// PrepareContext compiles a query on the DB server once.
// Benefits: security (prevents SQL injection), performance (no re-parse).
//
// Use when: same query runs many times with different params (batch inserts, etc.)

func batchInsert(ctx context.Context, db *sql.DB, users []User) error {
	// Prepare once — cached on DB server
	stmt, err := db.PrepareContext(ctx,
		`INSERT INTO users (name, email) VALUES ($1, $2)`,
	)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close() // releases the prepared statement on the server

	for _, u := range users {
		if _, err := stmt.ExecContext(ctx, u.Name, u.Email); err != nil {
			return fmt.Errorf("insert user %s: %w", u.Name, err)
		}
	}
	return nil
}

// -----------------------------------------------------------------------
// SECTION 6: Nullable Types — sql.Null*
// -----------------------------------------------------------------------
// SQL columns can be NULL. Go's int/string etc. can't represent NULL.
// Use sql.NullString, sql.NullInt64, sql.NullFloat64, etc.

type UserProfile struct {
	ID       int
	Name     string
	Bio      sql.NullString // bio can be NULL
	Age      sql.NullInt64  // age can be NULL
}

func scanNullable(ctx context.Context, db *sql.DB) error {
	query := `SELECT id, name, bio, age FROM users WHERE id = $1`
	var p UserProfile
	err := db.QueryRowContext(ctx, query, 1).Scan(&p.ID, &p.Name, &p.Bio, &p.Age)
	if err != nil {
		return err
	}

	// Check if the nullable field has a value
	if p.Bio.Valid {
		fmt.Printf("  bio: %s\n", p.Bio.String)
	} else {
		fmt.Println("  bio: (null)")
	}

	if p.Age.Valid {
		fmt.Printf("  age: %d\n", p.Age.Int64)
	}
	return nil
}

// -----------------------------------------------------------------------
// SECTION 7: Repository Pattern (Clean Architecture)
// -----------------------------------------------------------------------
// Wrap database operations in a repository to:
//   - Hide SQL from business logic
//   - Make code testable (swap real DB with mock)
//   - Centralize query logic

type UserRepository interface {
	GetByID(ctx context.Context, id int) (*User, error)
	Create(ctx context.Context, name, email string) (*User, error)
	List(ctx context.Context) ([]*User, error)
}

type postgresUserRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepo{db: db}
}

func (r *postgresUserRepo) GetByID(ctx context.Context, id int) (*User, error) {
	return getUserByID(ctx, r.db, id)
}

func (r *postgresUserRepo) Create(ctx context.Context, name, email string) (*User, error) {
	id, err := createUser(ctx, r.db, name, email)
	if err != nil {
		return nil, err
	}
	return &User{ID: int(id), Name: name, Email: email}, nil
}

func (r *postgresUserRepo) List(ctx context.Context) ([]*User, error) {
	return listUsers(ctx, r.db)
}

// -----------------------------------------------------------------------
// SECTION 8: Key Patterns Summary
// -----------------------------------------------------------------------

func summary() {
	fmt.Println("database/sql key patterns:")

	patterns := []struct{ pattern, rule string }{
		{"One *sql.DB per app", "it's a pool, not a single connection"},
		{"Always defer rows.Close()", "prevents connection leaks"},
		{"Check rows.Err()", "errors during iteration after rows.Next()"},
		{"Use *Context variants", "QueryRowContext, ExecContext etc. — support cancellation"},
		{"defer tx.Rollback()", "safe no-op after Commit(); prevents hanging transactions"},
		{"sql.ErrNoRows", "QueryRow with no result — not a real error, handle explicitly"},
		{"sql.Null* types", "for nullable columns in DB schema"},
		{"Prepared statements", "for repeated queries — compile once, execute many"},
		{"Repository pattern", "hide SQL, enable mocking in tests"},
		{"Never fmt.Sprintf SQL", "always use $1/$2 placeholders — prevents SQL injection"},
	}

	for _, p := range patterns {
		fmt.Printf("  %-35s → %s\n", p.pattern, p.rule)
	}
}

func main() {
	// Note: This file shows patterns — a real DB connection needs a driver.
	// To test with SQLite:
	//   import _ "github.com/mattn/go-sqlite3"
	//   db, _ := sql.Open("sqlite3", ":memory:")
	fmt.Println("database/sql patterns (requires a driver to run queries)")
	fmt.Println("Import drivers with blank import:")
	fmt.Println(`  import _ "github.com/lib/pq"                   // PostgreSQL`)
	fmt.Println(`  import _ "github.com/go-sql-driver/mysql"      // MySQL`)
	fmt.Println(`  import _ "github.com/mattn/go-sqlite3"         // SQLite`)
	fmt.Println()
	summary()

	// Demonstrate openDatabase signature (won't actually connect)
	fmt.Printf("\nopenDatabase signature: %T\n", openDatabase)
	fmt.Printf("getUserByID signature:  %T\n", getUserByID)
	fmt.Printf("transferMoney signature:%T\n", transferMoney)
}
