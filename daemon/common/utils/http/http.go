/*
Copyright © 2021 KeenTune

Package http for daemon, this package contains the http. The package contains the remote call the url with the specified method and request data, get success response after request url, and parses the return data.
*/
package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/log"
	"net"
	"net/http"
	"time"
)

// Protocol  ...
const Protocol = "http"

// StatusOk ...
const StatusOk = 200

type requester struct {
	client *http.Client
	url    string
	body   []byte
	method string
}

func newRequester(method string, uri string, data interface{}) (*requester, error) {
	url := Protocol + "://" + uri
	var client = &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 900 * time.Millisecond, // conn timeout
			}).DialContext,
			DisableKeepAlives:     true,
			ResponseHeaderTimeout: 5 * time.Minute,
		},
	}

	if data == nil && method == "GET" {
		return &requester{
			client: client,
			url:    url,
			body:   nil,
			method: method,
		}, nil
	}

	bytesData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("[newRequester] error in Unexpected type %T\n", data)
		return nil, err
	}

	log.Debugf("", "[%v] request info [%+v]", url, string(bytesData))
	return &requester{
		client: client,
		url:    url,
		body:   bytesData,
		method: method,
	}, nil
}

func (r *requester) execute() (*http.Response, error) {
	request, err := http.NewRequest(r.method, r.url, bytes.NewReader(r.body))

	if err != nil {
		return nil, err
	}

	if r.body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := r.client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// RemoteCall remote call the url with the specified method and request data
func RemoteCall(method, url string, data interface{}) ([]byte, error) {
	requester, err := newRequester(method, url, data)
	if err != nil {
		return nil, err
	}

	response, err := requester.execute()
	if err != nil {
		return nil, err
	}

	if response.Body == nil {
		return nil, fmt.Errorf("[%v] response is nil", method+url)
	}

	defer response.Body.Close()

	bytesData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != StatusOk {
		return nil, fmt.Errorf("%s", parseMessage(bytesData))
	}

	log.Debugf("", "[%v] response info [%+v]", Protocol+"://"+url, string(bytesData))
	return bytesData, nil
}

// ResponseSuccess get success response after request url
func ResponseSuccess(method, url string, request interface{}) error {
	resp, err := RemoteCall(method, url, request)
	if err != nil {
		return err
	}

	message := parseMessage(resp)
	if message != "" {
		return fmt.Errorf("response suc is false, msg is %v", message)
	}

	return nil
}

func parseMessage(resp []byte) string {
	var response struct {
		Success bool        `json:"suc"`
		Message interface{} `json:"msg"`
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		return string(resp)
	}

	if !response.Success {
		return fmt.Sprint(response.Message)
	}

	return ""
}


