package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"

	_ "embed"

	"github.com/anonyindian/url-shortener/utils/apierrors"
	apiutil "github.com/anonyindian/url-shortener/utils/apiutils"
	"github.com/anonyindian/url-shortener/utils/httputils"
	"github.com/anonyindian/url-shortener/utils/shortener"
	"github.com/anonyindian/url-shortener/utils/templates"

	fasthttp "github.com/valyala/fasthttp"
)

func mainPage(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(http.StatusOK)
	ctx.SetContentType("text/html")
	ctx.SetBodyString(templates.MAIN_HTML)
}

func worker(engine shortener.Engine) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())[1:]
		_url := engine.GetUrl(path)
		if _url == "" {
			_url = "/"
		}
		ctx.Redirect(_url, http.StatusPermanentRedirect)
	}
}

func createUrl(auth string, engine shortener.Engine) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		resp := apiutil.NewResponse(ctx)
		args := ctx.QueryArgs()
		token := string(args.Peek("token"))
		if token != "" && token != auth {
			resp.SendUnauthorized(apierrors.ErrAuthTokenInvalid)
			return
		}
		_url := string(args.Peek("url"))
		if _url == "" {
			resp.SendBadRequest(apierrors.ErrEmptyUrl)
			return
		}
		if _, err := url.ParseRequestURI(_url); err != nil {
			resp.SendBadRequest(apierrors.ErrInvalidUri)
			return
		}
		var hash string
		if cusHash := string(args.Peek("hash")); cusHash != "" {
			if token == "" {
				resp.SendUnauthorized(apierrors.ErrAuthTokenMissing)
				return
			}
			hash = cusHash
		}
		hash = engine.Shorten(_url, hash)
		surl := strings.Join([]string{string(ctx.Host()), hash}, "/")
		resp.SendSuccess(surl)
		log.Println("Created new shorturl from url (", _url, ") to", "surl (", surl, ")")
		return
	}
}

func printGeneralDetails() {
	log.Println("Version 1.0.0")
	log.Println("GOOS:", runtime.GOOS, "GOVersion:", runtime.Version(), "GOARCH:", runtime.GOARCH)
}

func getLocalIPAddr() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok &&
				!ipnet.IP.IsLoopback() &&
				!ipnet.IP.IsUnspecified() &&
				ipnet.IP.To4() != nil {
				return ipnet.IP
			}
		}
	}
	return net.IPv4(0, 0, 0, 0)
}

func main() {
	log.Println("Starting server")
	printGeneralDetails()

	conf := readServerConfig(DEFAULT_PORT)

	log.Println("Settings:", conf.data)

	log.Println("Creating new shortener engine")
	engine := shortener.NewEngine(shortener.EngineTypeInMemory)

	d := httputils.NewDispatcher(&httputils.DispatcherOpts{
		NotFoundHandler: worker(engine),
	})

	auth := conf.getString("token")
	fmt.Println("----------------------------------")
	log.Println("Your auth token is:", auth)
	fmt.Println("----------------------------------")

	d.HandleFunc("/api/new", createUrl(auth, engine), true)
	d.HandleFunc("/", mainPage, true)

	port := conf.getString("port")
	log.Println("Local network IPv4:", getLocalIPAddr())

	log.Println("Starting web server on port:", port)
	fasthttp.ListenAndServe(":"+port, d.Handle)
}
