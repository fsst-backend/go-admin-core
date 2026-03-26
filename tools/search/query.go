package search

import (
	"fmt"
	"reflect"
	"strings"
)

// stringSearchValue returns the string used for LIKE / ILIKE patterns and order direction.
// reflect.Value.String() only works when Kind == String; for *string it returns "<*string Value>",
// which breaks optional query DTO fields bound as pointers.
func stringSearchValue(v reflect.Value) string {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.String {
		return v.String()
	}
	return fmt.Sprint(v.Interface())
}

const (
	// FromQueryTag tag标记
	FromQueryTag = "search"
	// Mysql 数据库标识
	Mysql = "mysql"
	// Postgres 数据库标识
	Postgres = "postgres"
)

// ResolveSearchQuery 解析
/**
 * 	exact / iexact 等于
 * 	contains / icontains 包含
 *	gt / gte 大于 / 大于等于
 *	lt / lte 小于 / 小于等于
 *	startswith / istartswith 以…起始
 *	endswith / iendswith 以…结束
 *	in
 *	isnull
 *  order 排序		e.g. order[key]=desc     order[key]=asc
 */
func ResolveSearchQuery(driver string, q interface{}, condition Condition) {
	qType := reflect.TypeOf(q)
	qValue := reflect.ValueOf(q)
	var tag string
	var ok bool
	var t *resolveSearchTag

	for i := 0; i < qType.NumField(); i++ {
		tag, ok = "", false
		tag, ok = qType.Field(i).Tag.Lookup(FromQueryTag)
		if !ok {
			//递归调用
			ResolveSearchQuery(driver, qValue.Field(i).Interface(), condition)
			continue
		}
		switch tag {
		case "-":
			continue
		}

		field := qValue.Field(i)

		// ⭐ 唯一生效规则
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		t = makeTag(tag)

		//解析 Postgres `语法不支持，单独适配
		if driver == Postgres {
			pgSql(driver, t, condition, qValue, i)
		} else {
			otherSql(driver, t, condition, qValue, i)
		}
	}
}

func unwrapValue(v reflect.Value) reflect.Value {
	// 必须是有效值
	if !v.IsValid() {
		return v
	}

	// 一直解指针（支持 **int 这种情况）
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return v // 理论上不会发生（上层已过滤）
		}
		v = v.Elem()
	}
	return v
}

func pgSql(driver string, t *resolveSearchTag, condition Condition, qValue reflect.Value, i int) {

	field := unwrapValue(qValue.Field(i))
	val := field.Interface()

	switch t.Type {
	case "left":
		//左关联
		join := condition.SetJoinOn(t.Type, fmt.Sprintf(
			"left join %s on %s.%s = %s.%s", t.Join, t.Join, t.On[0], t.Table, t.On[1],
		))
		ResolveSearchQuery(driver, val, join)
	case "exact", "iexact":
		condition.SetWhere(fmt.Sprintf("%s.%s = ?", t.Table, t.Column), []interface{}{val})
	case "icontains":
		condition.SetWhere(fmt.Sprintf("%s.%s ilike ?", t.Table, t.Column), []interface{}{"%" + stringSearchValue(qValue.Field(i)) + "%"})
	case "contains":
		condition.SetWhere(fmt.Sprintf("%s.%s like ?", t.Table, t.Column), []interface{}{"%" + stringSearchValue(qValue.Field(i)) + "%"})
	case "gt":
		condition.SetWhere(fmt.Sprintf("%s.%s > ?", t.Table, t.Column), []interface{}{val})
	case "gte":
		condition.SetWhere(fmt.Sprintf("%s.%s >= ?", t.Table, t.Column), []interface{}{val})
	case "lt":
		condition.SetWhere(fmt.Sprintf("%s.%s < ?", t.Table, t.Column), []interface{}{val})
	case "lte":
		condition.SetWhere(fmt.Sprintf("%s.%s <= ?", t.Table, t.Column), []interface{}{val})
	case "istartswith":
		condition.SetWhere(fmt.Sprintf("%s.%s ilike ?", t.Table, t.Column), []interface{}{stringSearchValue(qValue.Field(i)) + "%"})
	case "startswith":
		condition.SetWhere(fmt.Sprintf("%s.%s like ?", t.Table, t.Column), []interface{}{stringSearchValue(qValue.Field(i)) + "%"})
	case "iendswith":
		condition.SetWhere(fmt.Sprintf("%s.%s ilike ?", t.Table, t.Column), []interface{}{"%" + stringSearchValue(qValue.Field(i))})
	case "endswith":
		condition.SetWhere(fmt.Sprintf("%s.%s like ?", t.Table, t.Column), []interface{}{"%" + stringSearchValue(qValue.Field(i))})
	case "in":
		condition.SetWhere(fmt.Sprintf("%s.%s in (?)", t.Table, t.Column), []interface{}{val})
	case "isnull":
		if !(unwrapValue(field).IsZero() && field.IsNil()) {
			condition.SetWhere(fmt.Sprintf("%s.%s isnull", t.Table, t.Column), make([]interface{}, 0))
		}
	case "order":
		dir := stringSearchValue(qValue.Field(i))
		switch strings.ToLower(dir) {
		case "desc", "asc":
			condition.SetOrder(fmt.Sprintf("%s.%s %s", t.Table, t.Column, dir))
		}
	}
}

