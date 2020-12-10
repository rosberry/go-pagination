package pagination

import "log"

type (
	DefaultCursorGetter func() *Cursor

	Pagination struct {
	}
)

func (p *Pagination) DefaultCursor() *Cursor {
	log.Println("Return default cursor")
	return &Cursor{
		Fields: []Field{
			Field{
				Name:      "id",
				Value:     nil,
				Direction: DirectionAsc,
			},
		},
		Limit:    defaultLimit,
		Backward: false,
	}
}

func DefaultCursor() *Cursor {
	return (&Pagination{}).DefaultCursor()
}
