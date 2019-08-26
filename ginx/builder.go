package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/obase/ginx/httpcache"
	"github.com/obase/httpx"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httputil"
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
					mentry = make(map[string]*Entry)
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
func (cc *builder) buildRoute(route gin.IRouter, plugins []*Plugin, cache httpcache.Cache, prefix string, node *RouteNode) error {

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
				if entry.Service != "" || entry.Target != "" {
					return errors.New("conflict http entry: " + method + " " + path)
				}
				var handlers []gin.HandlerFunc
				// 处理plugin
				for _, express := range (*entry).Plugin {
					if handler, err := EvaluePluginExpress(plugins, cc.Config.HttpPlugin, express); err != nil {
						return err
					} else {
						handlers = append(handlers, handler)
					}
				}
				// 处理cache
				if entry.Cache > 0 {
					handlers = append(handlers, cache.Cache(entry.Cache, c))
				} else {
					handlers = append(handlers, c)
				}
				// 添加路由
				route.Handle(method, p, handlers...)
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
		err := cc.buildRoute(route, plugins, cache, prefix, cnode)
		if err != nil {
			return err
		}
	}
	return nil
}

// 删除掉后只剩下反射代理的入口
func (cc *builder) buildProxy(route gin.IRouter, plugins []*Plugin, cache httpcache.Cache) error {
	for method, entries := range cc.indexes {
		for path, entry := range entries {
			if entry.Service == "" || entry.Target == "" {
				return errors.New("invalid proxy entry: " + method + " " + path)
			}
			var handlers []gin.HandlerFunc
			// 处理plugin
			for _, express := range (*entry).Plugin {
				if handler, err := EvaluePluginExpress(plugins, cc.Config.HttpPlugin, express); err != nil {
					return err
				} else {
					handlers = append(handlers, handler)
				}
			}
			// 处理cache
			c := proxyHandlerFunc(entry.Https, entry.Service, entry.Target)
			if entry.Cache > 0 {
				handlers = append(handlers, cache.Cache(entry.Cache, c))
			} else {
				handlers = append(handlers, c)
			}
			// 添加路由
			route.Handle(method, path, handlers...)
		}
	}
	return nil
}

func (cc *builder) dispose() {
	cc.Config = nil
	cc.indexes = nil
}

func proxyHandlerFunc(https bool, serviceName string, uri string) gin.HandlerFunc {
	var proxy *httputil.ReverseProxy
	if https {
		proxy = httpx.ProxyHandlerTLS(serviceName, uri)
	} else {
		proxy = httpx.ProxyHandler(serviceName, uri)
	}
	return func(context *gin.Context) {
		proxy.ServeHTTP(context.Writer, context.Request)
	}
}
