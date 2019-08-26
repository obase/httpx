package ginx

import (
	"github.com/gin-gonic/gin"
	"testing"
)

func TestNew(t *testing.T) {
	g := gin.New()
	g.RunUnix()
}
