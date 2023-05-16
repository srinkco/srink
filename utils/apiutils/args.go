package apiutil

import "github.com/valyala/fasthttp"

func Peek(qArgs, pArgs *fasthttp.Args, key string) string {
	buf := qArgs.Peek(key)
	if buf != nil && len(buf) != 0 {
		return string(buf)
	}
	return string(pArgs.Peek(key))
}
