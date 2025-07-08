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
)

var (
	// Regexes for parsing Cloudflare challenge page elements
	jschlRE  = regexp.MustCompile(`name="jschl_vc" value="(\w+)"`)
	passRE   = regexp.MustCompile(`name="pass" value="(.+?)"`)
	rRE      = regexp.MustCompile(`name="r" value="(.+?)"`) // Used for POST challenges
	actionRE = regexp.MustCompile(`action="(.*?)"`)         // Used for POST challenges
	keyRE    = regexp.MustCompile("<div style=\"display:none;visibility:hidden;\" id=\".*?\">(.*?)<")

	// Regexes for extracting and manipulating Cloudflare's JavaScript challenge code
	jsChallengeRE = regexp.MustCompile(
		"setTimeout\\(function\\(\\){\\s+(var " +
			"s,t,o,p,b,r,e,a,k,i,n,g,f.+?\\r?\\n[\\s\\S]+?a\\.value =.+?)\\r?\\n",
	)
	// Regexes for cleaning up the extracted JS (common and GET specific)
	jsGenericCleanup1RE = regexp.MustCompile("\\s{3,}[a-z](?: = |\\.).+")
	jsGenericCleanup2RE = regexp.MustCompile("[\\n\\\\']")
	// Regexes for GET specific JS cleanup (formerly jsReplace3RE, jsReplace4RE)
	jsGetCleanup1RE = regexp.MustCompile(";\\s*\\d+\\s*$")
	jsGetCleanup2RE = regexp.MustCompile("a\\.value\\s*\\=")
)

// RoundTripper is a http client RoundTripper that can handle the Cloudflare anti-bot.
type RoundTripper struct {
	upstream http.RoundTripper
	cookies  http.CookieJar
}

// ParsedChallengePage holds data extracted from a Cloudflare challenge page.
type ParsedChallengePage struct {
	JschlVc string
	Pass    string
	R       string // Specific to POST challenges
	Action  string // Specific to POST challenges
	Key     string // Specific to POST challenges
	Body    string
	Host    string
	Scheme  string
	FullURL *url.URL
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

// setDefaultHeaders sets default HTTP headers if they are not already present.
func setDefaultHeaders(r *http.Request) {
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
	}
}

// RoundTrip implements the RoundTripper interface for the Transport type.
func (rt RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	setDefaultHeaders(r)
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

		challengeData, err := parseChallengeResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse challenge response: %w", err)
		}

		answerReq, err := buildAnswerRequest(challengeData, resp.Request.Header, resp.Cookies())
		if err != nil {
			return nil, fmt.Errorf("failed to build answer request: %w", err)
		}

		// Store cookies from the original response before closing its body
		// (parseChallengeResponse closes the body)
		originalRespCookies := resp.Cookies()

		resp, err = rt.upstream.RoundTrip(answerReq)
		if err != nil {
			return nil, err
		}
		// It's important to persist cookies obtained from solving the challenge
		if cookies := originalRespCookies; len(cookies) > 0 {
			rt.cookies.SetCookies(answerReq.URL, cookies)
		}
		if cookies := resp.Cookies(); len(cookies) > 0 {
			rt.cookies.SetCookies(resp.Request.URL, cookies)
		}
	}
	return resp, err
}

// parseChallengeResponse extracts necessary data from the Cloudflare challenge page.
func parseChallengeResponse(resp *http.Response) (*ParsedChallengePage, error) {
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close() // Ensure body is closed after reading
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	bodyStr := string(b)

	data := &ParsedChallengePage{
		Body:    bodyStr,
		Host:    resp.Request.URL.Host,
		Scheme:  resp.Request.URL.Scheme,
		FullURL: resp.Request.URL,
	}

	if m := jschlRE.FindStringSubmatch(bodyStr); len(m) > 0 {
		data.JschlVc = m[1]
	}
	if m := passRE.FindStringSubmatch(bodyStr); len(m) > 0 {
		data.Pass = m[1]
	}

	// Fields specific to POST challenges
	if !strings.Contains(bodyStr, "method=\"get\"") {
		if m := rRE.FindStringSubmatch(bodyStr); len(m) > 0 {
			data.R = m[1]
		} else {
			return nil, errors.New("no r found for POST challenge")
		}
		if m := actionRE.FindStringSubmatch(bodyStr); len(m) > 0 {
			data.Action = m[1]
		} else {
			return nil, errors.New("no action found for POST challenge")
		}
		if m := keyRE.FindStringSubmatch(bodyStr); len(m) > 0 {
			data.Key = m[1]
		} else {
			return nil, errors.New("no key id found for POST challenge")
		}
	}
	return data, nil
}

