package ginx

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"testing"
)

func TestNew(t *testing.T) {
	s := server.New()
	s.Use(func(c *gin.Context) {
		fmt.Println("this is server.use")
	})
	s.GET("/test", func(c *gin.Context) {
		fmt.Println("this is test")
	})
	g := s.Group("/group", func(c *gin.Context) {
		fmt.Println("this is group")
	})
	g.Use(func(c *gin.Context) {
		fmt.Println("this is group use")
	})
	fmt.Println(s.child)
}
