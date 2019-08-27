package ginx

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/obase/center"
	"github.com/obase/httpx/cache"
	"net/http"
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
	defer cache.Close()

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
	s.Run(entry, defargs, cache, ":8080")
}
