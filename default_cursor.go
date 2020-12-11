package pagination

type (
	DefaultCursorGetter func() *Cursor

	Pagination struct {
	}
)

func (p *Pagination) DefaultCursor() *Cursor {
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
