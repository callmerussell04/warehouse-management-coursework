package dto

type Paging struct {
	Page  int `json:"page"`
	Size  int `json:"size"`
	Total int `json:"total"`
}
