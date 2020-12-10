package pagination

import (
	"log"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	defaultCursorWithAddField = &Cursor{
		Fields: []Field{
			Field{
				Name:      "id",
				Value:     nil,
				Direction: DirectionAsc,
			},
			Field{
				Name:      "name",
				Value:     nil,
				Direction: DirectionDesc,
			},
		},
		Limit:    defaultLimit,
		Backward: false,
	}

	queryCursor = &Cursor{
		Fields: []Field{
			Field{
				Name:      "id",
				Value:     15,
				Direction: DirectionAsc,
			},
			Field{
				Name:      "name",
				Value:     "ivan",
				Direction: DirectionDesc,
			},
		},
		Limit:    defaultLimit,
		Backward: false,
	}
	queryCursorStringWhere          = `SELECT * FROM "users" WHERE id > $1 OR (id = $2 AND name < $3)`
	queryCursorStringWhereWithOrder = `SELECT * FROM "users" WHERE id > $1 OR (id = $2 AND name < $3) ORDER BY id asc,name desc LIMIT 4`

	queryCursorStringWhereWithGroupCondition = `SELECT * FROM "users" WHERE (id > $1 OR (id = $2 AND name < $3)) AND count < $4 ORDER BY id asc,name desc LIMIT 4`

	defaultCursorEncodeBase64Str = `eyJmaWVsZHMiOlt7Im5hbWUiOiJpZCIsInZhbHVlIjpudWxsLCJkaXJlY3Rpb24iOiJhc2MifV0sImxpbWl0IjozLCJiYWNrd2FyZCI6ZmFsc2V9`
)

func TestNew(t *testing.T) {
	cursor := (&Cursor{}).New(defaultLimit, Field{"id", nil, DirectionAsc})

	if !reflect.DeepEqual(defaultCursor, cursor) {
		t.Error("Default cursor failed")
	}
}

func TestAddField(t *testing.T) {
	defCursor := DefaultCursor()
	defCursor.AddField("name", nil, DirectionDesc)

	if !reflect.DeepEqual(defaultCursorWithAddField, defCursor) {
		t.Error("Failed AddField to cursor")
	}
}

func TestBackward(t *testing.T) {
	dd := DirectionAsc.Backward()
	if dd != DirectionDesc {
		t.Error("Failed DirectionAsc backward")
	}

	da := DirectionDesc.Backward()
	if da != DirectionAsc {
		t.Error("Failed DirectionDesc backward")
	}
}

func TestEncode(t *testing.T) {
	defCursor := DefaultCursor()
	str := defCursor.Encode()
	if str == "" {
		t.Error("Failed cursor Encode")
	}
	log.Println(str)

	if str != defaultCursorEncodeBase64Str {
		t.Error("Encode string not equal result")
	}
}

func TestWhere(t *testing.T) {
	sqlDB, _, _ := sqlmock.New()
	db, _ := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		PrepareStmt: true,
	})
	db = db.Session(&gorm.Session{DryRun: true})

	type User struct {
		ID   uint
		Name string
	}
	var user User

	dbAdditionalQuery := queryCursor.where(db)
	stmt := dbAdditionalQuery.Find(user).Statement
	sql := stmt.SQL.String()

	if queryCursorStringWhere != sql {
		t.Errorf("Query\n`%v`\nnot equal\n`%v`\n", sql, queryCursorStringWhere)
	}
}

func TestOrder(t *testing.T) {
	sqlDB, _, _ := sqlmock.New()
	db, _ := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		PrepareStmt: true,
	})
	db = db.Session(&gorm.Session{DryRun: true})

	type User struct {
		ID   uint
		Name string
	}
	var user User

	dbAdditionalQuery := queryCursor.order(queryCursor.where(db))
	stmt := dbAdditionalQuery.Find(user).Statement
	sql := stmt.SQL.String()

	if queryCursorStringWhereWithOrder != sql {
		t.Errorf("Query\n`%v`\nnot equal\n`%v`\n", sql, queryCursorStringWhereWithOrder)
	}
}

func TestWhereWithGroupConditions(t *testing.T) {
	sqlDB, _, _ := sqlmock.New()
	db, _ := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		PrepareStmt: true,
	})
	db = db.Session(&gorm.Session{DryRun: true})

	type User struct {
		ID    uint
		Name  string
		Count uint
	}
	var user User

	dbAdditionalQuery := queryCursor.GroupConditions(db)
	//scope := queryCursor.Scope()
	stmt := dbAdditionalQuery.Where("count < ?", 20).Find(user).Statement
	sql := stmt.SQL.String()

	log.Println(sql)
	if queryCursorStringWhereWithGroupCondition != sql {
		t.Errorf("Query\n`%v`\nnot equal\n`%v`\n", sql, queryCursorStringWhereWithGroupCondition)
	}
}

func TestScope(t *testing.T) {
	sqlDB, _, _ := sqlmock.New()
	db, _ := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		PrepareStmt: true,
	})
	db = db.Session(&gorm.Session{DryRun: true})

	type User struct {
		ID    uint
		Name  string
		Count uint
	}
	var user User

	scope := queryCursor.Scope()
	stmt := db.Scopes(scope).Where("count < ?", 20).Find(user).Statement
	sql := stmt.SQL.String()

	log.Println(sql)
	if queryCursorStringWhereWithGroupCondition != sql {
		t.Errorf("Query\n`%v`\nnot equal\n`%v`\n", sql, queryCursorStringWhereWithGroupCondition)
	}
}
