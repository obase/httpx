package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/obase/ginx/httpcache"
	"net/http"
)

type builder struct {
	*Config
	indexes map[string]map[string]*Entry
}

func newEngineBuilder(config *Config) *builder {
	cc := &builder{
		Config: config,
	}
	cc.indexentry()
	return cc
}

var default_methos = []string{http.MethodGet, http.MethodPost}

func (cc *builder) indexentry() {
	if len(cc.Config.HttpEntry) > 0 {
		cc.indexes = make(map[string]map[string]*Entry)
		for _, entry := range cc.Config.HttpEntry {
			ms := entry.Method
			if len(ms) == 0 {
				ms = default_methos
			}
			for _, m := range ms {
				mentry, ok := cc.indexes[m]
				if !ok {
					mentry := make(map[string]*Entry)
					cc.indexes[m] = mentry
				}
				mentry[entry.Source] = entry
			}
		}
	}
}

func (cc *builder) dlinkentry(method string, path string) *Entry {
	mentry, ok := cc.indexes[method]
	if !ok {
		return nil
	}
	entry, ok := mentry[path]
	if !ok {
		return nil
	}
	delete(mentry, path)
	return entry
}

// 递归处理
func (cc *builder) building(prefix string, route gin.IRouter, node *RouteNode, plugins []*Plugin, cache httpcache.Cache) {

	if node.path != "" {
		prefix += node.path
		route = route.Group(node.path, node.use...)
	}

	for method, routes := range node.handle {
		for p, c := range routes {
			path := joinPath(prefix, p)
			entry := cc.dlinkentry(method, path)
			if entry != nil {
				// 需要执行特殊设置

			} else {
				// 没有plugin,cache等特殊设置
				route.Handle(method, p, c)
			}
		}
	}
	for path, file := range node.staticFile {
		route.StaticFile(path, file)
	}
	for path, file := range node.static {
		route.Static(path, file)
	}
	for path, fs := range node.staticFS {
		route.StaticFS(path, fs)
	}

	// 递归子树
	for _, cnode := range node.child {
		cc.building(prefix, route, cnode, plugins, cache)
	}
}

// 删除掉后只剩下反射代理的入口
func (cc *builder) proxying(engine *gin.Engine) {

}

//  最后检测入口是否清零,否则必有哪个地方配置问题
func (cc *builder) checking() error {
	return nil
}

func (cc *builder) dispose() {
	cc.Config = nil
	cc.indexes = nil
}
