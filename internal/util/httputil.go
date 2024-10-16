package util

import (
	"bytes"
	"net/http"
	"net/url"
	"strconv"
)

func SendGet(url string, header map[string]string, wrap func(response *http.Response) (interface{}, error)) (interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	host, err := hostname(url)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Host", host)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return wrap(resp)
}

func SendPost(url string, header map[string]string, body []byte, wrap func(response *http.Response) (interface{}, error)) (interface{}, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	host, err := hostname(url)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Host", host)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return wrap(resp)
}

func hostname(u string) (string, error) {
	e, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return e.Hostname(), nil
}
