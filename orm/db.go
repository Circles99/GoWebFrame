package orm

type DBOption func(db *DB)

type DB struct {
	r *register
}

func NewDB(opts ...DBOption) (*DB, error) {
	db := &DB{
		r: &register{},
	}
	for _, opt := range opts {
		opt(db)
	}
	return db, nil
}

// MustNewDB 创建一个 DB，如果失败则会 panic
// 我个人不太喜欢这种
func MustNewDB(opts ...DBOption) *DB {
	db, err := NewDB(opts...)
	if err != nil {
		panic(err)
	}
	return db
}
