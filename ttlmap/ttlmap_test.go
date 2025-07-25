package ttlmap_test

import (
	"testing"
	"time"

	"github.com/penndev/gopkg/ttlmap"
)

func TestNew(t *testing.T) {
	tm := ttlmap.New()
	tm.Set("penn", "penndev", 3*time.Second)
	t.Log(tm.Get("penn"))
	time.Sleep(2 * time.Second)
	t.Log(tm.Load("penn"))
	t.Log("-------------------------------------")
	tm.Set("penn0", "penndev", 0)
	t.Log(tm.Get("penn0"))
	t.Fail()
}
