package test

type ModelStructY struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type ModelStruct struct {
	X     int64        `json:"x"`
	Y     ModelStructY `json:"y"`
	Array []int64      `json:"array"`
}

type ModelArray_struct struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type Model struct {
	Id           int64               `json:"id"`
	Name         string              `json:"name"`
	Icon         string              `json:"icon"`
	Model        string              `json:"model"`
	Length       int64               `json:"length"`
	Width        int64               `json:"width"`
	Struct       ModelStruct         `json:"struct"`
	Array        []int64             `json:"array"`
	Array2d      [][]int64           `json:"array2d"`
	Array_struct []ModelArray_struct `json:"array_struct"`
}
