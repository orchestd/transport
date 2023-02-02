package contextValuesToHeaders

import (
	"fmt"
	"github.com/orchestd/transport/client"
	"net/http"
)

func ContextValuesToHeaders(extractedHeaders []string) client.HTTPClientInterceptor {
	return func(req *http.Request, handler client.HTTPHandler) (*http.Response, error) {
		for _, header := range extractedHeaders {
			if val := req.Context().Value(header); val != nil {
				req.Header.Add(header, fmt.Sprint(val))
			}
		}
		res, err := handler(req)
		return res, err
	}
}
