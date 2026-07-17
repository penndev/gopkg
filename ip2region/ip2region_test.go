package ip2region_test

import (
	"fmt"

	"github.com/penndev/gopkg/ip2region"
)

func ExampleFind() {
	fmt.Println(ip2region.Find("119.29.29.29"))
	fmt.Println(ip2region.Find("2001:4860:4860::8844"))
	// Output:
	// {美国    Google LLC}
}
