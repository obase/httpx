package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/obase/httpx/cache"
	"net/http"
	"strings"
)

/*
扩展gin.Engine:
1. 缓存所有gin.HandlerFunc
2. 整合httpCache, httpPlugin等扩展功能
3. 对等转换为gin.Engine
*/

type IRouter interface {
	Group(path string, fs ...gin.HandlerFunc) IRouter

	Use(fs ...gin.HandlerFunc) IRouter

	Handle(method string, path string, f gin.HandlerFunc) IRouter
	Any(path string, f gin.HandlerFunc) IRouter
	GET(path string, f gin.HandlerFunc) IRouter
	POST(path string, f gin.HandlerFunc) IRouter
	DELETE(path string, f gin.HandlerFunc) IRouter
	PATCH(path string, f gin.HandlerFunc) IRouter
	PUT(path string, f gin.HandlerFunc) IRouter
	OPTIONS(path string, f gin.HandlerFunc) IRouter
	HEAD(path string, f gin.HandlerFunc) IRouter

	StaticFile(path string, file string) IRouter
	Static(path string, file string) IRouter
	StaticFS(path string, fs http.FileSystem) IRouter
}

type RouteNode struct {
	path       string
	use        gin.HandlersChain
	handle     map[string]map[string]gin.HandlerFunc
	staticFile map[string]string
	static     map[string]string
	staticFS   map[string]http.FileSystem
	child      []*RouteNode
}

var _ IRouter = (*RouteNode)(nil)

func newRouteNode(path string, use gin.HandlersChain) *RouteNode {
	return &RouteNode{
		path:       path,
		use:        use,
		handle:     make(map[string]map[string]gin.HandlerFunc),
		staticFile: make(map[string]string),
		static:     make(map[string]string),
		staticFS:   make(map[string]http.FileSystem),
		child:      nil,
	}
}

func (r *RouteNode) Group(path string, use ...gin.HandlerFunc) IRouter {
	sr := newRouteNode(path, use)
	r.child = append(r.child, sr)
	return sr
}

func (r *RouteNode) Use(fs ...gin.HandlerFunc) IRouter {
	r.use = append(r.use, fs...)
	return r
}

func (r *RouteNode) Handle(method string, path string, f gin.HandlerFunc) IRouter {
	mhandle, ok := r.handle[method]
	if !ok {
		mhandle = make(map[string]gin.HandlerFunc)
		r.handle[method] = mhandle
	}
	mhandle[path] = f
	return r
}
func (r *RouteNode) Any(path string, f gin.HandlerFunc) IRouter {
	r.Handle(http.MethodGet, path, f)
	r.Handle(http.MethodPost, path, f)
	r.Handle(http.MethodDelete, path, f)
	r.Handle(http.MethodPatch, path, f)
	r.Handle(http.MethodPut, path, f)
	r.Handle(http.MethodOptions, path, f)
	r.Handle(http.MethodHead, path, f)
	return r
}
func (r *RouteNode) GET(path string, f gin.HandlerFunc) IRouter {
	r.Handle(http.MethodGet, path, f)
	return r
}
func (r *RouteNode) POST(path string, f gin.HandlerFunc) IRouter {
	r.Handle(http.MethodPost, path, f)
	return r
}
func (r *RouteNode) DELETE(path string, f gin.HandlerFunc) IRouter {
	r.Handle(http.MethodDelete, path, f)
	return r
}
func (r *RouteNode) PATCH(path string, f gin.HandlerFunc) IRouter {
	r.Handle(http.MethodPatch, path, f)
	return r
}
func (r *RouteNode) PUT(path string, f gin.HandlerFunc) IRouter {
	r.Handle(http.MethodPut, path, f)
	return r
}
func (r *RouteNode) OPTIONS(path string, f gin.HandlerFunc) IRouter {
	r.Handle(http.MethodOptions, path, f)
	return r
}
func (r *RouteNode) HEAD(path string, f gin.HandlerFunc) IRouter {
	r.Handle(http.MethodHead, path, f)
	return r
}

func (r *RouteNode) StaticFile(path string, file string) IRouter {
	r.staticFile[path] = file
	return r
}
func (r *RouteNode) Static(path string, file string) IRouter {
	r.static[path] = file
	return r
}
func (r *RouteNode) StaticFS(path string, fs http.FileSystem) IRouter {
	r.staticFS[path] = fs
	return r
}

type Server struct {
	*RouteNode
	Plugins []*Plugin
}

var _ IRouter = (*Server)(nil)

func New() *Server {
	return &Server{RouteNode: newRouteNode("", nil)}
}

// note: plugins是有序且不区分大小写
func (s *Server) Plugin(name string, f func(args []string) gin.HandlerFunc) {
	s.Plugins = append(s.Plugins, &Plugin{Name: strings.ToLower(name), Func: f})
}

// 整合cache,plugin生成gin的Engine, 由外部负责关闭cache
func (s *Server) build(config *Config) (*gin.Engine, cache.Cache, error) {
	bd := newEngineBuilder(config)
	defer bd.dispose()

	engine := gin.New()
	cache := cache.New(config.HttpCache)

	if err := bd.buildRoute(engine, s.Plugins, cache, "", s.RouteNode); err != nil {
		cache.Close()
		return nil, nil, err
	}

	if err := bd.buildProxy(engine, s.Plugins, cache); err != nil {
		cache.Close()
		return nil, nil, err
	}
	return engine, cache, nil
}

func (s *Server) Build(config *Config) (http.Handler, cache.Cache, error) {
	return s.build(config)
}

func (s *Server) Run(config *Config, addr ...string) (err error) {
	engine, cache, err := s.build(config)
	if err != nil {
		return
	}
	defer cache.Close()
	return engine.Run(addr...)
}

func (s *Server) RunTLS(config *Config, addr, certFile, keyFile string) (err error) {
	engine, cache, err := s.build(config)
	if err != nil {
		return
	}
	defer cache.Close()
	return engine.RunTLS(addr, certFile, keyFile)
}

func (s *Server) RunUnix(config *Config, file string) (err error) {
	engine, cache, err := s.build(config)
	if err != nil {
		return
	}
	defer cache.Close()
	return engine.RunUnix(file)
}

func (s *Server) RunFd(config *Config, fd int) (err error) {
	engine, cache, err := s.build(config)
	if err != nil {
		return
	}
	defer cache.Close()
	return engine.RunFd(fd)
}
