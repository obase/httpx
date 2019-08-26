package ginx

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/obase/conf"
	"strings"
)

type Plugin struct {
	Name string // 统一要求小写
	Func func(args []string) gin.HandlerFunc
}

// 解析表达式, 语法为<plugin_name>(<plugin_argument>,...)
func ParsePluginExpress(express string, defargs map[string]string) (name string, args []string, err error) {
	ps1 := strings.IndexByte(express, '(')
	ps2 := strings.LastIndexByte(express, ')')

	if ps1 == -1 && ps2 == -1 {
		name = strings.TrimSpace(express)
		if v, ok := defargs[name]; ok {
			args = conf.ToStringSlice(v)
		}
	} else if ps1 != -1 && ps2 != -1 {
		name = strings.TrimSpace(express[:ps1])
		for _, arg := range strings.Split(express[ps1+1:ps2], ",") {
			arg := strings.TrimSpace(arg)
			if len(arg) > 0 {
				if arg[0] == '$' {
					if v, ok := defargs[arg[1:]]; ok {
						args = append(args, v)
					} else {
						args = append(args, "")
					}
				} else {
					args = append(args, arg)
				}
			}
		}
	} else {
		err = errors.New("invalid syntax: " + express)
	}
	return
}

// 从前往后遍历, 确保插件是有序执行的
func EvaluePluginExpress(plugins []Plugin, defargs map[string]string, express string) (gin.HandlerFunc, error) {
	name, args, err := ParsePluginExpress(express, defargs)
	if err != nil {
		return nil, err
	}

	name = strings.ToLower(name)
	for _, plugin := range plugins {
		if name == plugin.Name {
			return plugin.Func(args), nil
		}
	}
	return nil, errors.New("invalid plugin: " + name)
}
