package http

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/orchestd/dependencybundler/interfaces/configuration"
	. "github.com/orchestd/servicereply"
	"github.com/orchestd/servicereply/status"
	"github.com/orchestd/transport/client"
	"github.com/orchestd/transport/discoveryService"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	ContentTypeJSON = "json"
	ContentTypeXML  = "xml"
)

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

func (h *httpClientWrapper) PostForm(c context.Context, uri string, postData, headers map[string]string) ([]byte, ServiceReply) {
	return h.doPostForm(c, uri, postData, headers)
}

func (h *httpClientWrapper) Get(c context.Context, host, handler string, target interface{}, headers map[string]string) ServiceReply {
	return h.do(c, http.MethodGet, nil, host, handler, target, headers, false)
}

func (h *httpClientWrapper) ExternalGet(c context.Context, host, handler string, payload map[string]string, target interface{},
	headers map[string]string, contentType string) ServiceReply {
	var payloadURLEncoded *string
	if payload != nil {
		v := url.Values{}
		for p := range payload {
			v.Add(p, payload[p])
		}
		t := v.Encode()
		payloadURLEncoded = &t
	}
	return h.doFull(c, http.MethodGet, payloadURLEncoded, host, handler, target, headers, false, contentType)
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

func (h *httpClientWrapper) doPostForm(c context.Context, uri string, postData, headers map[string]string) ([]byte, ServiceReply) {
	data := url.Values{}
	for k, v := range postData {
		data.Set(k, v)
	}

	request, err := http.NewRequest("POST", uri, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, NewInternalServiceError(err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	request = request.WithContext(c)

	resp, err := h.client.Do(request)
	if err != nil {
		return nil, NewIoError(err).WithLogMessage(fmt.Sprintf("couldn't send request"))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, NewIoError(err).WithLogMessage(fmt.Sprintf("cannot read response"))
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, NewIoError(err).WithLogMessage(fmt.Sprintf("got response status code %s", resp.StatusCode))
	}
	return body, nil
}

func (h *httpClientWrapper) do(c context.Context, httpMethod string, payload interface{}, host, handler string,
	target interface{}, headers map[string]string, internal bool) (srvReply ServiceReply) {
	return h.doFull(c, httpMethod, payload, host, handler, target, headers, internal, ContentTypeJSON)
}

func (h *httpClientWrapper) doFull(c context.Context, httpMethod string, payload interface{}, host, handler string,
	target interface{}, headers map[string]string, internal bool, contentType string) (srvReply ServiceReply) {
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
	if httpMethod != http.MethodGet {
		if b != nil {
			req, err = http.NewRequest(httpMethod, url, b)
		} else {
			req, err = http.NewRequest(httpMethod, url, nil)
		}
		if err != nil {
			return NewInternalServiceError(err).WithLogMessage(fmt.Sprintf("Cannot marshal request to %s", url))
		}
	} else {
		if payload != nil {
			if queryString, ok := payload.(*string); !ok {

			} else {
				url = url + "?" + *queryString
			}
		}
		req, err = http.NewRequest(httpMethod, url, nil)
	}

	req = req.WithContext(c)
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	if _, ok := headers["Content-Type"]; !ok {
		if contentType == ContentTypeJSON {
			req.Header.Add("Content-Type", "application/json")
		} else if contentType == ContentTypeXML {
			req.Header.Add("Content-Type", "application/xml")
		}
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
			srvReply = NewServiceError(&resType, fmt.Errorf(string(srvError.GetStatus())), srvError.GetMessageId(), 1)
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
	if err := unmarshalDataToStruct(body, target, contentType, url); err != nil {
		return err
	}
	return
}

type Marshaler interface {
	Unmarshal(data []byte) error
}

func unmarshalDataToStruct(data []byte, target interface{}, contentType string, logStrings ...interface{}) ServiceReply {
	if target == nil || data == nil {
		return nil
	}
	if f, ok := target.(Marshaler); ok {
		if err := f.Unmarshal(data); err != nil {
			return NewInternalServiceError(err)
		}
		return nil
	}
	if contentType == ContentTypeJSON || contentType == "" {
		if err := json.Unmarshal(data, &target); err != nil {
			return NewInternalServiceError(err).WithLogMessage("cannot read response")
		}
	} else if contentType == ContentTypeXML {
		if err := xml.Unmarshal(data, &target); err != nil {
			return NewInternalServiceError(err).WithLogMessage("cannot read response")
		} else {
			fmt.Println(target)
		}
	} else {
		return NewBadRequestError("can't unmarshal response into target struct. unsupported content type: " + contentType)
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
