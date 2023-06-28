package orm

import (
	"GoWebFrame/orm/errs"
	"GoWebFrame/orm/interal/model"
	"GoWebFrame/orm/interal/valuer"
	"context"
	"database/sql"
)

type DBOption func(db *DB)

type DB struct {
	core
	db *sql.DB
}

func Open(driver string, dsn string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dsn)

	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

// OpenDB
// 我可以利用 OpenDB 来传入一个 mock 的DB
// sqlmock.Open 的 DB
func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		core: core{
			r:          model.NewRegister(),
			valCreator: valuer.NewUnsafeValues,
			dialect:    MysqlDialect{},
		},
		db: db,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func NewDB(opts ...DBOption) (*DB, error) {
	db := &DB{
		core: core{
			r:          model.NewRegister(),
			valCreator: valuer.NewUnsafeValues,
			dialect:    MysqlDialect{},
		},
	}
	for _, opt := range opts {
		opt(db)
	}
	return db, nil
}

// BeginTx 开起事务
// @author: liujiming
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{
		tx: tx,
		db: db,
	}, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) getCore() core {
	return db.core
}

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

// DoTx 帮维护事务的执行
// @author: liujiming
func (db *DB) DoTx(ctx context.Context, fn func(ctx context.Context, tx *Tx) error, opts *sql.TxOptions) (err error) {
	var tx *Tx
	tx, err = db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	panicked := true
	defer func() {
		if panicked || err != nil {
			if e := tx.Rollback(); e != nil {
				err = errs.NewErrFailToRollbackTx(err, e, panicked)
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(ctx, tx)
	panicked = false
	return err
}

//
//// MustNewDB 创建一个 DB，如果失败则会 panic
//// 我个人不太喜欢这种
//func MustNewDB(opts ...DBOption) *DB {
//	db, err := NewDB(opts...)
//	if err != nil {
//		panic(err)
//	}
//	return db
//}
