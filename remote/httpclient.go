package remote

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var bodyType = "json"

type Request struct {
	serverHost   string
	relativePath string
	body         io.Reader
	headers      map[string]string
	querys       map[string]string
}

type Response struct {
	Data       []byte
	StatusCode int
	Error      error
}

func SetBodyType(t string) {
	bodyType = t
}

func NewRequest(address string, relativePath string, a ...interface{}) *Request {
	return &Request{
		serverHost:   address,
		relativePath: fmt.Sprintf(relativePath, a...),
		headers:      make(map[string]string),
		querys:       make(map[string]string),
	}
}

func (req *Request) SetBody(body interface{}) *Request {
	var data string
	if ret, ok := body.(string); ok {
		data = ret
	} else if ret, ok := body.([]byte); ok {
		data = string(ret)
	} else if bodyType == "xml" {
		ret, _ := xml.Marshal(body)
		data = xml.Header + string(ret)
	} else {
		ret, _ := json.Marshal(body)
		data = string(ret)
	}
	req.body = strings.NewReader(data)
	return req
}

func absPath(p string) string {
	var absPath string
	parts := strings.Split(p, "/")
	for _, part := range parts {
		if part == "~" {
			part = os.Getenv("HOME")
		} else if strings.HasPrefix(part, "$") {
			part = os.Getenv(part[1:])
		}
		absPath = filepath.Join(absPath, part)
	}
	return absPath
}

func (req *Request) SetFile(filePath string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	//关键的一步操作
	fileWriter, err := bodyWriter.CreateFormFile("chaincode", absPath(filePath))
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	//打开文件句柄操作
	fh, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error opening file", err.Error())
		return err
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		fmt.Println("error copy file", err.Error())
		return err
	}

	bodyWriter.Close()
	req.body = bodyBuf
	req.SetHeader("Content-Type", bodyWriter.FormDataContentType())
	return nil
}

func (req *Request) SetHeader(key, value string) *Request {
	req.headers[key] = value
	return req
}

func (req *Request) SetQuery(key, value string) *Request {
	if value != "" {
		req.querys[key] = value
	}
	return req
}

func (req *Request) GET() Response {
	return req.Do("GET")
}

func (req *Request) POST() Response {
	return req.Do("POST")
}

func (req *Request) PATCH() Response {
	return req.Do("PATCH")
}

func (req *Request) PUT() Response {
	return req.Do("PUT")
}

func (req *Request) DELETE() Response {
	return req.Do("DELETE")
}

func (req *Request) Do(method string) Response {
	client := &http.Client{}
	var strQuery string
	for k, v := range req.querys {
		strQuery += fmt.Sprintf("&%s=%s", k, v)
	}
	if strQuery != "" {
		strQuery = "?" + strQuery[1:]
	}

	strURL := req.serverHost
	if !strings.HasPrefix(req.relativePath, "/") {
		strURL += "/"
	}
	strURL += req.relativePath + strQuery
	r, err := http.NewRequest(method, strURL, req.body)
	if err != nil {
		return Response{nil, http.StatusBadRequest, err}
	}
	for k, v := range req.headers {
		r.Header.Set(k, v)
	}

	resp, err := client.Do(r)
	if err != nil {
		if resp == nil {
			return Response{nil, http.StatusBadRequest, err}
		}
		return Response{nil, resp.StatusCode, err}
	}
	defer resp.Body.Close()

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{ret, resp.StatusCode, err}
	}

	return Response{ret, resp.StatusCode, nil}
}
