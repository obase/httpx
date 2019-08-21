package httpx

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/obase/center"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

const (
	REVERSE_SCHEME = "rx-scheme"
	REVERSE_HOST   = "rx-host"
	REVERSE_PATH   = "rx-path"
)

var (
	defaultConfig       *Config
	defaultTransport    *http.Transport
	defaultClient       *http.Client
	defaultReverseProxy *httputil.ReverseProxy
)

func init() {
	Setup(LoadConfig())
}

func Setup(c *Config) {
	defaultConfig = mergeConfig(c)

	defaultTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   defaultConfig.ConnectTimeout,
			KeepAlive: defaultConfig.KeepAlive,
		}).DialContext,
		MaxIdleConns:           defaultConfig.MaxIdleConns,
		MaxIdleConnsPerHost:    defaultConfig.MaxIdleConnsPerHost,
		MaxConnsPerHost:        defaultConfig.MaxConnsPerHost,
		IdleConnTimeout:        defaultConfig.IdleConnTimeout,
		DisableCompression:     defaultConfig.DisableCompression,
		ResponseHeaderTimeout:  defaultConfig.ResponseHeaderTimeout,
		ExpectContinueTimeout:  defaultConfig.ExpectContinueTimeout,
		MaxResponseHeaderBytes: defaultConfig.MaxResponseHeaderBytes,
	}
	defaultClient = &http.Client{
		Transport: defaultTransport,
		Timeout:   defaultConfig.RequestTimeout,
	}
	defaultReverseProxy = &httputil.ReverseProxy{
		Transport:     defaultTransport,
		FlushInterval: defaultConfig.ProxyFlushInterval,
		Director: func(req *http.Request) {
			req.URL.Scheme = req.Header.Get(REVERSE_SCHEME)
			req.URL.Host = req.Header.Get(REVERSE_HOST)
			req.URL.Path = req.Header.Get(REVERSE_PATH)
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		},
		BufferPool:   proxyBufferPool(defaultConfig.ProxyBufferPool),
		ErrorHandler: proxyErrorHandler(defaultConfig.ProxyErrorHandler),
	}
}

func Request(method string, https bool, serviceName string, uri string, header map[string]string, body io.Reader) (state int, content string, err error) {
	service, err := center.Robin(serviceName)
	if err != nil {
		return
	}

	var schema string
	if https {
		schema = "https://"
	} else {
		schema = "http://"
	}
	url := schema + service.Host + ":" + strconv.Itoa(service.Port) + uri
	// 创建请求
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rsp, err := defaultClient.Do(req)
	if err != nil {
		return
	}
	defer rsp.Body.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, bufio.NewReader(rsp.Body))
	if err != nil {
		return
	}
	return rsp.StatusCode, buf.String(), nil
}

func RequestURL(method string, url string, header map[string]string, body io.Reader) (state int, content string, err error) {

	// 创建请求
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rsp, err := defaultClient.Do(req)
	if err != nil {
		return
	}
	defer rsp.Body.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, bufio.NewReader(rsp.Body))
	if err != nil {
		return
	}
	return rsp.StatusCode, buf.String(), nil
}

