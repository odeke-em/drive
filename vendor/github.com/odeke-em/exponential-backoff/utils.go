package expb

import (
	"net/http"
)

func tryGet(uri string) Producer {
	return func() (interface{}, error) {
		return http.Get(uri)
	}
}

func httpStatus(v interface{}) (ok, retryable bool) {
	res := v.(*http.Response)

	if res == nil {
		return false, false
	}

	statusCode := res.StatusCode
	if statusCode >= 200 && statusCode <= 299 {
		ok = true
		return
	}
	if statusCode == http.StatusForbidden {
		retryable = true
		return
	}
	return
}

func NewUrlGetter(url string, retryCount uint32) *ExponentialBacker {
	req := tryGet(url)
	return &ExponentialBacker{
		Do:          req,
		StatusCheck: httpStatus,
		RetryCount:  retryCount,
	}
}
