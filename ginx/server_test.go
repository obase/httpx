package ginx

import (
	"github.com/obase/httpx/cache"
	"testing"
)

func TestNew(t *testing.T) {

	config := &Config{
		HttpCache: &cache.Config{
			Type: cache.MEMORY,
		},
		HttpPlugin: map[string]string{
			"VerifyToken": "a,b,c,d",
		},
		HttpEntry: []*Entry{
			&Entry{
				Source:  "/now",
				Service: "target",
				Target:  "/now",
				Plugin:  []string{"demo"},
				Cache:   5,
			},
		},
	}

}
