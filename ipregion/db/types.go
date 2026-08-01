package db

// Area / ISP 字典类型（名称变长；中间产物与库内字典区默认 JSON）。

type Area struct {
	ID       uint32 `json:"id"`
	ParentID uint32 `json:"parent_id"`
	Name     string `json:"name"`
}

type ISP struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}
