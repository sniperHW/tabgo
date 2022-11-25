package test

type ModelStructY struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ModelStruct struct {
	X int `json:"x"`
	Y *ModelStructY `json:"y"`
}

type Model struct {
	Id int `json:"id"` 
	Name string `json:"name"` 
	Length int `json:"length"` 
	Width int `json:"width"` 
	Struct *ModelStruct `json:"struct"` 
	Array []int `json:"array"` 
}

