package ginx

import "net/http"

type Entry struct {
	Method  []string `json:"method" bson:"method" yaml:"method"`    // 请求方法(可选,默认)
	Source  string   `json:"source" bson:"source" yaml:"source"`    // 来源URI(必需且惟一)
	Https   bool     `json:"https" bson:"https" yaml:"https"`       // 是否使用tls
	Service string   `json:"service" bson:"service" yaml:"service"` // 目标服务(必需)
	Target  string   `json:"target" bson:"target" yaml:"target"`    // 目标URI(必需)
	Plugin  []string `json:"plugin" bson:"plugin" yaml:"plugin"`    // 服务插件(可选)
	Cache   int64    `json:"cache" bson:"cache" yaml:"cache"`       // 缓存时间(秒)
	Remark  string   `json:"remark" bson:"remark" yaml:"remark"`    // 备注描述(可选)
}

var DefaultMethods = []string{http.MethodGet, http.MethodPost}

func MergeEntry(entries []Entry) []Entry {
	for _, entry := range entries {
		if len(entry.Method) == 0 {
			entry.Method = DefaultMethods
		}
	}
	return entries
}
