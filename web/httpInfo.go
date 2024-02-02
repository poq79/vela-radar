package web

import (
	"bytes"
	"crypto/md5"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vela-ssoc/vela-radar/port"
	"github.com/vela-ssoc/vela-radar/web/finder"
)

//go:embed finger.json
var FingerData []byte

const (
	refusedStr   = "refused"
	ioTimeoutStr = "i/o timeout"
)

var httpClient *http.Client

func ProbeHttpInfo(url2 string, dialTimeout time.Duration) (httpInfo *port.HttpInfo, isDailErr bool) {

	if httpClient == nil {
		httpClient = newHttpClient(dialTimeout)
	}

	var err error
	var rewriteUrl string
	var body []byte
	var resp *http.Response

	var rewriteNum int
goReq:
	resp, body, err = getReq(url2)
	if err != nil {
		if strings.HasSuffix(err.Error(), ioTimeoutStr) || strings.Contains(err.Error(), refusedStr) {
			return nil, true
		}
		return nil, false
	}

	if resp != nil {
		if resp.ContentLength == -1 {
			resp.ContentLength = int64(len(body))
		}
		rewriteUrl2, _ := resp.Location()
		if rewriteUrl2 != nil {
			rewriteUrl = rewriteUrl2.String()
		} else {
			rewriteUrl = ""
		}
		location := GetLocation(body)
		if rewriteUrl == "" && location != "" {
			rewriteUrl = location
		}
		if location != "" && rewriteNum < 3 {
			if !strings.HasPrefix(location, "http") {
				location = resp.Request.URL.String() + location
			}
			url2 = location
			rewriteNum++
			goto goReq
		}
		//
		httpInfo = new(port.HttpInfo)
		httpInfo.URL = resp.Request.URL.String()
		httpInfo.StatusCode = resp.StatusCode
		httpInfo.ContentLength = int(resp.ContentLength)
		httpInfo.Location = rewriteUrl
		httpInfo.Server = resp.Header.Get("Server")
		httpInfo.Title = ExtractTitle(body)
		httpInfo.Body = string(body)
		httpInfo.Header = getHeadersString(resp)
		if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
			httpInfo.TLSCommonName = resp.TLS.PeerCertificates[0].Subject.CommonName
			httpInfo.TLSDNSNames = resp.TLS.PeerCertificates[0].DNSNames
		}
		// finger
		// err = finder.ParseWebFingerData(FingerData)
		if err == nil {
			resp.Body = io.NopCloser(bytes.NewReader(body))
			httpInfo.Fingerprints = finder.WebFingerIdent(resp)
			// favicon
			fau := finder.FindFaviconUrl(string(body))
			if fau != "" {
				if !strings.HasPrefix(fau, "http") {
					u := resp.Request.URL.String()
					if strings.HasSuffix("/", u) || strings.HasPrefix("/", fau) {
						fau = u[:len(u)-1] + fau
					} else {
						fau = resp.Request.URL.String() + fau
					}
				}
				_, body2, err2 := getReq(fau)
				httpInfo.FaviconMH3 = finder.Mmh3Hash32(finder.StandBase64(body2))
				httpInfo.FaviconMD5 = fmt.Sprintf("%x", md5.Sum(body2))
				if err2 == nil && len(body2) != 0 {
					httpInfo.Fingerprints = append(httpInfo.Fingerprints, finder.WebFingerIdentByFavicon_mh3(httpInfo.FaviconMH3)...)
				}
			}
		}

		if resp.StatusCode != 400 {
			return httpInfo, true
		}
	}

	return httpInfo, false
}

func getReq(url2 string) (resp *http.Response, body []byte, err error) {
	req, _ := http.NewRequest(http.MethodGet, url2, http.NoBody)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Close = true // disable keepalive
	resp, err = httpClient.Do(req)
	if err != nil {
		return
	}
	if resp.Body != http.NoBody && resp.Body != nil {
		body, _ = getBody(resp)
		if contentTypes, _ := resp.Header["Content-Type"]; len(contentTypes) > 0 {
			if strings.Contains(contentTypes[0], "text") {
				_body, err2 := DecodeData(body, resp.Header)
				if err2 == nil {
					body = _body
				}
				resp.Body = io.NopCloser(bytes.NewReader(body))
			}
		}
	}
	return
}

func getHeadersString(resp *http.Response) string {
	// Get the header map
	headers := resp.Header

	// Convert headers to a string
	var headerString string
	for key, values := range headers {
		for _, value := range values {
			headerString += fmt.Sprintf("%s: %s\n", key, value)
		}
	}
	return headerString
}
