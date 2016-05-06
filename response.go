package httpexpect

import (
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
)

type Response struct {
	checker Checker
	resp    *http.Response
	content []byte
}

func NewResponse(checker Checker, resp *http.Response) *Response {
	return &Response{checker, resp, nil}
}

func (r *Response) Raw() *http.Response {
	return r.resp
}

func (r *Response) Status(status int) *Response {
	if r.checker.Failed() {
		return r
	}
	r.checker.Equal(status, r.resp.StatusCode)
	return r
}

func (r *Response) Headers(headers map[string][]string) *Response {
	if r.checker.Failed() {
		return r
	}
	r.checker.Equal(headers, map[string][]string(r.resp.Header))
	return r
}

func (r *Response) Header(k, v string) *Response {
	if r.checker.Failed() {
		return r
	}
	r.checker.Equal(v, r.resp.Header.Get(k))
	return r
}

func (r *Response) NoContent() *Response {
	if r.checker.Failed() {
		return r
	}

	contentType := r.resp.Header.Get("Content-Type")

	content := string(r.getContent())

	r.checker.Equal("", contentType)
	r.checker.Equal("", content)

	return r
}

func (r *Response) JSON() *Value {
	value := r.getJSON()
	return NewValue(r.checker.Clone(), value)
}

func (r *Response) getContent() []byte {
	if r.content != nil {
		return r.content
	}

	content, err := ioutil.ReadAll(r.resp.Body)
	if err != nil {
		r.checker.Fail(err.Error())
		return nil
	}

	r.content = content
	return r.content
}

func (r *Response) getJSON() interface{} {
	if r.checker.Failed() {
		return nil
	}

	contentType := r.resp.Header.Get("Content-Type")

	mediaType, params, _ := mime.ParseMediaType(contentType)
	charset := params["charset"]

	r.checker.Equal("application/json", mediaType)
	if r.checker.Failed() {
		return nil
	}

	if charset != "" && strings.ToLower(charset) != "utf-8" {
		r.checker.Fail("bad charset: expected empty or 'utf-8', got '"+charset+"'")
		return nil
	}

	content := r.getContent()

	var value interface{}
	if err := json.Unmarshal(content, &value); err != nil {
		r.checker.Fail(err.Error())
		return nil
	}

	return value
}
