// Copyright 2018 Jeff Foley. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package config

import (
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
)

const (
	// UserAgent is the default user agent used by HTTP requests.
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36"
	//UserAgent = "Mozilla/5.0 (Linux; Android 7.1.1; CPH1609) AppleWebKit/537.36 (KHTML, like Gecko) coc_coc_browser/78.0.142 Mobile Chrome/72.0.3626.142 Mobile Safari/537.36"
)

var (
	jschlRE  = regexp.MustCompile(`name="jschl_vc" value="(\w+)"`)
	passRE   = regexp.MustCompile(`name="pass" value="(.+?)"`)
	rRE      = regexp.MustCompile(`name="r" value="(.+?)"`)
	actionRE = regexp.MustCompile(`action="(.*?)"`)

	keyRE = regexp.MustCompile("<div style=\"display:none;visibility:hidden;\" id=\".*?\">(.*?)<")
	/*jsRE    = regexp.MustCompile(
		`setTimeout\(function\(\){\s+(var ` +
			`s,t,o,p,b,r,e,a,k,i,n,g,f.+?\r?\n[\s\S]+?a\.value =.+?)\r?\n`,
	)
	jsReplace1RE = regexp.MustCompile(`a\.value = (.+ \+ t\.length).+`)
	jsReplace2RE = regexp.MustCompile(`\s{3,}[a-z](?: = |\.).+`)
	jsReplace3RE = regexp.MustCompile(`[\n\\']`)
	*/
	jsRE = regexp.MustCompile(
		"setTimeout\\(function\\(\\){\\s+(var " +
			"s,t,o,p,b,r,e,a,k,i,n,g,f.+?\\r?\\n[\\s\\S]+?a\\.value =.+?)\\r?\\n",
	)
	jsReplace1RE = regexp.MustCompile("\\s{3,}[a-z](?: = |\\.).+")
	jsReplace2RE = regexp.MustCompile("[\\n\\\\']")
	jsReplace3RE = regexp.MustCompile(";\\s*\\d+\\s*$")
	jsReplace4RE = regexp.MustCompile("a\\.value\\s*\\=")
)

// RoundTripper is a http client RoundTripper that can handle the Cloudflare anti-bot.
type RoundTripper struct {
	upstream http.RoundTripper
	cookies  http.CookieJar
}

// New wraps a http client transport with one that can handle the Cloudflare anti-bot.
func New(upstream http.RoundTripper) (*RoundTripper, error) {
	if upstream == nil {
		upstream = &http.Transport{}
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &RoundTripper{upstream, jar}, nil
}

// RoundTrip implements the RoundTripper interface for the Transport type.
func (rt RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", UserAgent)
	}
	if r.Header.Get("Accept") == "" {
		r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	}
	if r.Header.Get("Accept-Language") == "" {
		r.Header.Set("Accept-Language", "en-US,en;q=0.9")
	}
	if r.Header.Get("Accept-Encoding") == "" {
		r.Header.Set("Accept-Encoding", "gzip, deflate")
	}
	if r.Header.Get("DNT") == "" {
		r.Header["DNT"] = []string{"1"}
		//r.Header.Add("Dnt", "1")
	}
	/*if r.Header.Get("Upgrade-Insecure-Requests") == "" {
		r.Header.Set("Upgrade-Insecure-Requests", "1")
	}*/
	/*if r.Header.Get("Connection") == "" {
		r.Header.Set("Connection", "keep-alive")
	}*/
	// Pass along Cloudflare cookies obtained previously
	for _, cookie := range rt.cookies.Cookies(r.URL) {
		r.AddCookie(cookie)
	}

	if os.Getenv("CFRT_DEBUG") != "" {
		d, _ := httputil.DumpRequest(r, true)
		fmt.Fprintln(os.Stderr, "===== [DUMP Request] =====\n", string(d))
	}
	resp, err := rt.upstream.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	if os.Getenv("CFRT_DEBUG") != "" {
		d, _ := httputil.DumpResponse(resp, false)
		fmt.Fprintln(os.Stderr, "===== [DUMP Response] =====\n", string(d))
	}
	// Check if the Cloudflare anti-bot has prevented the request
	if resp.StatusCode == 503 && strings.HasPrefix(resp.Header.Get("Server"), "cloudflare") {
		// Cloudflare requires a delay before solving the challenge
		time.Sleep(5 * time.Second)
		if cookies := resp.Cookies(); len(cookies) > 0 {
			rt.cookies.SetCookies(resp.Request.URL, resp.Cookies())
		}
		req, err := buildAnswerRequest(resp)
		if err != nil {
			return nil, err
		}
		resp, err = rt.upstream.RoundTrip(req)
		if err != nil {
			return nil, err
		}
	}
	return resp, err
}

