package handlers

import (
	"net/http"
	"net/url"
)

func sanitizeRequest(req *http.Request) *http.Request {
	req.URL = sanitizeURL(req.URL)
	if req.Method == "" {
		req.Method = "GET"
	}
	if req.Host == "" {
		req.Host = req.URL.Host
	} else if req.URL.Host == "" {
		req.URL.Host = req.Host
	}
	return req
}

func sanitizeURL(u *url.URL) *url.URL {
	if u.Path == "" {
		u.Path = "/"
	}
	if u.Scheme == "" {
		port := u.Port()
		if port == "" || port != "443" {
			u.Scheme = "http"
		} else {
			u.Scheme = "https"
		}
	}
	return u
}

func mergeRequests(target *http.Request, sources ...*http.Request) *http.Request {
	for _, source := range sources {
		target.Method = source.Method
		target.URL = mergeURLs(target.URL, source.URL)
		target.Header = source.Header
		target.Body = source.Body
	}
	return target
}

func mergeURLs(target *url.URL, sources ...*url.URL) *url.URL {
	for _, source := range sources {
		if source.Host != "" {
			target.Host = source.Host
		}
		if source.Scheme != "" {
			target.Scheme = source.Scheme
		}
		if source.Path != "" {
			target.Path = source.Path
			target.RawPath = source.RawPath
			target.RawQuery = source.RawQuery
			target.ForceQuery = source.ForceQuery
		}
	}
	return target
}
