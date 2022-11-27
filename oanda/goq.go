/*
 * Get,Post,Put等のHTTPリクエストを実行するStructを定義
 */

package oanda

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var (
	LIVE_URL = "https://api-fxtrade.oanda.com/v3"
	DEMO_URL = "https://api-fxpractice.oanda.com/v3"
)

type (
	Goquest struct {
		Auth   *apiKey
		Client *http.Client
		url    string
	}
)

type strMap map[string]string
type iMap map[string]interface{}

// Oanda-API実行用のハンドラを返す。
// fpath:APIキー等が入ったファイル({"live":{"id":string,"token":string},"demo":{"id":string,"token":string}})。
// mode: "live" ->本番 "demo" ->　デモ。
func NewGoquest(fpath string, mode string) *Goquest {
	host := ""
	if mode == "live" {
		host = LIVE_URL
	} else if mode == "demo" {
		host = DEMO_URL
	}
	return &Goquest{
		Auth:   newApiKey(fpath, mode),
		Client: &http.Client{},
		url:    host,
	}
}

// paramをuriにエンコードし、フルurlを返す。
// urlを生成するので、goq.Getの時にだけ使います。
func (g *Goquest) genUrl(ep string, param strMap) string {
	query := url.Values{}
	for k, v := range param {
		query.Add(k, v)
	}
	url := g.url + ep + "?" + query.Encode()
	return url
}

// 認証authをヘッダにセット
func (g *Goquest) auth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+g.Auth.Token)
}

// Content-Typeをヘッダにセット
func (g *Goquest) contenType(req *http.Request, mime string) {
	req.Header.Set("Content-Type", mime)
}

// GET 実行。Responseはjsonを想定。responseのbodyをinterfaceにpopulateする。
func (goq *Goquest) Get(ep string, param strMap, i Checker) {
	uri := goq.genUrl(ep, param)
	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		fmt.Println(err)
	}

	goq.auth(req)

	res, err := goq.Client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(b, i)
	if err != nil {
		fmt.Println(err)
	}
	// Check interfaceにステータスコードを設定
	i.Status(res.StatusCode)

}

func (goq *Goquest) Post(ep string, param iMap, i Checker) {
	goq.exec("POST", ep, param, i)
}

func (goq *Goquest) Put(ep string, param iMap, i Checker) {
	goq.exec("PUT", ep, param, i)
}

func (goq *Goquest) exec(method string, ep string, param iMap, i Checker) {
	uri := goq.genUrl(ep, nil)
	body, err := json.Marshal(param)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(uri)
	req, err := http.NewRequest(method, uri, strings.NewReader(string(body)))
	if err != nil {
		fmt.Println(err)
	}

	goq.auth(req)
	// add content-type
	goq.contenType(req, "application/json")

	res, err := goq.Client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(resBody, i)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(res.Status)
	i.Status(res.StatusCode)
}
