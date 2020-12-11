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

	queryCursorStringWhere          = `SELECT * FROM "users" WHERE (id > $1 OR (id = $2 AND name < $3))`
	queryCursorStringWhereWithOrder = `SELECT * FROM "users" WHERE (id > $1 OR (id = $2 AND name < $3)) ORDER BY id asc,name desc LIMIT 4`

	queryCursorStringWhereWithGroupCondition = `SELECT * FROM "users" WHERE (id > $1 OR (id = $2 AND name < $3)) AND count < $4 ORDER BY id asc,name desc LIMIT 4`

	defaultCursorEncodeBase64Str = `eyJmaWVsZHMiOlt7Im5hbWUiOiJpZCIsInZhbHVlIjpudWxsLCJkaXJlY3Rpb24iOiJhc2MifV0sImxpbWl0IjozLCJiYWNrd2FyZCI6ZmFsc2V9`

	queryCursorOneField = &Cursor{
		Fields: []Field{
			Field{
				Name:      "id",
				Value:     15,
				Direction: DirectionAsc,
			},
		},
		Limit:    defaultLimit,
		Backward: false,
	}

	queryCursorStringWhereForOneField                   = `SELECT * FROM "users" WHERE id > $1`
	queryCursorStringWhereWithOrderForOneField          = `SELECT * FROM "users" WHERE id > $1 ORDER BY id asc LIMIT 4`
	queryCursorStringWhereWithGroupConditionForOneField = `SELECT * FROM "users" WHERE id > $1 AND count < $2 ORDER BY id asc LIMIT 4`

	queryCursorBackward = &Cursor{
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
		Backward: true,
	}
	queryCursorBackwardStringWhere          = `SELECT * FROM "users" WHERE (id < $1 OR (id = $2 AND name > $3))`
	queryCursorBackwardStringWhereWithOrder = `SELECT * FROM "users" WHERE (id < $1 OR (id = $2 AND name > $3)) ORDER BY id desc,name asc LIMIT 4`

	queryCursorNameID = &Cursor{
		Fields: []Field{
			Field{
				Name:      "name",
				Value:     nil,
				Direction: DirectionAsc,
			},
			Field{
				Name:      "id",
				Value:     nil,
				Direction: DirectionDesc,
			},
		},
		Limit:    defaultLimit,
		Backward: true,
	}
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
	dd := DirectionAsc.Backward(true)
	if dd != DirectionDesc {
		t.Error("Failed DirectionAsc backward")
	}

	da := DirectionDesc.Backward(true)
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

	//Backward
	dbAdditionalQuery = queryCursorBackward.where(db)
	stmt = dbAdditionalQuery.Find(user).Statement
	sql = stmt.SQL.String()

	if queryCursorBackwardStringWhere != sql {
		t.Errorf("[Backward=true] Query\n`%v`\nnot equal\n`%v`\n", sql, queryCursorBackwardStringWhere)
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

	//Backward
	dbAdditionalQuery = queryCursorBackward.order(queryCursorBackward.where(db))
	stmt = dbAdditionalQuery.Find(user).Statement
	sql = stmt.SQL.String()

	if queryCursorBackwardStringWhereWithOrder != sql {
		t.Errorf("[Backward=true] Query\n`%v`\nnot equal\n`%v`\n", sql, queryCursorBackwardStringWhereWithOrder)
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

	if queryCursorStringWhereWithGroupCondition != sql {
		t.Errorf("Query\n`%v`\nnot equal\n`%v`\n", sql, queryCursorStringWhereWithGroupCondition)
	}
}

func TestWhereForOneFieldCursor(t *testing.T) {
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

	dbAdditionalQuery := queryCursorOneField.where(db)
	stmt := dbAdditionalQuery.Find(user).Statement
	sql := stmt.SQL.String()

	if queryCursorStringWhereForOneField != sql {
		t.Errorf("Query\n`%v`\nnot equal\n`%v`\n", sql, queryCursorStringWhereForOneField)
	}
}

func TestOrderForOneFieldCursor(t *testing.T) {
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

	dbAdditionalQuery := queryCursorOneField.order(queryCursorOneField.where(db))
	stmt := dbAdditionalQuery.Find(user).Statement
	sql := stmt.SQL.String()

	if queryCursorStringWhereWithOrderForOneField != sql {
		t.Errorf("Query\n`%v`\nnot equal\n`%v`\n", sql, queryCursorStringWhereWithOrderForOneField)
	}
}

func TestWhereWithGroupConditionsForOneFieldCursor(t *testing.T) {
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

	dbAdditionalQuery := queryCursorOneField.GroupConditions(db)
	//scope := queryCursor.Scope()
	stmt := dbAdditionalQuery.Where(db.Where("count < ?", 20)).Find(user).Statement
	sql := stmt.SQL.String()

	if queryCursorStringWhereWithGroupConditionForOneField != sql {
		t.Errorf("Query\n`%v`\nnot equal\n`%v`\n", sql, queryCursorStringWhereWithGroupConditionForOneField)
	}
}
func TestScopeForOneFieldCursor(t *testing.T) {
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

	scope := queryCursorOneField.Scope()
	stmt := db.Scopes(scope).Where("count < ?", 20).Find(user).Statement
	sql := stmt.SQL.String()

	if queryCursorStringWhereWithGroupConditionForOneField != sql {
		t.Errorf("Query\n`%v`\nnot equal\n`%v`\n", sql, queryCursorStringWhereWithGroupConditionForOneField)
	}
}

func TestResult(t *testing.T) {
	type User struct {
		ID         uint
		NameOfUser string `gorm:"column:name" json:"name" cursor:"name"`
		Count      uint
	}

	var users []User = []User{
		User{ID: 0, NameOfUser: "A", Count: 99},
		User{ID: 1, NameOfUser: "B", Count: 90},
		User{ID: 2, NameOfUser: "C", Count: 80},
		User{ID: 3, NameOfUser: "C", Count: 70},
		User{ID: 4, NameOfUser: "C", Count: 70},
		User{ID: 5, NameOfUser: "C", Count: 70},
		User{ID: 6, NameOfUser: "D", Count: 40},
		User{ID: 7, NameOfUser: "E", Count: 30},
		User{ID: 8, NameOfUser: "F", Count: 20},
		User{ID: 9, NameOfUser: "G", Count: 10},
	}

	cursor := &Cursor{
		Fields: []Field{
			Field{
				Name:      "name",
				Value:     nil,
				Direction: DirectionAsc,
			},
			Field{
				Name:      "id",
				Value:     nil,
				Direction: DirectionDesc,
			},
		},
		Limit:    4,
		Backward: false,
	}

	nextCursor := &Cursor{
		Fields: []Field{
			Field{
				Name:      "name",
				Value:     "C",
				Direction: DirectionAsc,
			},
			Field{
				Name:      "id",
				Value:     3,
				Direction: DirectionDesc,
			},
		},
		Limit:    4,
		Backward: false,
	}
	nextCursorStr := nextCursor.Encode()

	prevCursor := &Cursor{
		Fields: []Field{
			Field{
				Name:      "name",
				Value:     "A",
				Direction: DirectionAsc,
			},
			Field{
				Name:      "id",
				Value:     0,
				Direction: DirectionDesc,
			},
		},
		Limit:    4,
		Backward: true,
	}
	prevCursorStr := prevCursor.Encode()

	response, usersResp := cursor.Result(users)

	log.Printf("resp: %+v\n", response)
	log.Printf("users: %+v\n", usersResp)

	if response.Next != nextCursorStr {
		t.Errorf("Fail. Bad next cursor")
	}
	if !response.HasNext {
		t.Errorf("Fail. Bad hasNext")
	}

	if response.Prev != prevCursorStr {
		t.Errorf("Fail. Bad prev cursor")
	}
	if response.HasPrev {
		t.Errorf("Fail. Bad hasPrev")
	}

}
