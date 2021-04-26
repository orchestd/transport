package contextValuesToHeaders

import (
	"fmt"
	"net/http"
	"bitbucket.org/HeilaSystems/transport/client"

)

func TokenClientInterceptors() client.HTTPClientInterceptor {
	return func(req *http.Request, handler client.HTTPHandler) (*http.Response, error) {
		if val := req.Context().Value("token");val != nil {
			req.Header.Add("Token",	fmt.Sprint(val))
		}
		res, err := handler(req)
		return res, err
	}
}