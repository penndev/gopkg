package db

import "fmt"

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

// CheckArea 按字典顺序逐条校验 Area；prevID 为上一条的 ID（首条传 0）。
// 约束：ID 严格递增（因此不重复、不为 0）；ParentID 为 0 或小于 ID（因此父链不成环）。
func CheckArea(prevID uint32, a Area) error {
	if a.ID <= prevID {
		return fmt.Errorf("area id %d 未严格递增（上一条 %d）", a.ID, prevID)
	}
	if a.ParentID >= a.ID {
		return fmt.Errorf("area id %d 的 parent_id %d 不小于自身", a.ID, a.ParentID)
	}
	return nil
}

// CheckISP 按字典顺序逐条校验 ISP；约束：ID 严格递增。
func CheckISP(prevID uint32, isp ISP) error {
	if isp.ID <= prevID {
		return fmt.Errorf("isp id %d 未严格递增（上一条 %d）", isp.ID, prevID)
	}
	return nil
}
