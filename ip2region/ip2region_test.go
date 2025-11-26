package ip2region_test

import (
	"fmt"

	"github.com/penndev/gopkg/ip2region"
)

func ExampleFind() {
	fmt.Println(ip2region.Find("223.5.5.5"))
	// Output:
	// {中国 浙江 杭州  阿里巴巴anycast公共DNS}
}
