package pagination

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/go-testfixtures/testfixtures/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/rosberry/go-pagination/common"
	"github.com/rosberry/go-pagination/cursor"
)

var (
	notComparePageInfo bool = false
	pageLimit               = 2
	debug                   = false
)

var fixtures *testfixtures.Loader

type (
	q = []map[string]string
	r struct {
		IDs      []uint
		PageInfo *PageInfo
	}

	tCase struct {
		Name   string
		Params []map[string]string
		Result r
	}
)

func TestMainFlow(t *testing.T) {
	// Grab our router
	router := SetupRouter()

	tCases := []tCase{
		{
			Name:   "Default query",
			Params: q{},
			Result: r{
				IDs: []uint{1, 2},
				PageInfo: &PageInfo{
					Next:      cursor.New(pageLimit).AddField("id", 2, common.DirectionAsc).Encode(),
					Prev:      cursor.New(pageLimit).AddField("id", 1, common.DirectionAsc).SetBackward().Encode(),
					HasNext:   true,
					HasPrev:   false,
					TotalRows: 7,
				},
			},
		},
		{
			Name: "Simple cursor: id desc (sorting query)",
			Params: q{
				{"sorting": `[
				{
					"field": "id",
					"direction": "desc"
				}
			]`},
			},
			Result: r{
				IDs: []uint{7, 6},
				PageInfo: &PageInfo{
					Next:    cursor.New(pageLimit).AddField("id", 6, common.DirectionDesc).Encode(),
					Prev:    cursor.New(pageLimit).AddField("id", 7, common.DirectionDesc).SetBackward().Encode(),
					HasNext: true, HasPrev: false, TotalRows: 7,
				},
			},
		},
		{
			Name: "Simple cursor: id desc (cursor query: page 2)",
			Params: q{
				{"cursor": cursor.New(pageLimit).AddField("id", 6, common.DirectionDesc).Encode()},
			},
			Result: r{
				IDs: []uint{5, 4},
				PageInfo: &PageInfo{
					Next:    cursor.New(pageLimit).AddField("id", 4, common.DirectionDesc).Encode(),
					Prev:    cursor.New(pageLimit).AddField("id", 5, common.DirectionDesc).SetBackward().Encode(),
					HasNext: true, HasPrev: true, TotalRows: 7,
				},
			},
		},
		{
			Name: "Simple cursor two field: comment asc, id desc (sorting query)",
			Params: q{
				{"sorting": `[
					{
						"field": "comment",
						"direction": "asc"
					},
					{
						"field": "id",
						"direction": "desc"
					}
			]`},
			},
			Result: r{
				IDs: []uint{7, 6},
				PageInfo: &PageInfo{
					Next:    cursor.New(pageLimit).AddField("comment", "A", common.DirectionAsc).AddField("id", 6, common.DirectionDesc).Encode(),
					Prev:    cursor.New(pageLimit).AddField("comment", "A", common.DirectionAsc).AddField("id", 7, common.DirectionDesc).SetBackward().Encode(),
					HasNext: true, HasPrev: false, TotalRows: 7,
				},
			},
		},
		{
			Name: "Limit > row_count",
			Params: q{
				{
					"sorting": `[
					{
						"field": "id",
						"direction": "asc"
					}
				]`,
					"limit": "10",
				},
			},
			Result: r{
				IDs: []uint{1, 2, 3, 4, 5, 6, 7},
				PageInfo: &PageInfo{
					Next:    cursor.New(10).AddField("id", 7, common.DirectionAsc).Encode(),
					Prev:    cursor.New(10).AddField("id", 1, common.DirectionAsc).SetBackward().Encode(),
					HasNext: false, HasPrev: false, TotalRows: 7,
				},
			},
		},
		{
			Name: "Field with custom cursor name (sorting query)",
			Params: q{
				{
					"sorting": `[
					{
						"field": "item_id_cursor",
						"direction": "asc"
					}
				]`,
					"limit": "4",
				},
			},
			Result: r{
				IDs: []uint{1, 5, 3, 4},
				PageInfo: &PageInfo{
					Next:    cursor.New(4).AddField("item_id", "a4", common.DirectionAsc).AddField("id", 4, common.DirectionAsc).Encode(),
					Prev:    cursor.New(4).AddField("item_id", "a1", common.DirectionAsc).AddField("id", 1, common.DirectionAsc).SetBackward().Encode(),
					HasNext: true, HasPrev: false, TotalRows: 7,
				},
			},
		},
		{
			Name: "Field with custom cursor name (cursor query: page 2)",
			Params: q{
				{
					"cursor": cursor.New(4).AddField("item_id", "a4", common.DirectionAsc).AddField("id", 4, common.DirectionAsc).Encode(),
				},
			},
			Result: r{
				IDs: []uint{6, 7, 2},
				PageInfo: &PageInfo{
					Next:    cursor.New(4).AddField("item_id", "c1", common.DirectionAsc).AddField("id", 2, common.DirectionAsc).Encode(),
					Prev:    cursor.New(4).AddField("item_id", "b1", common.DirectionAsc).AddField("id", 6, common.DirectionAsc).SetBackward().Encode(),
					HasNext: false, HasPrev: true, TotalRows: 7,
				},
			},
		},
		{
			Name: "Field from embedded struct: author.name (sorting query)",
			Params: q{
				{
					"sorting": `[
					{
						"field": "author.name",
						"direction": "asc"
					}
				]`,
				},
			},
			Result: r{
				IDs: []uint{2, 4},
				PageInfo: &PageInfo{
					Next:    cursor.New(pageLimit).AddField(`"Author__name"`, "A", common.DirectionAsc).AddField("id", 4, common.DirectionAsc).Encode(),
					Prev:    cursor.New(pageLimit).AddField(`"Author__name"`, "A", common.DirectionAsc).AddField("id", 2, common.DirectionAsc).SetBackward().Encode(),
					HasNext: true, HasPrev: false, TotalRows: 7,
				},
			},
		},
		{ // 8
			Name: "Field from subquery: claps (sorting query)",
			Params: q{
				{
					"sorting": `[
					{
						"field": "claps",
						"direction": "desc"
					}
				]`,
				},
			},
			Result: r{
				IDs: []uint{6, 1},
				PageInfo: &PageInfo{
					Next:    cursor.New(pageLimit).AddField(`claps`, 1, common.DirectionDesc).AddField("id", 1, common.DirectionAsc).Encode(),
					Prev:    cursor.New(pageLimit).AddField(`claps`, 2, common.DirectionDesc).AddField("id", 6, common.DirectionAsc).SetBackward().Encode(),
					HasNext: true, HasPrev: false, TotalRows: 7,
				},
			},
		},
		{
			Name: "Field from subquery: claps (cursor query: page 2)",
			Params: q{
				{
					"cursor": cursor.New(pageLimit).AddField(`claps`, 1, common.DirectionDesc).AddField("id", 2, common.DirectionAsc).Encode(),
				},
			},
			Result: r{
				IDs: []uint{3, 7},
				PageInfo: &PageInfo{
					Next:    cursor.New(pageLimit).AddField(`claps`, 1, common.DirectionDesc).AddField("id", 7, common.DirectionAsc).Encode(),
					Prev:    cursor.New(pageLimit).AddField(`claps`, 1, common.DirectionDesc).AddField("id", 3, common.DirectionAsc).SetBackward().Encode(),
					HasNext: true, HasPrev: true, TotalRows: 7,
				},
			},
		},
		{
			Name: "Embedded and subquery: claps (sorting query)",
			Params: q{
				{
					"sorting": `[
					{
						"field": "author.id",
						"direction": "asc"
					},
					{
						"field": "claps",
						"direction": "desc"
					},
					{
						"field": "id",
						"direction": "desc"
					}
				]`,
					"limit": "4",
				},
			},
			Result: r{
				IDs: []uint{2, 4, 6, 3},
				PageInfo: &PageInfo{
					Next:    cursor.New(4).AddField(`"Author__id"`, 3, common.DirectionAsc).AddField(`claps`, 1, common.DirectionDesc).AddField("id", 3, common.DirectionDesc).Encode(),
					Prev:    cursor.New(4).AddField(`"Author__id"`, 1, common.DirectionAsc).AddField(`claps`, 1, common.DirectionDesc).AddField("id", 2, common.DirectionDesc).SetBackward().Encode(),
					HasNext: true, HasPrev: false, TotalRows: 7,
				},
			},
		},
		{
			Name: "Time field: json",
			Params: q{
				{
					"sorting": `[
					{
						"field": "updated_at",
						"direction": "desc"
					}
				]`,
					"limit": "4",
				},
			},
			Result: r{
				IDs: []uint{1, 2, 3, 4},
				PageInfo: &PageInfo{
					Next:    cursor.New(4).AddField(`updated_at`, "2020-12-31 23:59:59+06", common.DirectionDesc).AddField("id", 3, common.DirectionDesc).Encode(),
					Prev:    cursor.New(4).AddField(`updated_at`, "2020-12-31 23:56:59+06", common.DirectionDesc).AddField("id", 2, common.DirectionDesc).SetBackward().Encode(),
					HasNext: true, HasPrev: false, TotalRows: 7,
				},
			},
		},
		//""
	}

	runList := true
	oneTestCase := func(ind int) {
		if ind == -1 {
			return
		}

		debug = true
		runList = false

		tc := tCases[ind]
		w := performRequest(router, "GET", "/list", tc.Params)

		var response materialListResponse
		err := json.Unmarshal([]byte(w.Body.String()), &response)
		if err != nil {
			log.Println(err)
		}

		log.Printf("%+v\n", response)
		if ok, err := checkResult(&response, &tc.Result); !ok {
			t.Errorf("%s. Fail: %v\n", tc.Name, err)
		}
	}

	listTestCases := func(run bool) {
		if !run {
			return
		}
		for i, tc := range tCases {
			w := performRequest(router, "GET", "/list", tc.Params)

			var response materialListResponse
			err := json.Unmarshal([]byte(w.Body.String()), &response)
			if err != nil {
				log.Println(err)
			}

			log.Printf("%+v\n", response)
			if ok, err := checkResult(&response, &tc.Result); !ok {
				t.Errorf("%v) %s. Fail: %v\n\n\n", i, tc.Name, err)
			}
		}
	}

	oneTestCase(11)
	listTestCases(runList)
	// assert.Equal(t, response["result"], true)
}

