package main

import (
	"encoding/base64"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/bclicn/color"
)

type Rss struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Item        []Item `xml:"item"`
}

type Item struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
}

func main() {
	defaultUrl := "https://news.ycombinator.com/rss"

	var hostArg,
		proxyArg,
		proxyUsrArg,
		proxyPwdArg string

	var proxy *url.URL

	flag.StringVar(&hostArg, "host", defaultUrl, "The URI of the Rss feed. Eg. ["+defaultUrl+"]")
	flag.StringVar(&proxyArg, "proxy", "", "The Proxy URL, optional")
	flag.StringVar(&proxyUsrArg, "proxy_usr", "", "The Proxy username, optional. Required proxy_pwd.")
	flag.StringVar(&proxyPwdArg, "proxy_pwd", "", "The Proxy password, optional. Required proxy_usr.")
	flag.Parse()

	if hostArg == "" {
		hostArg = defaultUrl
	}

	//TODO validate proxy_usr and proxy_pwd

	// parse the host
	hostURL, err := url.Parse(hostArg)
	checkError(err)

	// parse the proxy if existing
	if len(proxyArg) > 0 {
		proxy, err = url.Parse(proxyArg)
		checkError(err)
	}

	request, err := http.NewRequest("GET", hostURL.String(), nil)

	if proxy != nil && proxy.String() != "" && proxyUsrArg != "" && proxyPwdArg != "" {
		auth := proxyUsrArg + ":" + proxyPwdArg
		basic := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
		request.Header.Add("Proxy-Authentication", basic)
		dump, _ := httputil.DumpRequest(request, false)
		fmt.Println(string(dump))
	}

	transport := &http.Transport{Proxy: http.ProxyURL(proxy)}
	client := &http.Client{Transport: transport}

	response, err := client.Do(request)
	checkError(err)

	if response.Status != "200 OK" {
		fmt.Println(response.Status)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(response.Body)
	checkError(err)
	response.Body.Close()

	var rssData Rss

	err2 := xml.Unmarshal([]byte(string(body)), &rssData)
	checkError(err2)

	for i := 0; i < len(rssData.Channel.Item); i++ {
		fmt.Println(color.BRed(rssData.Channel.Item[i].Title))
		fmt.Printf("> %s\n", color.Underline(rssData.Channel.Item[i].Link))
	}

	os.Exit(0)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
		os.Exit(1)
	}
}
