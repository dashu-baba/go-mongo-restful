package common

//PagedList godoc
// @Summary The PagedList entity.
type PagedList struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}
