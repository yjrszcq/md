// SQL query builder utility
package util

import (
	"regexp"
	"strconv"
	"strings"
)

// Pre-compiled regex for count query generation
var selectFromRegex = regexp.MustCompile(`(?i)(?s)select(.*?)from`)

type SqlCompletion struct {
	initSql      string
	initCountSql string
	whereSql     strings.Builder
	groupSql     string
	havingSql    string
	orderSql     strings.Builder
	limitSql     string
	paramIndex   int
	whereParams  []interface{}
	limitParams  []interface{}
}

// InitSql sets the base SQL query and auto-generates count query.
// Note: auto-generation may fail for complex queries with subqueries.
// Use InitSqlAndCount for complex cases.
func (s *SqlCompletion) InitSql(sql string) *SqlCompletion {
	s.initSql = sql
	s.initCountSql = selectFromRegex.ReplaceAllString(sql, "select count(*) as count from")
	return s
}

// 设置初始sql语句和总行数语句
func (s *SqlCompletion) InitSqlAndCount(sql, countSql string) *SqlCompletion {
	s.initSql = sql
	s.initCountSql = countSql
	return s
}

// 获取sql语句
func (s *SqlCompletion) GetSql() string {
	return s.initSql + s.whereSql.String() + s.groupSql + s.havingSql + s.orderSql.String() + s.limitSql
}

// 获取行数sql语句
func (s *SqlCompletion) GetCountSql() string {
	return s.initCountSql + s.whereSql.String() + s.groupSql + s.havingSql
}

// 获取条件列表
func (s *SqlCompletion) GetParams() []interface{} {
	return append(s.whereParams, s.limitParams...)
}

// 获取行数条件列表
func (s *SqlCompletion) GetCountParams() []interface{} {
	return s.whereParams
}

// =
func (s *SqlCompletion) Eq(field string, param interface{}, isAnd bool) *SqlCompletion {
	s.where(field, param, "=", false, isAnd)
	return s
}

// !=
func (s *SqlCompletion) Ne(field string, param interface{}, isAnd bool) *SqlCompletion {
	s.where(field, param, "!=", false, isAnd)
	return s
}

// >
func (s *SqlCompletion) Gt(field string, param interface{}, isAnd bool) *SqlCompletion {
	s.where(field, param, ">", false, isAnd)
	return s
}

// <
func (s *SqlCompletion) Lt(field string, param interface{}, isAnd bool) *SqlCompletion {
	s.where(field, param, "<", false, isAnd)
	return s
}

// >=
func (s *SqlCompletion) Ge(field string, param interface{}, isAnd bool) *SqlCompletion {
	s.where(field, param, ">=", false, isAnd)
	return s
}

// <=
func (s *SqlCompletion) Le(field string, param interface{}, isAnd bool) *SqlCompletion {
	s.where(field, param, "<=", false, isAnd)
	return s
}

// like
func (s *SqlCompletion) Like(field string, param interface{}, isAnd bool) *SqlCompletion {
	s.where(field, param, "like", true, isAnd)
	return s
}

// in
func (s *SqlCompletion) In(field string, params []interface{}, isAnd bool) *SqlCompletion {
	s.whereIn(field, params, "in", isAnd)
	return s
}

// not in
func (s *SqlCompletion) NotIn(field string, params []interface{}, isAnd bool) *SqlCompletion {
	s.whereIn(field, params, "not in", isAnd)
	return s
}

// is null
func (s *SqlCompletion) IsNull(field string, isAnd bool) *SqlCompletion {
	s.whereHyphen(isAnd)
	s.whereSql.WriteString(field)
	s.whereSql.WriteString(" is null")
	return s
}

// is not null
func (s *SqlCompletion) IsNotNull(field string, isAnd bool) *SqlCompletion {
	s.whereHyphen(isAnd)
	s.whereSql.WriteString(field)
	s.whereSql.WriteString(" is not null")
	return s
}

// 分组，只设置最后一次
func (s *SqlCompletion) Group(fields string) *SqlCompletion {
	s.groupSql = " group by " + fields
	return s
}

// having条件，只设置最后一次
func (s *SqlCompletion) Having(fields string) *SqlCompletion {
	s.havingSql = " having " + fields
	return s
}

// 排序
func (s *SqlCompletion) Order(field string, isAsc bool) *SqlCompletion {
	needComma := true
	if s.orderSql.Len() == 0 {
		s.orderSql.WriteString(" order by ")
		needComma = false
	}
	if needComma {
		s.orderSql.WriteString(",")
	}
	s.orderSql.WriteString(field)
	if !isAsc {
		s.orderSql.WriteString(" desc")
	}
	return s
}

// Maximum allowed page size to prevent abuse
const maxPageSize = 100

// 分页，只设置最后一次
func (s *SqlCompletion) Limit(page, size int) *SqlCompletion {
	if page > 0 && size > 0 {
		// Cap size to prevent excessive data retrieval
		if size > maxPageSize {
			size = maxPageSize
		}
		offset := size * (page - 1)
		s.paramIndex = s.paramIndex + 1
		paramPlaceholder1 := "$" + strconv.Itoa(s.paramIndex)
		s.paramIndex = s.paramIndex + 1
		paramPlaceholder2 := "$" + strconv.Itoa(s.paramIndex)
		s.limitSql = " limit " + paramPlaceholder1 + " offset " + paramPlaceholder2
		s.limitParams = []interface{}{size, offset}
	}
	return s
}

// 添加where条件
func (s *SqlCompletion) where(field string, param interface{}, symbol string, isLike bool, isAnd bool) {
	s.paramIndex = s.paramIndex + 1
	paramPlaceholder := "$" + strconv.Itoa(s.paramIndex)
	s.whereHyphen(isAnd)
	s.whereSql.WriteString(field)
	s.whereSql.WriteString(" ")
	s.whereSql.WriteString(symbol)
	if isLike {
		s.whereSql.WriteString(" '%'||" + paramPlaceholder + "||'%'")
	} else {
		s.whereSql.WriteString(" " + paramPlaceholder)
	}
	s.whereParams = append(s.whereParams, param)
}

// 添加where条件（in / not in）
func (s *SqlCompletion) whereIn(field string, params []interface{}, symbol string, isAnd bool) {
	if len(params) == 0 {
		return
	}
	s.whereHyphen(isAnd)
	s.whereSql.WriteString(field)
	s.whereSql.WriteString(" ")
	s.whereSql.WriteString(symbol)
	s.whereSql.WriteString(" (")
	for i, v := range params {
		if i != 0 {
			s.whereSql.WriteString(",")
		}
		s.paramIndex = s.paramIndex + 1
		paramPlaceholder := "$" + strconv.Itoa(s.paramIndex)
		s.whereSql.WriteString(paramPlaceholder)
		s.whereParams = append(s.whereParams, v)
	}
	s.whereSql.WriteString(")")
}

// where条件，拼接where和and/or
func (s *SqlCompletion) whereHyphen(isAnd bool) {
	needHyphen := true
	if s.whereSql.Len() == 0 {
		s.whereSql.WriteString(" where ")
		needHyphen = false
	}
	if needHyphen {
		if isAnd {
			s.whereSql.WriteString(" and ")
		} else {
			s.whereSql.WriteString(" or ")
		}
	}
}
