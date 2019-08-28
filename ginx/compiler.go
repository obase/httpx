package ginx

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/obase/httpx"
	"github.com/obase/httpx/cache"
	"net/http/httputil"
)

type EngineCompiler struct {
	Entries map[string]map[string]*Entry
	Plugins []Plugin
	Defargs map[string]string
	Cache   cache.Cache
}

func NewEngineCompiler(httpEntry []Entry, plugins []Plugin, defargs map[string]string, cache cache.Cache) *EngineCompiler {
	return &EngineCompiler{
		Entries: _index(httpEntry),
		Plugins: plugins,
		Defargs: defargs,
		Cache:   cache,
	}
}

func (c *EngineCompiler) Compile(root *RouteNode) (*gin.Engine, error) {
	ge := gin.New()

	if err := c.compileRoute(ge, "", root); err != nil {
		return nil, err
	}
	if err := c.compileProxy(ge); err != nil {
		return nil, err
	}

	return ge, nil
}

// 递归处理. 必须处理compiler里面entries, plugins, defargs, cache等为空的情况. 确保能正常处理
func (compiler *EngineCompiler) compileRoute(router gin.IRouter, prefix string, node *RouteNode) error {
	if node.path != "" {
		prefix += node.path
		router = router.Group(node.path, node.use...)
	}
	for method, routes := range node.handle {
		for p, h := range routes {
			path := joinPath(prefix, p)
			entry := _dlink(compiler.Entries, method, path)
			if entry != nil {
				// 需要执行特殊设置
				if entry.Service != "" || entry.Target != "" {
					return fmt.Errorf("conflict entry: method=%v, source=%v, service=%v, target=%v, handle=%p", method, path, entry.Service, entry.Target, h)
				}
				var handlers []gin.HandlerFunc
				if len(compiler.Plugins) > 0 {
					// 处理plugin
					for _, express := range (*entry).Plugin {
						if handler, err := EvaluePluginExpress(compiler.Plugins, compiler.Defargs, express); err != nil {
							return err
						} else if handler != nil {
							handlers = append(handlers, handler)
						}
					}
				}
				// 处理cache
				if compiler.Cache != nil && entry.Cache > 0 {
					handlers = append(handlers, compiler.Cache.Cache(entry.Cache, h))
				} else {
					handlers = append(handlers, h...)
				}
				// 添加路由
				router.Handle(method, p, handlers...)
			} else {
				// 没有plugin,cache等特殊设置
				router.Handle(method, p, h)
			}
		}
	}
	for path, file := range node.staticFile {
		router.StaticFile(path, file)
	}
	for path, file := range node.static {
		router.Static(path, file)
	}
	for path, fs := range node.staticFS {
		router.StaticFS(path, fs)
	}

	// 递归子树
	for _, cnode := range node.child {
		if err := compiler.compileRoute(router, prefix, cnode); err != nil {
			return err
		}
	}
	return nil
}

func (compiler *EngineCompiler) compileProxy(router gin.IRouter) error {
	for method, entrymap := range compiler.Entries {
		for path, entry := range entrymap {
			if entry.Service == "" || entry.Target == "" {
				return fmt.Errorf("invalid entry: method=%v, source=%v, service=%v, target=%v, handle=<nil>", method, path, entry.Service, entry.Target)
			}
			var handlers []gin.HandlerFunc
			if len(compiler.Plugins) > 0 {
				// 处理plugin
				for _, express := range (*entry).Plugin {
					if handler, err := EvaluePluginExpress(compiler.Plugins, compiler.Defargs, express); err != nil {
						return err
					} else if handler != nil {
						handlers = append(handlers, handler)
					}
				}
			}
			// 处理cache
			h := _proxy(entry.Https, entry.Service, entry.Target)
			// 处理cache
			if compiler.Cache != nil && entry.Cache > 0 {
				handlers = append(handlers, compiler.Cache.Cache(entry.Cache, h))
			} else {
				handlers = append(handlers, h)
			}
			// 添加路由
			router.Handle(method, path, handlers...)
		}
	}
	return nil
}

func _proxy(https bool, serviceName string, uri string) gin.HandlerFunc {
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

// 索引化Entry, 格式为map{method: map{path: engine}}, 第一层索引是method, 第二层索引是path
func _index(httpEntry []Entry) map[string]map[string]*Entry {
	indexes := make(map[string]map[string]*Entry)
	for _, entry := range httpEntry {
		methods := entry.Method
		if len(methods) == 0 {
			methods = DefaultMethods
		}
		for _, method := range methods {
			entrymap, ok := indexes[method]
			if !ok {
				entrymap = make(map[string]*Entry)
				indexes[method] = entrymap
			}
			entrymap[entry.Source] = &entry
		}
	}
	return indexes
}

// 删除索引并返回Entry
func _dlink(indexes map[string]map[string]*Entry, method, path string) *Entry {
	entrymap, ok := indexes[method]
	if !ok {
		return nil
	}
	entry, ok := entrymap[path]
	if !ok {
		return nil
	}
	delete(entrymap, path)
	return entry
}