// buildAnswerRequest determines the method and constructs the challenge answer request.
func buildAnswerRequest(challengeData *ParsedChallengePage, originalHeaders http.Header, challengeCookies []*http.Cookie) (*http.Request, error) {
	var req *http.Request
	var err error

	if strings.Contains(challengeData.Body, "method=\"get\"") {
		req, err = buildGETAnswerRequest(challengeData)
	} else {
		req, err = buildPOSTAnswerRequest(challengeData)
	}
	if err != nil {
		return nil, err
	}

	// Copy all the header values from the original request
	if originalHeaders != nil {
		for key, vals := range originalHeaders {
			for _, val := range vals {
				req.Header[key] = []string{val} // ensure keep case sensitivity
			}
		}
	}
	req.Header.Set("Referer", challengeData.FullURL.String())
	req.Header.Set("Origin", challengeData.Scheme+"://"+challengeData.Host)

	// Add cookies obtained from the Cloudflare challenge response itself
	for _, cookie := range challengeCookies {
		req.AddCookie(cookie)
	}

	if os.Getenv("CFRT_DEBUG") != "" {
		d, _ := httputil.DumpRequest(req, true)
		fmt.Fprintln(os.Stderr, "===== [Challenge Answer Request] =====\n", string(d)+"\n\n")
	}
	return req, nil
}

// buildGETAnswerRequest constructs the GET request for submitting the Cloudflare challenge answer.
func buildGETAnswerRequest(challengeData *ParsedChallengePage) (*http.Request, error) {
	js, err := extractJSGET(challengeData.Body, challengeData.Host)
	if err != nil {
		return nil, fmt.Errorf("extractJSGET failed: %w", err)
	}

	num, err := evaluateJS(js)
	if err != nil {
		return nil, fmt.Errorf("evaluateJS failed for GET: %w", err)
	}
	answer := fmt.Sprintf("%.10f", num)

	chkURL, _ := url.Parse("/cdn-cgi/l/chk_jschl")
	u := challengeData.FullURL.ResolveReference(chkURL)

	params := make(url.Values)
	if challengeData.JschlVc == "" {
		return nil, errors.New("jschl_vc not found in challenge data for GET")
	}
	params.Set("jschl_vc", challengeData.JschlVc)

	if challengeData.Pass == "" {
		return nil, errors.New("pass not found in challenge data for GET")
	}
	params.Set("pass", challengeData.Pass)
	params.Set("jschl_answer", answer)
	u.RawQuery = params.Encode()

	return http.NewRequest("GET", u.String(), nil)
}

