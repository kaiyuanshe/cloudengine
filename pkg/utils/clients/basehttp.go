package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultHttpTimeout = time.Second * 2
)

type HttpClient struct {
	Host string
	cli  *http.Client
}

func (h HttpClient) Get(path string, query map[string]string, result interface{}) error {
	return h.do(http.MethodGet, path, query, nil, result)
}

func (h HttpClient) Post(path string, body, result interface{}) error {
	return h.do(http.MethodPost, path, nil, body, result)
}

func (h HttpClient) do(method, path string, query map[string]string, body, result interface{}) error {
	u, err := url.Parse(h.Host)
	if err != nil {
		return fmt.Errorf("parse url failed: %s", err.Error())
	}

	q := u.Query()
	for qk, qv := range query {
		q.Add(qk, qv)
	}

	u.Path = path
	u.RawQuery = q.Encode()

	klog.V(7).Infof("http %s to %s", method, u.String())
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(body); err != nil {
		return fmt.Errorf("encode body failed: %s", err.Error())
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return fmt.Errorf("build request failed: %s", err.Error())
	}

	resp, err := h.cli.Do(req)
	if err != nil {
		return err
	}

	klog.V(7).Infof("http %s to %s, status code: %d", method, u.String(), resp.StatusCode)

	defer func() {
		_, _ = ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}()

	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(result)
}

func NewDefaultHttpClient(host string) HttpClient {
	return HttpClient{
		Host: host,
		cli: &http.Client{
			Timeout: defaultHttpTimeout,
		},
	}
}
