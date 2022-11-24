package test

type ModelStructTestY struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ModelStructTest struct {
	X int               `json:"x"`
	Y *ModelStructTestY `json:"y"`
}

type Model struct {
	Id     int              `json:"id"`
	Name   string           `json:"name"`
	Icon   string           `json:"icon"`
	Model  string           `json:"model"`
	Length int              `json:"length"`
	Width  int              `json:"width"`
	Struct *ModelStructTest `json:"struct"`
}
