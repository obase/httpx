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
			Source:  "/g/now",
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

	g := s.Group("/g", func(context *gin.Context) {
		fmt.Println("this is in group...")
	})

	g.GET("/now", func(context *gin.Context) {
		fmt.Printf("this is in now ...\n")
		context.JSON(http.StatusOK, "more good to it...")
	})

	if err := s.Run(entry, defargs, cache, ":8080"); err != nil {
		fmt.Println(err)
	}
}
