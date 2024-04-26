package ghttp

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func GetKeyFromHeaders(headers http.Header, key string) string {
	var result string

	resultHeader := getValueIgnoringCase(headers, key)
	if resultHeader != nil {
		result = resultHeader[0]
	}

	return result
}

func getValueIgnoringCase(headers map[string][]string, key string) []string {
	keyLower := strings.ToLower(key)

	for k, v := range headers {
		if strings.ToLower(k) == keyLower {
			return v
		}
	}

	return nil
}

func ReplaceValuesInURL(url string, params map[string]string) string {
	regex := regexp.MustCompile(`\{([^}]+)\}`)

	replacedURL := regex.ReplaceAllStringFunc(url, func(match string) string {
		key := match[1 : len(match)-1]
		replacement, exists := params[key]

		if exists {
			return replacement
		}

		return match
	})

	return replacedURL
}

func ExtractUUIDFromURL(rawURL, key string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	segments := strings.Split(u.Path, "/")
	for i, segment := range segments {
		if segment == key && i+1 < len(segments) {
			uuid := segments[i+1]
			return uuid, nil
		}
	}

	return "", fmt.Errorf("UUID not found after the key: %s", key)
}
