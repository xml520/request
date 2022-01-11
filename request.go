package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client http请求客户端
var Client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Timeout: time.Second * 30,
}

// Request  基础请求参数
type Request struct {
	BaseUri     string            // 基础uri
	BaseHeaders map[string]string //基础请求头
	handleError func(s string) error
}

// Response 响应
type Response struct {
	BodyStr string //body字符串
	*http.Response
}

func (br Request) Get(uri string, headers ...map[string]string) (*Response, error) {
	return br.request("GET", uri, "", headers...)
}
func (br Request) Post(uri string, body interface{}, headers ...map[string]string) (*Response, error) {
	return br.request("POST", uri, body, headers...)
}
func (br Request) Put(uri string, body interface{}, headers ...map[string]string) (*Response, error) {
	return br.request("PUT", uri, body, headers...)
}
func (br Request) Delete(uri string, body interface{}, headers ...map[string]string) (*Response, error) {
	return br.request("DELETE", uri, body, headers...)
}
func (br Request) Upload(uri string, buf []byte, headers ...map[string]string) (*Response, error) {
	//合并url
	url := MergeUri(br.BaseUri, uri)
	if url == "" {
		return nil, errors.New("url不能为空")
	}
	req, err := http.NewRequest("PUT", url, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	//req.Header = map[string][]string{}
	for k, v := range br.BaseHeaders {
		req.Header.Set(k, v)
	}
	if len(headers) > 0 {
		for k, v := range headers[0] {
			req.Header.Set(k, v)
		}
	}
	res, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	rs := &Response{BodyStr: "", Response: res}
	return rs, nil
}

// 自定义请求
func (br *Request) request(m string, uri string, body interface{}, headers ...map[string]string) (*Response, error) {
	//合并url
	url := MergeUri(br.BaseUri, uri)
	if url == "" {
		return nil, errors.New("url不能为空")
	}
	var params io.Reader
	switch body.(type) {
	case string:
		params = strings.NewReader(body.(string))
	case nil:
		params = nil
	default:
		b, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		params = bytes.NewReader(b)
	}
	req, err := http.NewRequest(m, url, params)
	for k, v := range br.BaseHeaders {
		req.Header.Set(k, v)
	}
	if len(headers) > 0 {
		for k, v := range headers[0] {
			req.Header.Set(k, v)
		}
	}
	res, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	resBody := string(buf)
	rs := &Response{BodyStr: resBody, Response: res}
	var throwErr error
	if res.StatusCode > 399 {
		if br.handleError != nil {
			throwErr = br.handleError(resBody)
		} else {
			throwErr = errors.New(strconv.Itoa(res.StatusCode))
		}
	}
	return rs, throwErr
}

// MergeCookie 合并cookie
func (res *Response) MergeCookie() string {
	oldCookies := res.Request.Header.Get("Cookie")
	cookies := make(map[string]string)
	if oldCookies != "" {
		// 旧cookie
		oldCookie := strings.Split(oldCookies, ";")
		for _, v := range oldCookie {
			i := strings.Index(v, "=")
			if i == -1 {
				continue
			}
			cookies[v[:i]] = v[i+1:]
		}
	}
	// 新cookie
	//fmt.Println("新cookie", res.Cookies())
	for _, v := range res.Cookies() {
		if v.Value == "" || v.Value == "0" {
			continue
		}
		cookies[v.Name] = v.Value
	}
	cookieStr := ""
	for k, v := range cookies {
		cookieStr = cookieStr + k + "=" + v + ";"
	}
	return cookieStr
}

// GetCookie cookie
func (res *Response) GetCookie() string {
	cookieStr := ""
	for _, v := range res.Cookies() {
		if v.Value == "" || v.Value == "0" {
			continue
		}
		cookieStr = cookieStr + v.Name + "=" + v.Value + ";"
	}
	return cookieStr
}

func MergeUri(b string, u string) string {
	if strings.Index(u, "://") != -1 {
		return u
	}
	if b == "" && u == "" {
		return ""
	}
	if b == "" {
		return u
	}
	if b[len(b)-1:] != "/" {
		b = b + "/"
	}
	if u != "" && u[:1] == "/" {
		u = u[1:]
	}
	return b + u
}
