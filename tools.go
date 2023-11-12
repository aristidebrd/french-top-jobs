package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// This function takes an url as input, checks it is responding, and then returns back a bool according to the result
func isUrlWorking(url string) bool {
	client := http.Client{}
	_, err := client.Get(url)
	if err != nil {
		return false
	} else {
		return true
	}
}

// This function takes an url as input that might have trailing slashes at it's end, and return it back without them
func removeTrailingSlash(urlStr string) string {
	urlStr = strings.TrimSpace(urlStr)
	u, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	if len(u.Path) > 0 && u.Path[len(u.Path)-1] == '/' {
		u.Path = u.Path[:len(u.Path)-1]
	}

	return u.String()
}

// This function returns the absolute url of the href given as parameter combined with it's website url
func getAbsoluteUrl(website string, href string) string {
	attrUrl, _ := url.Parse(href)
	isAbsURL := attrUrl.IsAbs()
	if !isAbsURL {
		href = website + href
	}
	return href
}
