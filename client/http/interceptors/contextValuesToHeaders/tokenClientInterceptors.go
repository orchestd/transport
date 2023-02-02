package contextValuesToHeaders

import (
	"fmt"
	"github.com/orchestd/transport/client"
	"net/http"
)

func TokenClientInterceptors() client.HTTPClientInterceptor {
	return func(req *http.Request, handler client.HTTPHandler) (*http.Response, error) {
		if val := req.Context().Value("token"); val != nil {
			req.Header.Add("Token", fmt.Sprint(val))
		}
		res, err := handler(req)
		return res, err
	}
}
