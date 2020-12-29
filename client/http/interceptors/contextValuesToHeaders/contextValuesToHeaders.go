package contextValuesToHeaders

import (
	"bitbucket.org/HeilaSystems/transport/client"
	"fmt"
	"net/http"
)

func ContextValuesToHeaders(extractedHeaders []string) client.HTTPClientInterceptor {
	return func(req *http.Request, handler client.HTTPHandler) (*http.Response, error) {
		for _, header := range extractedHeaders {
			req.Header.Add(header,	fmt.Sprint(req.Context().Value(header)))
		}
		res, err := handler(req)
		return res, err
	}
}