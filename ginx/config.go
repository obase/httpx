package ginx

import (
	"github.com/obase/ginx/httpcache"
)

type Config struct {
	HttpCache  *httpcache.Config `json:"httpCache" bson:"httpCache" yaml:"httpCache"`    // 是否启用Redis缓存
	HttpPlugin map[string]string `json:"httpPlugin" bson:"httpPlugin" yaml:"httpPlugin"` // 默认参数
	HttpEntry  []*Entry          `json:"httpEntry" bson:"httpEntry" yaml:"httpEntry"`    // 代理入口配置
}
