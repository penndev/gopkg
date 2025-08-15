package ip2region

type RegionMap struct {
	Name     string      `json:"name"`
	Children []RegionMap `json:"children,omitempty"`
}

var Region []RegionMap
