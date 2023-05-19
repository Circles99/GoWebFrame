package orm

import (
	"GoWebFrame/orm/errs"
	"context"
)

type Selector[T any] struct {
	builder
	tbl     string
	where   []Predicate
	columns []selectedAlias
	db      *DB
	offset  int
	limit   int
	groupBy []Column
	orderBy []Column
	having  []Predicate
}

// 应该在DB上创建Selector比较合适 golang 不支持这种写法 Method cannot have type parameters
//func (d DB) NewSelector[T any]() *Selector[T] {
//	return &Selector[T]{
//		db: d,
//	}
//}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
		// builder实例化
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
	}
}

// Select 加入需要查询的字段
func (s *Selector[T]) Select(cols ...selectedAlias) *Selector[T] {
	s.columns = cols
	return s
}

// From 加入表名，为了链式调用返回Selector[T]
func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.tbl = tbl
	return s
}

func (s *Selector[T]) Where(p ...Predicate) *Selector[T] {
	s.where = append(s.where, p...)
	return s
}

func (s *Selector[T]) Offset(offset int) *Selector[T] {
	s.offset = offset
	return s
}

func (s *Selector[T]) Limit(limit int) *Selector[T] {
	s.limit = limit
	return s
}

func (s *Selector[T]) GroupBy(cols ...Column) *Selector[T] {
	s.groupBy = cols
	return s
}

func (s *Selector[T]) Having(p ...Predicate) *Selector[T] {
	s.having = p
	return s
}

// todo 需要支持两个表达式 暂未实现
func (s *Selector[T]) OrderBy(cols ...Column) *Selector[T] {
	s.orderBy = cols
	return s
}

func (s *Selector[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)

	s.model, err = s.db.r.Get(&t)
	if err != nil {
		return nil, err
	}

	s.sb.WriteString("SELECT ")

	// 进行拼接
	if err = s.buildColumns(); err != nil {
		return nil, err
	}

	s.sb.WriteString(" FROM ")
	if s.tbl == "" {
		//var t T
		//// 获取类型
		//typ := reflect.TypeOf(t)
		// 获取名字
		s.quote(s.model.TableName)
	} else {
		s.sb.WriteString(s.tbl)
	}

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		if err = s.buildPredicates(s.where); err != nil {
			return nil, err
		}
	}

	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		if err = s.buildGroupBy(); err != nil {
			return nil, err
		}
	}

	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING ")
		if err = s.buildPredicates(s.having); err != nil {
			return nil, err
		}
	}

	if s.limit > 0 {
		s.sb.WriteString(" LIMIT ?")
		s.addArgs(s.limit)
	}

	if s.offset > 0 {
		s.sb.WriteString(" OFFSET ?")
		s.addArgs(s.offset)
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	query, err := s.Build()
	if err != nil {
		return nil, err
	}
	// 获取数据
	rows, err := s.db.db.QueryContext(ctx, query.SQL, query.Args)
	if err != nil {
		return nil, err
	}

	t := new(T)
	val := s.db.ValCreator(t, s.model)
	// 在这里灵活切换反射或者 unsafe

	return t, val.SetColumns(rows)
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

//func (s *Selector[T]) buildColumn(c Column) error {
//	fd, ok := s.model.FieldMap[c.name]
//	if !ok {
//		return fmt.Errorf("位置字段： %v", c.name)
//	}
//	s.sb.WriteByte('`')
//	s.sb.WriteString(fd.ColName)
//	s.sb.WriteByte('`')
//
//	if c.alias != "" {
//		s.buildAlias(c.alias)
//	}
//
//	return nil
//}

func (s *Selector[T]) buildAlias(alias string) {
	s.sb.WriteString(" AS ")
	s.sb.WriteByte('`')
	s.sb.WriteString(alias)
	s.sb.WriteByte('`')
}

func (s *Selector[T]) buildAggregate(a Aggregate) error {
	s.sb.WriteString(a.fn)
	s.sb.WriteByte('(')
	err := s.buildColumn(a.arg)
	if err != nil {
		return err
	}
	s.sb.WriteByte(')')
	if a.alias != "" {
		s.buildAlias(a.alias)
	}
	return nil
}

func (s *Selector[T]) buildGroupBy() error {
	for i, column := range s.groupBy {
		if i > 0 {
			s.sb.WriteByte(',')
		}

		fd, ok := s.model.FieldMap[column.name]
		if !ok {
			return errs.NewErrUnknownField(column.name)
		}
		s.quote(fd.ColName)

	}
	return nil
}

func (s *Selector[T]) buildColumns() error {

	if len(s.columns) == 0 {
		s.sb.WriteByte('*')
		return nil
	}
	for i, c := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		switch val := c.(type) {
		case Column:
			err := s.buildColumn(val.name)
			if err != nil {
				return err
			}
		case Aggregate:
			if err := s.buildAggregate(val); err != nil {
				return err
			}
		case RawExpr:
			s.sb.WriteString(val.raw)
			if len(val.args) > 0 {
				s.args = append(s.args, val.args...)
			}
		}
	}
	return nil

}
