package contextValuesToHeaders

import (
	"bitbucket.org/HeilaSystems/transport/client"
	"fmt"
	"net/http"
)

func ServiceNameToHeader(serviceName string) client.HTTPClientInterceptor {
	return func(req *http.Request, handler client.HTTPHandler) (*http.Response, error) {
		req.Header.Add("Caller", fmt.Sprint(serviceName))
		res, err := handler(req)
		return res, err
	}
}
