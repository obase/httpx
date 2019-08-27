package ginx

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

func joinPath(v1 string, v2 string) string {
	return v1 + v2
}

func joinChain(v1 gin.HandlersChain, v2 gin.HandlersChain) gin.HandlersChain {
	n1 := len(v1)
	n2 := len(v2)
	if n := n1 + n2; n > 0 {
		ret := make([]gin.HandlerFunc, 0, n)
		if n1 > 0 {
			ret = append(ret, v1...)
		}
		if n2 > 0 {
			ret = append(ret, v2...)
		}
		return ret
	}
	return nil
}

func toStringSlice(val interface{}) []string {
	switch val := val.(type) {
	case nil:
		return nil
	case []interface{}:
		ret := make([]string, len(val))
		for i, v := range val {
			ret[i] = toString(v)
		}
		return ret
	case string:
		return strings.Split(val, ",")
	}
	panic(fmt.Sprintf("invalid value to strSlice: %v", val))
}

func toString(val interface{}) string {
	switch val := val.(type) {
	case nil:
		return ""
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}