func buildAnswerRequest(resp *http.Response) (*http.Request, error) {
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var req *http.Request
	// old GET method
	if strings.Contains(string(b), "method=\"get\"") {
		js, err := extractJSGET(string(b), resp.Request.URL.Host)
		if err != nil {
			return nil, err
		}
		// Obtain the answer from the JavaScript challenge
		num, err := evaluateJS(js)
		answer := fmt.Sprintf("%.10f", num)
		if err != nil {
			return nil, err
		}
		// Begin building the URL for submitting the answer
		chkURL, _ := url.Parse("/cdn-cgi/l/chk_jschl")
		u := resp.Request.URL.ResolveReference(chkURL)
		// Obtain all the parameters for the URL
		var params = make(url.Values)
		if m := jschlRE.FindStringSubmatch(string(b)); len(m) > 0 {
			params.Set("jschl_vc", m[1])
		}
		if m := passRE.FindStringSubmatch(string(b)); len(m) > 0 {
			params.Set("pass", m[1])
		}
		params.Set("jschl_answer", answer)
		u.RawQuery = params.Encode()

		req, err = http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}

		// new POST request
	} else {
		// get hidden key from page
		key := ""
		if m := keyRE.FindStringSubmatch(string(b)); len(m) > 0 {
			//fmt.Printf("x: %s\n\n",x[1])
			key = m[1]
		} else {
			return nil, errors.New("no key id found")
		}

		action := ""
		if m := actionRE.FindStringSubmatch(string(b)); len(m) > 0 {
			//fmt.Printf("x: %s\n\n",x[1])
			action = m[1]
		} else {
			return nil, errors.New("no key id found")
		}

		js, err := extractJSPOST(string(b), resp.Request.URL.Host, key)
		if err != nil {
			return nil, err
		}
		// Obtain the answer from the JavaScript challenge
		num, err := evaluateJS(js)
		answer := fmt.Sprintf("%.10f", num)
		if err != nil {
			return nil, err
		}

		// Obtain all the parameters for the URL
		var params = make(url.Values)
		if m := jschlRE.FindStringSubmatch(string(b)); len(m) > 0 {
			params.Set("jschl_vc", m[1])
		} else {
			return nil, errors.New("no jschl_vc found")
		}
		if m := passRE.FindStringSubmatch(string(b)); len(m) > 0 {
			params.Set("pass", m[1])
		} else {
			return nil, errors.New("no pass found")
		}
		if m := rRE.FindStringSubmatch(string(b)); len(m) > 0 {
			params.Set("r", m[1])
		} else {
			return nil, errors.New("no r found")
		}
		params.Set("jschl_answer", answer)

		body := fmt.Sprintf("r=%s&jschl_vc=%s&pass=%s&jschl_answer=%s",
			url.QueryEscape(params.Get("r")),
			params.Get("jschl_vc"),
			params.Get("pass"),
			params.Get("jschl_answer"),
		)

		//req, err = http.NewRequest("POST", resp.Request.URL.Scheme + "://" + resp.Request.URL.Host + html.UnescapeString(action), strings.NewReader(params.Encode()))
		req, err = http.NewRequest("POST",
			resp.Request.URL.Scheme+"://"+resp.Request.URL.Host+html.UnescapeString(action),
			strings.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	}
	// Copy all the header values from the original request
	if resp.Request.Header != nil {
		for key, vals := range resp.Request.Header {
			for _, val := range vals {
				// ensure keep case sensitivity
				req.Header[key] = []string{val}
				//req.Header.Add(key, val)
			}
		}
	}
	//req.Header.Set("Referer", resp.Request.URL.Scheme + "://" + resp.Request.URL.Host + resp.Request.URL.Path)
	req.Header.Set("Referer", resp.Request.URL.String())
	req.Header.Set("Origin", resp.Request.URL.Scheme+"://"+resp.Request.URL.Host)
	// send the cookies obtained from the Cloudflare challenge
	for _, cookie := range resp.Cookies() {
		req.AddCookie(cookie)
	}
	if os.Getenv("CFRT_DEBUG") != "" {
		d, _ := httputil.DumpRequest(req, true)
		fmt.Fprintln(os.Stderr, "===== [Challenge Answer Request] =====\n", string(d)+"\n\n")
	}
	return req, nil
}