func performRequest(r http.Handler, method, path string, query []map[string]string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)

	for _, q := range query {
		if len(q) > 0 {
			query := req.URL.Query()
			for k, v := range q {
				query.Add(k, v)
				log.Println(v)
			}
			req.URL.RawQuery = query.Encode()
			log.Println("RawQuery:", req.URL.RawQuery)
		}
	}

	log.Println("Request url:", req.URL.String())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func checkResult(response *materialListResponse, result *r) (ok bool, err error) {
	if response == nil && result == nil {
		return true, nil
	}

	if response == nil && result != nil {
		return false, fmt.Errorf("response == nil && result != nil")
	}

	if response != nil && result == nil {
		return false, fmt.Errorf("response != nil && result == nil")
	}

	if ok, err := comparePageInfo(response.Paging, result.PageInfo); !ok {
		return false, fmt.Errorf("Not equals PageInfo:\n %#v\n", err)
	}

	if ok, mIDs := compareIDs(result.IDs, response.Materials); !ok {
		return false, fmt.Errorf("Not equals results IDs:\nExpected: %#v\n  Actual: %#v", result.IDs, mIDs)
	}

	return true, nil
}

func comparePageInfo(a, b *PageInfo) (ok bool, err error) {
	if notComparePageInfo {
		return true, nil
	}
	if a == nil && b == nil {
		return true, nil
	}

	if a == nil && b != nil {
		return false, fmt.Errorf("a == nil BUT b != nil")
	}
	if a != nil && b == nil {
		return false, fmt.Errorf("a != nil BUT b == nil")
	}

	switch {
	case a.Next != b.Next:
		return false, fmt.Errorf("Not equal Next: %v AND %v", a.Next, b.Next)
	case a.Prev != b.Prev:
		return false, fmt.Errorf("Not equal Prev: %v AND %v", a.Prev, b.Prev)
	case a.HasNext != b.HasNext:
		return false, fmt.Errorf("Not equal HasNext: %v AND %v", a.HasNext, b.HasNext)
	case a.HasPrev != b.HasPrev:
		return false, fmt.Errorf("Not equal HasPrev: %v AND %v", a.HasPrev, b.HasPrev)
	case a.TotalRows != b.TotalRows:
		return false, fmt.Errorf("Not equal TotalRows: %v AND %v", a.TotalRows, b.TotalRows)
	}

	return true, nil
}