func Post(serviceName string, uri string, header map[string]string, reqobj interface{}, rspobj interface{}) (status int, err error) {
	data, err := json.Marshal(reqobj)
	if err != nil {
		return
	}
	status, content, err := Request(http.MethodPost, false, serviceName, uri, header, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	if status < 200 || status > 299 {
		err = errors.New(content)
	} else {
		err = json.Unmarshal([]byte(content), &rspobj)
	}
	return
}

func PostTLS(serviceName string, uri string, header map[string]string, reqobj interface{}, rspobj interface{}) (status int, err error) {
	data, err := json.Marshal(reqobj)
	if err != nil {
		return
	}
	status, content, err := Request(http.MethodPost, true, serviceName, uri, header, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	if status < 200 || status > 299 {
		err = errors.New(content)
	} else {
		err = json.Unmarshal([]byte(content), &rspobj)
	}
	return
}

func PostURL(url string, header map[string]string, reqobj interface{}, rspobj interface{}) (status int, err error) {
	data, err := json.Marshal(reqobj)
	if err != nil {
		return
	}
	status, content, err := RequestURL(http.MethodPost, url, header, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	if status < 200 || status > 299 {
		err = errors.New(content)
	} else {
		err = json.Unmarshal([]byte(content), &rspobj)
	}
	return
}

func Proxy(serviceName string, uri string, writer http.ResponseWriter, request *http.Request) (err error) {
	service, err := center.Robin(serviceName)
	if service != nil && err == nil {
		request.Header.Set(REVERSE_SCHEME, "http")
		request.Header.Set(REVERSE_HOST, service.Host+":"+strconv.Itoa(service.Port))
		request.Header.Set(REVERSE_PATH, uri)
		defaultReverseProxy.ServeHTTP(writer, request)
	} else {
		writer.WriteHeader(http.StatusBadGateway)
	}
	return
}

func ProxyTLS(serviceName string, uri string, writer http.ResponseWriter, request *http.Request) (err error) {
	service, err := center.Robin(serviceName)
	if service != nil && err == nil {
		request.Header.Set(REVERSE_SCHEME, "https")
		request.Header.Set(REVERSE_HOST, service.Host+":"+strconv.Itoa(service.Port))
		request.Header.Set(REVERSE_PATH, uri)
		defaultReverseProxy.ServeHTTP(writer, request)
	} else {
		writer.WriteHeader(http.StatusBadGateway)
	}
	return
}

func ProxyURL(rurl string, writer http.ResponseWriter, request *http.Request) (err error) {
	purl, err := url.Parse(rurl)
	if purl != nil && err == nil {
		request.Header.Set(REVERSE_SCHEME, purl.Scheme)
		request.Header.Set(REVERSE_HOST, purl.Host)
		request.Header.Set(REVERSE_PATH, purl.Path)
		defaultReverseProxy.ServeHTTP(writer, request)
	} else {
		writer.WriteHeader(http.StatusBadGateway)
	}
	return
}

func ProxyHandler(serviceName string, uri string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Transport:     defaultTransport,
		FlushInterval: defaultConfig.ProxyFlushInterval,
		Director: func(req *http.Request) {
			service, _ := center.Robin(serviceName)
			if service != nil {
				req.URL.Scheme = "http"
				req.URL.Host = service.Host + ":" + strconv.Itoa(service.Port)
				req.URL.Path = uri
				if _, ok := req.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					req.Header.Set("User-Agent", "")
				}
			}
		},
		BufferPool:   proxyBufferPool(defaultConfig.ProxyBufferPool),
		ErrorHandler: proxyErrorHandler(defaultConfig.ProxyErrorHandler),
	}
}

func ProxyHandlerTLS(serviceName string, uri string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Transport:     defaultTransport,
		FlushInterval: defaultConfig.ProxyFlushInterval,
		Director: func(req *http.Request) {
			service, _ := center.Robin(serviceName)
			if service != nil {
				req.URL.Scheme = "https"
				req.URL.Host = service.Host + ":" + strconv.Itoa(service.Port)
				req.URL.Path = uri
				if _, ok := req.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					req.Header.Set("User-Agent", "")
				}
			}
		},
		BufferPool:   proxyBufferPool(defaultConfig.ProxyBufferPool),
		ErrorHandler: proxyErrorHandler(defaultConfig.ProxyErrorHandler),
	}
}

func ProxyHandlerURL(rurl string) *httputil.ReverseProxy {
	purl, _ := url.Parse(rurl)
	return &httputil.ReverseProxy{
		Transport:     defaultTransport,
		FlushInterval: defaultConfig.ProxyFlushInterval,
		Director: func(req *http.Request) {
			req.URL.Scheme = purl.Scheme
			req.URL.Host = purl.Host
			req.URL.Path = purl.Path
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		},
		BufferPool:   proxyBufferPool(defaultConfig.ProxyBufferPool),
		ErrorHandler: proxyErrorHandler(defaultConfig.ProxyErrorHandler),
	}
}
