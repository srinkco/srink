package main

import (
	"log"
	"strings"

	"github.com/anonyindian/url-shortener/utils/apierrors"
	apiutil "github.com/anonyindian/url-shortener/utils/apiutils"
	"github.com/valyala/fasthttp"
)

type client struct {
	conf   *config
	log    *log.Logger
	apiUrl string
}

func newClient(l *log.Logger) *client {
	c := &client{
		log: l,
	}
	c.readConfig()
	c.apiUrl = strings.TrimSuffix(
		c.conf.getString("api-url"),
		"/",
	)
	return c
}

func (c *client) updateApiUrl(apiUrl string) {
	apiUrl = strings.TrimSuffix(apiUrl, "/")
	c.apiUrl = apiUrl
	c.conf.add("api-url", apiUrl)
	c.conf.write()
}

func (c *client) readConfig() {
	c.conf = readUserConfig("client-conf.yml", c.log)
	c.conf.tryAdd("api-url", DEFAULT_API_URL)
	c.conf.write()
}

func (c *client) shortenUrl(hash, url string) (string, error) {
	args := fasthttp.AcquireArgs()
	args.Add("hash", hash)
	args.Add("url", url)
	args.Add("token", c.conf.getString("token"))
	defer fasthttp.ReleaseArgs(args)
	code, body, err := fasthttp.Post(
		nil,
		strings.Join(
			[]string{c.apiUrl, "api/new"},
			"/",
		),
		args,
	)
	if err != nil {
		return "", err
	}
	resp := apiutil.UnmarshalResponse(body)
	if code != 200 {
		return "", apierrors.New("shorten url", code, resp.Description)
	}
	return resp.ShortenedUrl, nil
}