func extractJSGET(body, domain string) (string, error) {
	matches := jsRE.FindStringSubmatch(body)
	if len(matches) == 0 {
		return "", errors.New("Unable to identify Cloudflare IUAM Javascript on the page")
	}

	// check if we're i testing mode and overwrite localhost value
	if strings.Contains(domain, "127.0.0.1") {
		domain = "torrentz2.eu"
	}

	js := matches[1]
	js = strings.Replace(js, "s,t,o,p,b,r,e,a,k,i,n,g,f,", "s,t = \""+domain+"\",o,p,b,r,e,a,k,i,n,g,f,", 1)
	js = jsReplace1RE.ReplaceAllString(js, "")
	js = jsReplace2RE.ReplaceAllString(js, "")
	js = jsReplace3RE.ReplaceAllString(js, "")
	js = jsReplace4RE.ReplaceAllString(js, "return ")
	if os.Getenv("CFRT_DEBUG_JS") != "" {
		fmt.Fprintln(os.Stderr, "===== [JavaScript GET] =====\n\n\n", js)
	}
	return js, nil
}

func extractJSPOST(body, domain string, key string) (string, error) {
	matches := jsRE.FindStringSubmatch(body)
	if len(matches) == 0 {
		return "", errors.New("Unable to identify Cloudflare IUAM Javascript on the page")
	}

	// check if we're i testing mode and overwrite localhost value
	if strings.Contains(domain, "127.0.0.1") {
		domain = "torrentz2.eu"
	}

	// extract and set domain
	js := matches[1]
	js = strings.Replace(js, "s,t,o,p,b,r,e,a,k,i,n,g,f,", "s,t = \""+domain+"\",o,p,b,r,e,a,k,i,n,g,f,", 1)

	re2 := regexp.MustCompile("\\s{3,}[atf](?: = |\\.).+")
	re31 := regexp.MustCompile("function\\(p\\){var p = eval\\(eval\\(e.*?; return \\+\\(p\\)}\\(\\)")
	re32 := regexp.MustCompile("function\\(p\\){return eval\\(\\(.*?}")
	re4 := regexp.MustCompile("\\s';\\s121'$")
	re5 := regexp.MustCompile("a\\.value\\s*\\=")

	js = re2.ReplaceAllString(js, "")
	js = re31.ReplaceAllString(js, key)
	js = re32.ReplaceAllString(js, "t.charCodeAt")
	js = re4.ReplaceAllString(js, "")
	js = re5.ReplaceAllString(js, "return ")
	js = strings.Replace(js, ";", ";\n", -1)

	if os.Getenv("CFRT_DEBUG_JS") != "" {
		fmt.Fprintln(os.Stderr, "===== [JavaScript POST] =====\n\n\n", js)
	}
	return js, nil
}

type ottoReturn struct {
	Result float64
	Err    error
}

var errHalt = errors.New("Stop")

func evaluateJS(js string) (float64, error) {
	var err error
	var result float64
	interrupt := make(chan func())
	ret := make(chan *ottoReturn)
	t := time.NewTimer(5 * time.Second)
	defer t.Stop()

	go executeUnsafeJS(js, interrupt, ret)
loop:
	for {
		select {
		case <-t.C:
			interrupt <- func() {
				panic(errHalt)
			}
		case r := <-ret:
			result = r.Result
			err = r.Err
			break loop
		}
	}
	return result, err
}

func executeUnsafeJS(js string, interrupt chan func(), ret chan *ottoReturn) {
	var num float64

	vm := otto.New()
	vm.Interrupt = interrupt

	defer func() {
		if caught := recover(); caught != nil {
			if caught == errHalt {
				ret <- &ottoReturn{
					Result: num,
					Err:    errors.New("The unsafe Javascript ran for too long"),
				}
				return
			}
			panic(caught)
		}
	}()

	//result, err := vm.Run(js)
	result, err := vm.Eval("(function () {" + js + "})()")
	if err == nil {
		num, err = result.ToFloat()
	}
	ret <- &ottoReturn{
		Result: num,
		Err:    err,
	}
}
