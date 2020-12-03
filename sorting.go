package pagination

import "log"

type (
	sortingElem struct {
		Field     string `json:"field" form:"field"`
		Direction string `json:"direction" form:"direction"`
	}

	sorting []sortingElem
)

func (srt *sorting) toCursor() *Cursor {
	if srt == nil {
		return nil
	}
	log.Printf("%+v\n", srt)
	log.Println("Func sorting.toCursor() not implement!")
	cursor := &Cursor{
		Limit:    defaultLimit,
		Backward: false,
	}

	for _, e := range *srt {
		direction, ok := DirectionByString[e.Direction]
		if !ok {
			direction = DirectionAsc
		}
		cursor.AddField(e.Field, nil, direction)
	}
	return cursor
}
