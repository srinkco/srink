package httputils

import (
	"log"
	"sync"

	"github.com/valyala/fasthttp"
)

type Dispatcher struct {
	mu       sync.RWMutex
	m        map[string]fasthttp.RequestHandler
	notFound fasthttp.RequestHandler
}

type DispatcherOpts struct {
	NotFoundHandler fasthttp.RequestHandler
}

func NewDispatcher(opts *DispatcherOpts) *Dispatcher {
	if opts == nil {
		opts = &DispatcherOpts{}
	}
	if opts.NotFoundHandler == nil {
		opts.NotFoundHandler = defaultNotFound
	}
	return &Dispatcher{
		m:        make(map[string]fasthttp.RequestHandler),
		notFound: opts.NotFoundHandler,
	}
}

func (d *Dispatcher) Handle(ctx *fasthttp.RequestCtx) {
	d.mu.RLock()
	h, ok := d.m[string(ctx.Path())]
	d.mu.RUnlock()
	if !ok {
		d.notFound(ctx)
		return
	}
	h(ctx)
}

func (d *Dispatcher) HandleFunc(method string, handler fasthttp.RequestHandler, verbose bool) {
	if verbose {
		defer log.Println("Registered handler:", method)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.m[method] = handler
}

func defaultNotFound(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(404)
	ctx.SetBodyString("404 NOT FOUND")
}