// buildPOSTAnswerRequest constructs the POST request for submitting the Cloudflare challenge answer.
func buildPOSTAnswerRequest(challengeData *ParsedChallengePage) (*http.Request, error) {
	if challengeData.Key == "" {
		return nil, errors.New("key not found in challenge data for POST")
	}
	if challengeData.Action == "" {
		return nil, errors.New("action not found in challenge data for POST")
	}

	jsExecutableCode, err := extractJSPOST(challengeData.Body, challengeData.Host, challengeData.Key)
	if err != nil {
		return nil, fmt.Errorf("extractJSPOST failed: %w", err)
	}

	jsResultValue, err := evaluateJS(jsExecutableCode)
	if err != nil {
		return nil, fmt.Errorf("evaluateJS failed for POST: %w", err)
	}
	answer := fmt.Sprintf("%.10f", jsResultValue)

	if challengeData.JschlVc == "" {
		return nil, errors.New("jschl_vc not found for POST challenge")
	}
	if challengeData.Pass == "" {
		return nil, errors.New("pass not found for POST challenge")
	}
	if challengeData.R == "" {
		return nil, errors.New("r not found for POST challenge")
	}

	formBody := fmt.Sprintf("r=%s&jschl_vc=%s&pass=%s&jschl_answer=%s",
		url.QueryEscape(challengeData.R),
		challengeData.JschlVc,
		challengeData.Pass,
		answer,
	)

	postURL := challengeData.Scheme + "://" + challengeData.Host + html.UnescapeString(challengeData.Action)
	req, err := http.NewRequest("POST", postURL, strings.NewReader(formBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

// extractJSGET processes the Cloudflare JavaScript challenge for GET requests.
func extractJSGET(pageBody, domain string) (string, error) {
	matches := jsChallengeRE.FindStringSubmatch(pageBody)
	if len(matches) < 2 { // Ensure the capturing group found a match
		return "", errors.New("unable to identify Cloudflare IUAM Javascript on the page (GET challenge)")
	}
	extractedJS := matches[1]

	// Substitute the current domain into the script.
	// Cloudflare's script often uses 't' as a variable for the domain.
	// The placeholder "s,t,o,p,b,r,e,a,k,i,n,g,f," is a known pattern in these scripts.
	// During local testing, the domain might be 127.0.0.1, which Cloudflare's script might not handle.
	// Replace with a real domain if testing locally.
	// TODO: Consider making this test domain configurable or removing if not broadly applicable.
	effectiveDomain := domain
	if strings.Contains(domain, "127.0.0.1") {
		effectiveDomain = "jkanime.net" // Using a more relevant example domain
	}
	processedJS := strings.Replace(extractedJS, "s,t,o,p,b,r,e,a,k,i,n,g,f,",
		"s,t = \""+effectiveDomain+"\",o,p,b,r,e,a,k,i,n,g,f,", 1)

	// Apply regex-based transformations to make the JS executable by Otto.
	// These regexes remove or alter parts of the script that are problematic for Otto
	// or are part of browser-specific behavior not replicated.
	processedJS = jsGenericCleanup1RE.ReplaceAllString(processedJS, "") // Removes lines like "a.value = t.length..."
	processedJS = jsGenericCleanup2RE.ReplaceAllString(processedJS, "") // Removes newlines and backslashes within strings
	processedJS = jsGetCleanup1RE.ReplaceAllString(processedJS, "")     // Removes trailing semicolon and number (e.g., "; 121")
	processedJS = jsGetCleanup2RE.ReplaceAllString(processedJS, "return ") // Changes "a.value = ..." to "return ..." to get the result

	if os.Getenv("CFRT_DEBUG_JS") != "" {
		fmt.Fprintf(os.Stderr, "===== [JavaScript GET Processing] =====\nOriginal Extracted JS:\n%s\n\nProcessed JS for Otto:\n%s\n", extractedJS, processedJS)
	}
	return processedJS, nil
}

// extractJSPOST processes the Cloudflare JavaScript challenge for POST requests.
func extractJSPOST(pageBody, domain, challengeKey string) (string, error) {
	matches := jsChallengeRE.FindStringSubmatch(pageBody)
	if len(matches) < 2 { // Ensure the capturing group found a match
		return "", errors.New("unable to identify Cloudflare IUAM Javascript on the page (POST challenge)")
	}
	extractedJS := matches[1]

	// Substitute the current domain into the script.
	effectiveDomain := domain
	if strings.Contains(domain, "127.0.0.1") {
		effectiveDomain = "jkanime.net" // Using a more relevant example domain
	}
	processedJS := strings.Replace(extractedJS, "s,t,o,p,b,r,e,a,k,i,n,g,f,",
		"s,t = \""+effectiveDomain+"\",o,p,b,r,e,a,k,i,n,g,f,", 1)

	// Specific regexes for POST challenge JS modifications.
	// These are based on observed patterns in Cloudflare's POST challenge scripts.
	// Their exact meaning can be obscure and tied to specific obfuscation techniques
	// used by Cloudflare at the time the original code was written.
	postSpecificCleanup1RE := regexp.MustCompile("\\s{3,}[atf](?: = |\\.).+")
	// Replaces a complex eval function with the provided challengeKey.
	postSpecificCleanup2RE := regexp.MustCompile("function\\(p\\){var p = eval\\(eval\\(e.*?; return \\+\\(p\\)}\\(\\)")
	// Replaces another function pattern with 't.charCodeAt'.
	postSpecificCleanup3RE := regexp.MustCompile("function\\(p\\){return eval\\(\\(.*?}")
	// Removes a specific trailing pattern like " '; 121'".
	postSpecificCleanup4RE := regexp.MustCompile("\\s';\\s121'$")
	// Changes "a.value = ..." to "return ..." to get the result.
	postSpecificReturnRE := regexp.MustCompile("a\\.value\\s*\\=")

	processedJS = postSpecificCleanup1RE.ReplaceAllString(processedJS, "")
	processedJS = postSpecificCleanup2RE.ReplaceAllString(processedJS, challengeKey)
	processedJS = postSpecificCleanup3RE.ReplaceAllString(processedJS, "t.charCodeAt")
	processedJS = postSpecificCleanup4RE.ReplaceAllString(processedJS, "")
	processedJS = postSpecificReturnRE.ReplaceAllString(processedJS, "return ")
	// Adding newlines after semicolons can sometimes help with debugging or Otto's parsing.
	processedJS = strings.Replace(processedJS, ";", ";\n", -1)

	if os.Getenv("CFRT_DEBUG_JS") != "" {
		fmt.Fprintf(os.Stderr, "===== [JavaScript POST Processing] =====\nOriginal Extracted JS:\n%s\n\nProcessed JS for Otto:\n%s\n", extractedJS, processedJS)
	}
	return processedJS, nil
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