func otherSql(driver string, t *resolveSearchTag, condition Condition, qValue reflect.Value, i int) {

	field := unwrapValue(qValue.Field(i))
	val := field.Interface()

	switch t.Type {
	case "left":
		//左关联
		join := condition.SetJoinOn(t.Type, fmt.Sprintf(
			"left join `%s` on `%s`.`%s` = `%s`.`%s`",
			t.Join,
			t.Join,
			t.On[0],
			t.Table,
			t.On[1],
		))
		ResolveSearchQuery(driver, val, join)
	case "exact", "iexact":
		condition.SetWhere(fmt.Sprintf("`%s`.`%s` = ?", t.Table, t.Column), []interface{}{val})
	case "contains", "icontains":
		condition.SetWhere(fmt.Sprintf("`%s`.`%s` like ?", t.Table, t.Column), []interface{}{"%" + stringSearchValue(qValue.Field(i)) + "%"})
	case "gt":
		condition.SetWhere(fmt.Sprintf("`%s`.`%s` > ?", t.Table, t.Column), []interface{}{val})
	case "gte":
		condition.SetWhere(fmt.Sprintf("`%s`.`%s` >= ?", t.Table, t.Column), []interface{}{val})
	case "lt":
		condition.SetWhere(fmt.Sprintf("`%s`.`%s` < ?", t.Table, t.Column), []interface{}{val})
	case "lte":
		condition.SetWhere(fmt.Sprintf("`%s`.`%s` <= ?", t.Table, t.Column), []interface{}{val})
	case "startswith", "istartswith":
		condition.SetWhere(fmt.Sprintf("`%s`.`%s` like ?", t.Table, t.Column), []interface{}{stringSearchValue(qValue.Field(i)) + "%"})
	case "endswith", "iendswith":
		condition.SetWhere(fmt.Sprintf("`%s`.`%s` like ?", t.Table, t.Column), []interface{}{"%" + stringSearchValue(qValue.Field(i))})
	case "in":
		condition.SetWhere(fmt.Sprintf("`%s`.`%s` in (?)", t.Table, t.Column), []interface{}{val})
	case "isnull":
		if !(field.IsZero() && field.IsNil()) {
			condition.SetWhere(fmt.Sprintf("`%s`.`%s` isnull", t.Table, t.Column), make([]interface{}, 0))
		}
	case "order":
		dir := stringSearchValue(qValue.Field(i))
		switch strings.ToLower(dir) {
		case "desc", "asc":
			condition.SetOrder(fmt.Sprintf("`%s`.`%s` %s", t.Table, t.Column, dir))
		}
	}
}
