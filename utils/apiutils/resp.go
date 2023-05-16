package apiutil

import (
	"encoding/json"
	"net/http"

	"github.com/valyala/fasthttp"
)

type Response struct {
	Ok           bool   `json:"ok"`
	Description  string `json:"error,omitempty"`
	ShortenedUrl string `json:"shortened_url,omitempty"`
	ctx          *fasthttp.RequestCtx
}

func NewResponse(ctx *fasthttp.RequestCtx) *Response {
	ctx.SetContentType("application/json")
	return &Response{
		ctx: ctx,
	}
}

func (r *Response) SendSuccess(surl string) {
	r.ctx.SetStatusCode(http.StatusOK)
	r.Ok = true
	r.ShortenedUrl = surl
	r.ctx.SetBody(r.marshal())
}

func (r *Response) SendError(status int, err error) {
	r.ctx.SetStatusCode(status)
	r.Ok = false
	r.Description = err.Error()
	r.ctx.SetBody(r.marshal())
}

func (r *Response) SendBadRequest(err error) {
	r.SendError(http.StatusBadRequest, err)
}

func (r *Response) SendUnauthorized(err error) {
	r.SendError(http.StatusUnauthorized, err)
}

func (r *Response) marshal() []byte {
	buf, _ := json.Marshal(r)
	return buf
}

func UnmarshalResponse(data []byte) *Response {
	var r Response
	_ = json.Unmarshal(data, &r)
	return &r
}
