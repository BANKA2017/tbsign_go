package _type

type TypeLastDo struct {
	Num    int64  `php:"num"`
	LastDo string `php:"lastdo"`
}

type TypeCookie struct {
	Bduss   string
	Stoken  string
	Tbs     string
	IsLogin bool

	ID       int32
	UID      int32
	Name     string
	Portrait string
}
