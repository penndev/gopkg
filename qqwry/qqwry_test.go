package qqwry

import (
	"log"
	"testing"
)

func TestMain(t *testing.T) {
	r := Find("2402:4e00::")
	log.Println(r.BeginIP, r.EndIP, r.Country, r.Area)
	if r.Area == "" {
		t.Fail()
	}
}
