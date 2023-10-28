package _function

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func Fetch[T any](client *http.Client, _url string, _method string, _body string, _headers map[string]string, responseTemplate T) (*T, error) {
	var body io.Reader
	if strings.ToUpper(_method) == "POST" {
		body = strings.NewReader(_body)
	} else {
		body = nil
	}
	req, err := http.NewRequest(_method, _url, body)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 9; ONEPLUS A3010 Build/PKQ1.181203.001; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/117.0.0.0 Mobile Safari/537.36 tieba/12.22.1.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for k, v := range _headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// fmt.Println(string(response[:]))

	if err = JsonDecode(response, &responseTemplate); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &responseTemplate, err
}
