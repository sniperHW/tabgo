package test

type ModelStructY struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ModelStruct struct {
	X     int          `json:"x"`
	Y     ModelStructY `json:"y"`
	Array []int        `json:"array"`
}

type ModelArray_struct struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Model struct {
	Id           int                 `json:"id"`
	Name         string              `json:"name"`
	Icon         string              `json:"icon"`
	Model        string              `json:"model"`
	Length       int                 `json:"length"`
	Width        int                 `json:"width"`
	Struct       ModelStruct         `json:"struct"`
	Array        []int               `json:"array"`
	Array2d      [][]int             `json:"array2d"`
	Array_struct []ModelArray_struct `json:"array_struct"`
}
