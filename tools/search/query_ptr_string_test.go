package search

import (
	"testing"
)

func TestResolveSearchQuery_containsPointerString(t *testing.T) {
	type Q struct {
		Name *string `search:"type:contains;column:name;table:sys_permission" form:"name"`
	}
	s := "用户管理-设置角色"
	q := Q{Name: &s}
	cond := &GormCondition{
		GormPublic: GormPublic{},
		Join:       make([]*GormJoin, 0),
	}
	ResolveSearchQuery("mysql", q, cond)
	args, ok := cond.Where["`sys_permission`.`name` like ?"]
	if !ok {
		t.Fatalf("missing expected where clause, got %#v", cond.Where)
	}
	if len(args) != 1 || args[0] != "%用户管理-设置角色%" {
		t.Fatalf("unexpected args: %#v", args)
	}
}