func compareIDs(IDs []uint, materials []Material) (ok bool, materialIDs []uint) {
	mIDs := materialsToResultIDs(materials)

	for i := range mIDs {
		if mIDs[i] != IDs[i] {
			return false, mIDs
		}
	}
	return true, mIDs
}

func materialsToResultIDs(materials []Material) (IDs []uint) {
	IDs = make([]uint, len(materials), len(materials))
	for i, m := range materials {
		IDs[i] = m.ID
	}

	return IDs
}

// ------------ Code example
// main
func SetupRouter() *gin.Engine {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.GET("/list", List)

	return router
}

// controller
type (
	materialListResponse struct {
		Result    bool       `json:"result"`
		Materials []Material `json:"materials"`
		Paging    *PageInfo  `json:"paging"`
	}
)

func List(c *gin.Context) {
	var limit int
	limit, _ = strconv.Atoi(c.Query("limit"))
	defaultLimit := 2
	if limit <= 0 {
		limit = defaultLimit
	} else {
		log.Println("!!! Set limit:", limit)
	}

	db := liveDB()

	paginator, err := New(Options{
		GinContext: c,
		Limit:      uint(limit),
		DB:         db,
		Model:      &Material{},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, `{"result":false}`)
	}
	data := GetList(paginator)
	if data == nil || len(data) == 0 {
		// c.JSON(http.StatusBadRequest, cm.Error[cm.ErrItemNotFound])
		c.JSON(http.StatusOK, materialListResponse{
			Result:    true,
			Materials: data,
			Paging:    nil,
		})
		return
	}

	log.Printf("%#v\n", paginator.PageInfo)

	c.JSON(http.StatusOK, materialListResponse{
		Result:    true,
		Materials: data,
		Paging:    paginator.PageInfo,
	})
}

