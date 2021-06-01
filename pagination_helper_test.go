package pagination_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-testfixtures/testfixtures/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/rosberry/go-pagination"
)

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
		PublicAt  *time.Time     `json:"PublicTime"`

		Link    string
		Status  Status
		Comment string

		ItemID      string `cursor:"item_id_cursor"` // Unical +
		ItemOwnerID int    // Unical
		ItemType    string

		UserID        uint
		Author        User `gorm:"foreignKey:UserID"`
		AuthorPreload User `gorm:"foreignKey:UserID"`
		LikesCount    uint
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

var fixtures *testfixtures.Loader

// DB connection
var gormConf = &gorm.Config{
	PrepareStmt: true,
}

// GetList return all materials
func GetList(paginator *pagination.Paginator) (materials Materials) {
	// db := mockDB()
	db := liveDB()
	q := db.Model(&Material{}).
		Preload("AuthorPreload").
		Select(`materials.*,
			(select count(1) from claps where claps.material_id = materials.id and claps.success = true) as claps,
			(select count(1) from claps where claps.material_id = materials.id and claps.success = false) as failed_claps
			`).
		Joins("Author")

	paginator.Find(q, &materials)
	return
}

func liveDB() *gorm.DB {
	connString := os.Getenv("DB_CONNECT_STRING") //	"host=localhost port=5432 user=postgres dbname=pagination password=123 sslmode=disable"
	if connString == "" {
		log.Print("Use DB_CONNECT_STRING env for setup db connection string")
		os.Exit(1)
	}

	db, err := gorm.Open(postgres.Open(connString), gormConf)
	if err != nil {
		log.Println(err)
		return nil
	}

	db.AutoMigrate(&User{}, &Material{}, &Clap{})

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

	return db
}

func prepareTestDatabase() {
	if err := fixtures.Load(); err != nil {
		log.Printf("prepareTestDatabase: %v", err)
	}
}

// controller
type (
	materialListResponse struct {
		Result    bool                 `json:"result"`
		Materials []Material           `json:"materials"`
		Paging    *pagination.PageInfo `json:"paging"`
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

	paginator, err := pagination.New(pagination.Options{
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

func SetupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.GET("/list", List)

	return router
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

func comparePageInfo(a, b *pagination.PageInfo) (ok bool, err error) {
	notComparePageInfo := false

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
