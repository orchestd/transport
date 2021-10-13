package http

import (
	"bitbucket.org/HeilaSystems/dependencybundler/interfaces/configuration"
	. "bitbucket.org/HeilaSystems/servicereply"
	"bitbucket.org/HeilaSystems/servicereply/status"
	"bitbucket.org/HeilaSystems/transport/client"
	"bitbucket.org/HeilaSystems/transport/discoveryService"
	"bytes"
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type httpClientWrapper struct {
	client                   *http.Client
	conf                     configuration.Config
	discoveryServiceProvider discoveryService.DiscoveryServiceProvider
}

func (h *httpClientWrapper) Call(c context.Context, payload interface{}, host, handler string, target interface{}, headers map[string]string) ServiceReply {
	return h.do(c, http.MethodPost, payload, host, handler, target, headers, true)
}

func (h *httpClientWrapper) Post(c context.Context, payload interface{}, host, handler string, target interface{}, headers map[string]string) ServiceReply {
	return h.do(c, http.MethodPost, payload, host, handler, target, headers, false)
}

func (h *httpClientWrapper) Get(c context.Context, host, handler string, target interface{}, headers map[string]string) ServiceReply {
	return h.do(c, http.MethodGet, nil, host, handler, target, headers, false)
}

func (h *httpClientWrapper) Put(c context.Context, payload interface{}, host, handler string, target interface{}, headers map[string]string) ServiceReply {
	return h.do(c, http.MethodPut, payload, host, handler, target, headers, false)
}

func (h *httpClientWrapper) Delete(c context.Context, host, handler string, target interface{}, headers map[string]string) ServiceReply {
	return h.do(c, http.MethodDelete, nil, host, handler, target, headers, false)
}

func (h *httpClientWrapper) SetDiscoveryServiceProvider(dsp discoveryService.DiscoveryServiceProvider) {
	h.discoveryServiceProvider = dsp
}

func NewHttpClientWrapper(client *http.Client, conf configuration.Config) (client.HttpClient, error) {
	return &httpClientWrapper{client: client, conf: conf}, nil
}

func (h *httpClientWrapper) do(c context.Context, httpMethod string, payload interface{}, host, handler string, target interface{}, headers map[string]string, internal bool) (srvReply ServiceReply) {
	var url string
	if sRep := h.discoveryServiceProvider.GetAddress(host); !sRep.IsSuccess() {
		return sRep
	} else if v, ok := sRep.GetReplyValues()["address"]; !ok {
		return sRep.WithError(fmt.Errorf("cant resolve host:%s", host))
	} else {
		url = fmt.Sprintf("%s/%s", v, handler)
	}

	srvReply = NewNil()
	b, sErr := getPayload(payload, url)
	if sErr != nil {
		return sErr
	}
	var req *http.Request
	var err error
	if b != nil {
		req, err = http.NewRequest(httpMethod, url, b)
	} else {
		req, err = http.NewRequest(httpMethod, url, nil)
	}
	if err != nil {
		return NewInternalServiceError(err).WithLogMessage(fmt.Sprintf("Cannot marshal request to %s", url))
	}
	req = req.WithContext(c)
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	if _, ok := headers["Content-Type"]; !ok {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return NewIoError(err).WithLogMessage(fmt.Sprintf("couldn't send %s request to %s", httpMethod, url))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return NewInternalServiceError(err).WithLogMessage(fmt.Sprintf("cannot read response from %s", url))
	}
	if internal {
		var srvError Response
		if err := json.Unmarshal(body, &srvError); err != nil {
			return NewInternalServiceError(err).WithLogMessage(fmt.Sprintf("cannot read response from %s", url)).WithLogValues(ValuesMap{"rawResponse": string(body)})
		}
		if srvError.Status != status.SuccessStatus {
			resType := status.GetTypeByStatus(srvError.GetStatus())
			msgValues := srvError.GetMessageValues()
			srvReply = NewServiceError(&resType, nil, srvError.GetMessageId(), 1)
			if msgValues != nil {
				srvReply = srvReply.WithReplyValues(*msgValues)
			}
			return srvReply
		}
		if srvError.Message != nil {
			msgValues := srvError.GetMessageValues()
			srvReply = NewMessage(srvError.GetMessageId())
			if msgValues != nil {
				srvReply = srvReply.WithReplyValues(*msgValues)
			}
		}
		if srvError.Data != nil {
			if dataJson, err := json.Marshal(srvError.Data); err != nil {
				return NewInternalServiceError(err).WithLogMessage(fmt.Sprintf("cannot marshal data from %s", url))
			} else {
				body = dataJson
			}
		} else {
			body = nil
		}
	}
	if err := unmarshalDataToStruct(body, target, url); err != nil {
		return err
	}
	return
}

type Marshaler interface {
	Unmarshal(data []byte) error
}

func unmarshalDataToStruct(data []byte, target interface{}, logStrings ...interface{}) ServiceReply {
	if target == nil {
		return nil
	}
	if data == nil {
		return NewInternalServiceError(nil).WithLogMessage(fmt.Sprintf("Cannot unmarshal empty response from %s to target struct", logStrings...))
	}
	if f, ok := target.(Marshaler); ok {
		if err := f.Unmarshal(data); err != nil {
			return NewInternalServiceError(err)
		}
		return nil
	}
	if err := json.Unmarshal(data, &target); err != nil {
		return NewInternalServiceError(err).WithLogMessage("cannot read response")
	}
	return nil
}
func getPayload(payload interface{}, url string) (*bytes.Buffer, ServiceReply) {
	if payload != nil {
		request, err := json.Marshal(payload)
		if err != nil {
			return nil, NewInternalServiceError(err).WithLogMessage(fmt.Sprintf("cannot read response from %s", url)).WithLogValues(ValuesMap{"rawResponse": payload})
		}
		return bytes.NewBuffer(request), nil
	}
	return nil, nil
}
