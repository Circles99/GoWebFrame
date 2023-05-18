package orm

import "GoWebFrame/orm/errs"

// Package orm
// @Description: dialect用于区别不同的数据库特殊写法

type Dialect interface {
	quoter() byte
	buildUpsert(b *builder, odk *OnDuplicateKey) error
}

type MysqlDialect struct {
}

func (m MysqlDialect) quoter() byte {
	return '`'
}

func (m MysqlDialect) buildUpsert(b *builder, odk *OnDuplicateKey) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			b.sb.WriteByte('`')
			fd, ok := b.model.FieldMap[a.column]
			if !ok {
				return errs.NewErrUnknownField(a.column)
			}
			b.sb.WriteString(fd.ColName)
			b.sb.WriteByte('`')
			b.sb.WriteString("=?")
			b.addArgs(a.val)
		case Column:
			b.sb.WriteByte('`')
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.sb.WriteString(fd.ColName)
			b.sb.WriteByte('`')
			b.sb.WriteString("=VALUES(`")
			b.sb.WriteString(fd.ColName)
			b.sb.WriteString("`)")
		default:
			return errs.NewErrUnsupportedAssignableType(a)
		}
	}
	return nil
}
