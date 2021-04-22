package pagination_test

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/rosberry/go-pagination"
	"github.com/rosberry/go-pagination/common"
	"github.com/rosberry/go-pagination/cursor"
)

type (
	q = []map[string]string
	r struct {
		IDs      []uint
		PageInfo *pagination.PageInfo
	}

	tCase struct {
		Name   string
		Params []map[string]string
		Result r
	}
)

var pageLimit = 2

func TestPreload(t *testing.T) {
	router := SetupRouter()

	tCases := []tCase{
		{
			Name:   "Default query",
			Params: q{},
			Result: r{
				IDs: []uint{1, 2},
				PageInfo: &pagination.PageInfo{
					Next:      cursor.New(pageLimit).AddField("id", 2, common.DirectionAsc).Encode(),
					Prev:      cursor.New(pageLimit).AddField("id", 1, common.DirectionAsc).SetBackward().Encode(),
					HasNext:   true,
					HasPrev:   false,
					TotalRows: 7,
				},
			},
		},
	}

	for i, tc := range tCases {
		w := performRequest(router, "GET", "/list", tc.Params)

		var response materialListResponse
		err := json.Unmarshal([]byte(w.Body.String()), &response)
		if err != nil {
			log.Println(err)
		}

		// log.Printf("%+v\n", response)
		if ok, err := checkResult(&response, &tc.Result); !ok {
			t.Errorf("%v) %s. Fail: %v\n\n\n", i, tc.Name, err)
		}
		for _, m := range response.Materials {
			if m.Author.ID == 0 {
				t.Error("Fail: Author not preload")
			}
			if m.AuthorPreload.ID == 0 {
				t.Error("Fail: AuthorPreload not preload")
			}
		}
	}
}
