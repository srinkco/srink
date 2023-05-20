package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/anonyindian/url-shortener/utils/apierrors"
	apiutil "github.com/anonyindian/url-shortener/utils/apiutils"
	"github.com/anonyindian/url-shortener/utils/httputils"
	"github.com/anonyindian/url-shortener/utils/randomiser"
	"github.com/anonyindian/url-shortener/utils/shortener"
	"github.com/anonyindian/url-shortener/utils/templates"
	fasthttp "github.com/valyala/fasthttp"
)

type server struct {
	auth            string
	conf            *config
	dp              *httputils.Dispatcher
	engine          shortener.Engine
	log             *log.Logger
	fallback        bool
	cssBuf, htmlBuf []byte
}

func newServer(l *log.Logger) *server {
	return &server{
		log: l,
	}
}

func (s *server) start() {
	s.log.Println("Starting server")
	s.printGeneralDetails()

	s.readConfig(DEFAULT_PORT)

	s.log.Println("Settings:", s.conf.data)

	s.log.Println("Creating new shortener engine")
	s.engine = shortener.NewEngine(shortener.EngineTypeInMemory)

	s.dp = httputils.NewDispatcher(&httputils.DispatcherOpts{
		NotFoundHandler: s.worker,
	})

	s.auth = s.conf.getString("token")
	fmt.Println("----------------------------------")
	s.log.Println("Your auth token is:", s.auth)
	fmt.Println("----------------------------------")

	s.log.Println("Fetching frontend for the index page...")
	s.initHTML("frontend/index.html")
	s.initCSS("frontend/tailwind.css")

	s.dp.HandleFunc("/api/new", s.createUrl, true)
	s.dp.HandleFunc("/", s.mainPage, true)
	s.dp.HandleFunc("/tailwind.css", s.serverCSS, true)

	port := s.conf.getString("port")
	s.log.Println("Local network IPv4:", s.getLocalIPAddr())

	s.log.Println("Starting web server on port:", port)
	fasthttp.ListenAndServe(":"+port, s.dp.Handle)
}

func (s *server) readConfig(port int64) {
	s.conf = readUserConfig("server-conf.yml", s.log)
	s.conf.tryAdd("token", randomiser.GetString(16))
	s.conf.tryAdd("port", port)
	s.conf.write()
}

func (s *server) mainPage(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("text/html")
	if s.fallback {
		ctx.SetBodyString(templates.MAIN_HTML)
		return
	}
	ctx.SetBody(s.htmlBuf)
}

func (s *server) worker(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())[1:]
	_url := s.engine.GetUrl(path)
	if _url == "" {
		_url = "/"
	}
	ctx.Redirect(_url, fasthttp.StatusPermanentRedirect)
}

func (s *server) createUrl(ctx *fasthttp.RequestCtx) {
	resp := apiutil.NewResponse(ctx)
	qArgs, pArgs := ctx.QueryArgs(), ctx.PostArgs()
	token := apiutil.Peek(qArgs, pArgs, "token")
	if token != "" && token != s.auth {
		resp.SendUnauthorized(apierrors.ErrAuthTokenInvalid)
		return
	}
	_url := apiutil.Peek(qArgs, pArgs, "url")
	if _url == "" {
		resp.SendBadRequest(apierrors.ErrEmptyUrl)
		return
	}
	if _, err := url.ParseRequestURI(_url); err != nil {
		fmt.Println(_url)
		resp.SendBadRequest(apierrors.ErrInvalidUri)
		return
	}
	var hash string
	if cusHash := apiutil.Peek(qArgs, pArgs, "hash"); cusHash != "" {
		if token == "" {
			resp.SendUnauthorized(apierrors.ErrAuthTokenMissing)
			return
		}
		hash = cusHash
	}
	hash = s.engine.Shorten(_url, hash)
	surl := strings.Join([]string{string(ctx.Host()), hash}, "/")
	resp.SendSuccess(surl)
	s.log.Println("Created new shorturl from url (", _url, ") to", "surl (", surl, ")")
	return
}

func (s *server) printGeneralDetails() {
	s.log.Println("Version 1.0.0")
	s.log.Println("GOOS:", runtime.GOOS, "GOVersion:", runtime.Version(), "GOARCH:", runtime.GOARCH)
}

func (s *server) getLocalIPAddr() net.IP {
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

func (s *server) initCSS(name string) {
	buf, err := os.ReadFile(name)
	if err != nil {
		s.log.Println("Failed to open CSS file:", err)
		s.fallback = true
		s.log.Println("Switched to fallback mode...")
		return
	}
	s.cssBuf = buf
}

func (s *server) initHTML(name string) {
	buf, err := os.ReadFile(name)
	if err != nil {
		s.log.Println("Failed to open HTML file:", err)
		s.fallback = true
		s.log.Println("Switched to fallback mode...")
		return
	}
	s.htmlBuf = buf
}

func (s *server) serverCSS(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("text/css")
	ctx.SetBody(s.cssBuf)
}

func (s *server) passCORS(header *fasthttp.ResponseHeader) {
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Headers", "Cache-Control, Pragma, Origin, Authorization, Content-Type, X-Requested-With")
	header.Set("Access-Control-Allow-Methods", "GET, PUT, POST")
}
