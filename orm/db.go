package orm

import (
	"GoWebFrame/orm/interal/model"
	"GoWebFrame/orm/interal/valuer"
	"database/sql"
)

type DBOption func(db *DB)

type DB struct {
	r          model.Register
	ValCreator valuer.Creator
	db         *sql.DB
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
		r:          model.NewRegister(),
		db:         db,
		ValCreator: valuer.NewUnsafeValues,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func NewDB(opts ...DBOption) (*DB, error) {
	db := &DB{
		r:          model.NewRegister(),
		ValCreator: valuer.NewUnsafeValues,
	}
	for _, opt := range opts {
		opt(db)
	}
	return db, nil
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
