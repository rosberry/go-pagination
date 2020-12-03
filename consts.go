package pagination

const (
	DirectionDesc DirectionType = "desc"
	DirectionAsc  DirectionType = "asc"
)

var CompareTerms map[DirectionType]string = map[DirectionType]string{
	DirectionDesc: "<",
	DirectionAsc:  ">",
}

var DirectionByString map[string]DirectionType = map[string]DirectionType{
	"asc":  DirectionAsc,
	"desc": DirectionDesc,
}

const (
	defaultLimit = 3
)
