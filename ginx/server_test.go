package ginx

import (
	"github.com/obase/center"
	"github.com/obase/httpx/cache"
	"testing"
)

func TestNew(t *testing.T) {
	center.Setup(&center.Config{
		Address: "10.11.165.44:18500",
	})
	entry := []Entry{
		Entry{
			Source:  "/now",
			Service: "target",
			Target:  "/now",
			Plugin:  []string{"demo"},
			Cache:   5,
		}}
	defargs := map[string]string{
		"demo": "a,b,c,d",
	}
	cache := cache.New(&cache.Config{
		Type: cache.MEMORY,
	})

	s := New()
	s.Run(entry, defargs, cache, ":8080")
}
