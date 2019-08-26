package ginx

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/obase/httpx/cache"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {

	config := &Config{
		HttpCache: &cache.Config{
			Type: cache.MEMORY,
		},
		HttpPlugin: map[string]string{
			"demo": "a,b,c,d",
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

	s := New()
	s.Plugin("demo", func(args []string) gin.HandlerFunc {
		return func(context *gin.Context) {
			fmt.Println(args)
			flag := context.Query("flag")
			if flag == "abort" {
				context.AbortWithStatus(http.StatusOK)
			}
		}
	})
	s.Run(config, ":8080")
}
