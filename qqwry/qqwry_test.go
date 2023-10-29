package qqwry

import (
	"log"
	"testing"
)

func TestMain(t *testing.T) {
	r := Find("192.168.7.1")
	log.Println(r.BeginIP, r.EndIP, r.Country, r.Area)
	if r.Area == "" {
		t.Fail()
	}
}
