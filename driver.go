package sqlcomment

import (
	"context"
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
)

type (
	// Driver is a driver that adds an SQL comment.
	// See: https://google.github.io/sqlcommenter.
	Driver struct {
		dialect.Driver // underlying driver.
		commenter
	}

	// Tx is a transaction implementation that adds an SQL comment.
	Tx struct {
		dialect.Tx                 // underlying transaction.
		ctx        context.Context // underlying transaction context.
		commenter
	}

	commenter struct {
		options
	}
)

// NewDriver decorates the given driver and adds an SQL comment to every query.
func NewDriver(drv dialect.Driver, options ...Option) dialect.Driver {
	taggers := []Tagger{contextTagger{}}
	opts := buildOptions(append(options, WithTagger(taggers...)))
	return &Driver{drv, commenter{opts}}
}

func (c commenter) withComment(ctx context.Context, query string) string {
	tags := make(Tags)
	for _, h := range c.taggers {
		tags.Merge(h.Tag(ctx))
	}
	return fmt.Sprintf("%s /*%s*/", query, tags.Marshal())
}

// Query adds an SQL comment to the original query and calls the underlying driver Query method.
func (d *Driver) Query(ctx context.Context, query string, args, v interface{}) error {
	return d.Driver.Query(ctx, d.withComment(ctx, query), args, v)
}

// QueryContext calls QueryContext of the underlying driver, or fails if it is not supported.
func (d *Driver) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	drv, ok := d.Driver.(interface {
		QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	})
	if !ok {
		return nil, fmt.Errorf("Driver.QueryContext is not supported")
	}
	return drv.QueryContext(ctx, d.withComment(ctx, query), args...)
}

// Exec adds an SQL comment to the original query and calls the underlying driver Exec method.
func (d *Driver) Exec(ctx context.Context, query string, args, v interface{}) error {
	return d.Driver.Exec(ctx, d.withComment(ctx, query), args, v)
}

// ExecContext calls ExecContext of the underlying driver, or fails if it is not supported.
func (d *Driver) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	drv, ok := d.Driver.(interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	})
	if !ok {
		return nil, fmt.Errorf("Driver.ExecContext is not supported")
	}
	return drv.ExecContext(ctx, d.withComment(ctx, query), args...)
}

// Tx wraps the underlying Tx command with a commenter.
func (d *Driver) Tx(ctx context.Context) (dialect.Tx, error) {
	tx, err := d.Driver.Tx(ctx)
	if err != nil {
		return nil, err
	}
	return &Tx{tx, ctx, d.commenter}, nil
}

// BeginTx wraps the underlying transaction with commenter and calls the underlying driver BeginTx command if it's supported.
func (d *Driver) BeginTx(ctx context.Context, opts *sql.TxOptions) (dialect.Tx, error) {
	drv, ok := d.Driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	})
	if !ok {
		return nil, fmt.Errorf("Driver.BeginTx is not supported")
	}
	tx, err := drv.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx, ctx, d.commenter}, nil
}

// Exec adds an SQL comment and calls the underlying transaction Exec method.
func (d *Tx) Exec(ctx context.Context, query string, args, v interface{}) error {
	return d.Tx.Exec(ctx, d.withComment(ctx, query), args, v)
}

// ExecContext logs its params and calls the underlying transaction ExecContext method if it is supported.
func (d *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	tx, ok := d.Tx.(interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	})
	if !ok {
		return nil, fmt.Errorf("Tx.ExecContext is not supported")
	}
	return tx.ExecContext(ctx, d.withComment(ctx, query), args...)
}

// Query adds an SQL comment and calls the underlying transaction Query method.
func (d *Tx) Query(ctx context.Context, query string, args, v interface{}) error {
	return d.Tx.Query(ctx, d.withComment(ctx, query), args, v)
}

// QueryContext logs its params and calls the underlying transaction QueryContext method if it is supported.
func (d *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	tx, ok := d.Tx.(interface {
		QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	})
	if !ok {
		return nil, fmt.Errorf("Tx.QueryContext is not supported")
	}
	return tx.QueryContext(ctx, d.withComment(ctx, query), args...)
}

// Commit commits the underlying Tx.
func (d *Tx) Commit() error {
	return d.Tx.Commit()
}

// Rollback rolls back the underlying Tx.
func (d *Tx) Rollback() error {
	return d.Tx.Rollback()
}