// model
type (
	BaseModel struct {
		ID        uint `gorm:"primary_key"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	BaseModelWithSoftDelete struct {
		BaseModel
		DeletedAt gorm.DeletedAt `gorm:"index"`
	}

	// User is the user model of the mobile application.
	User struct {
		BaseModelWithSoftDelete
		Role     uint `gorm:"not null" sql:"DEFAULT:0"`
		AuthType uint `gorm:"not null" sql:"DEFAULT:0"`
		AuthID   string
		Name     string
		Photo    string
	}

	// Users list
	Users []User

	Material struct {
		ID        uint `gorm:"primary_key"`
		CreatedAt time.Time
		UpdatedAt time.Time      `json:"updated_at"`
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Link    string
		Status  Status
		Comment string

		ItemID      string `cursor:"item_id_cursor"` // Unical +
		ItemOwnerID int    // Unical
		ItemType    string

		Claps       int64 `sql:"-" gorm:"->"` // calculate
		FailedClaps int64 `sql:"-" gorm:"->"` // calc

		UserID     uint
		Author     User `gorm:"foreignKey:UserID"`
		LikesCount uint
	}

	Materials []Material

	Status uint

	Clap struct {
		MaterialID uint `gorm:"primary_key"`
		ClapperID  uint `gorm:"primary_key"`
		ClapAt     time.Time
		Success    bool
	}
)

// GetList return all materials
func GetList(paginator *Paginator) (materials Materials) {
	// db := mockDB()
	db := liveDB()
	q := db.Table("(?) as tabl", db.Model(&Material{}).
		Select(`materials.*,
			(select count(1) from claps where claps.material_id = materials.id and claps.success = true) as claps,
			(select count(1) from claps where claps.material_id = materials.id and claps.success = false) as failed_claps
			`).
		Joins("Author"))

	paginator.Find(q, &materials)
	return
}

// DB connection
var gormConf = &gorm.Config{
	PrepareStmt: true,
}

func mockDB() *gorm.DB {
	sqlDB, _, _ := sqlmock.New()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), gormConf)
	if err != nil {
		log.Println(err)
		return nil
	}
	return db
}

func liveDB() *gorm.DB {
	connString := "host=localhost port=5432 user=postgres dbname=pagination password=123 sslmode=disable"
	db, err := gorm.Open(postgres.Open(connString), gormConf)
	if err != nil {
		log.Println(err)
		return nil
	}

	// db.AutoMigrate(&User{}, &Material{}, &Clap{})

	sqlDB, _ := db.DB()
	fixtures, err = testfixtures.New(
		testfixtures.DangerousSkipTestDatabaseCheck(),
		testfixtures.Database(sqlDB),       // You database connection
		testfixtures.Dialect("postgres"),   // Available: "postgresql", "timescaledb", "mysql", "mariadb", "sqlite" and "sqlserver"
		testfixtures.Directory("testdata"), // the directory containing the YAML files
	)
	if err != nil {
		log.Fatal(err)
	}
	prepareTestDatabase()

	if debug {
		return db.Debug()
	}
	return db
}

func prepareTestDatabase() {
	if err := fixtures.Load(); err != nil {
		log.Printf("prepareTestDatabase: %v", err)
	}
}
