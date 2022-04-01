package lib

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Resp struct {
	// 状态码
	StatusCode int

	// 响应包体的长度
	ContentLength int64

	// 响应的body
	Content string
}

// 三个参数
// - protocol: http | https
// - ip
// - host
func SendRequest(protocol, ip, host string) (Resp, error) {

	// 跳过TLS
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, time.Second*3)
			if err != nil {
				return nil, err
			}
			conn.SetDeadline(time.Now().Add(time.Second * 2))
			return conn, nil
		},
		ResponseHeaderTimeout: time.Second * 2,
	}

	// 设置5秒超时
	client := &http.Client{Transport: tr, Timeout: time.Second * 5}

	url := fmt.Sprintf("%s://%s", protocol, ip)

	// 设置Header
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Resp{}, err
	}

	req.Host = host
	req.Header.Add("Host", host)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")

	// 不进行重定向跳转
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// 获得返回结果
	resp, err := client.Do(req)
	if err != nil {
		return Resp{}, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Resp{}, err
	}

	response := Resp{resp.StatusCode, resp.ContentLength, string(content)}

	return response, nil
}
