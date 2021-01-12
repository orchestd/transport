package http

import (
	"bitbucket.org/HeilaSystems/dependencybundler/interfaces/configuration"
	. "bitbucket.org/HeilaSystems/servicereply"
	"bitbucket.org/HeilaSystems/servicereply/status"
	"bitbucket.org/HeilaSystems/transport/client"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type httpClientWrapper struct {
	client *http.Client
	conf configuration.Config
}

func (h *httpClientWrapper) Call(c context.Context,payload interface{},host, handler string,target interface{},headers map[string]string) ServiceReply {
	return h.do(c,http.MethodPost,payload,host,handler,target,headers,true)
}

func (h *httpClientWrapper) Post(c context.Context,payload interface{},host, handler string,target interface{},headers map[string]string) ServiceReply {
	return h.do(c,http.MethodPost,payload,host,handler,target,headers,false)
}

func (h *httpClientWrapper) Get(c context.Context,host, handler string,target interface{},headers map[string]string)  ServiceReply  {
	return h.do(c,http.MethodGet,nil,host,handler,target,headers,false)
}

func (h *httpClientWrapper) Put(c context.Context,payload interface{},host, handler string,target interface{},headers map[string]string)  ServiceReply  {
	return h.do(c,http.MethodPut,payload,host,handler,target,headers,false)
}

func (h *httpClientWrapper) Delete(c context.Context,host, handler string,target interface{},headers map[string]string) ServiceReply  {
	return h.do(c,http.MethodDelete,nil,host,handler,target,headers,false)
}

func NewHttpClientWrapper(client *http.Client ,conf  configuration.Config) (client.HttpClient,error) {
	return &httpClientWrapper{client: client, conf: conf},nil
}

func (h *httpClientWrapper) do(c context.Context,httpMethod string ,payload interface{},host, handler string,target interface{},headers map[string]string,internal bool) (srvReply ServiceReply)  {
	url := fmt.Sprintf("http://%s/%s", host, handler)
	if overrideHost := os.Getenv(host+urlKeyword); len(overrideHost) > 0 {
		url = fmt.Sprintf("%s/%s" , overrideHost, handler)
	}
	srvReply = NewNil()
	b , sErr := getPayload(payload , url)
	if sErr != nil{
		return sErr
	}
	req, err := http.NewRequest(httpMethod,url,  b)
	if err != nil {
		return NewInternalServiceError( err).WithLogMessage(fmt.Sprintf("Cannot marshal request to %s" , url))
	}
	req = req.WithContext(c)
	req.Header.Add("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return NewIoError(err).WithLogMessage(fmt.Sprintf("couldn't send %s request to %s" , httpMethod,url))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return NewInternalServiceError( err).WithLogMessage(fmt.Sprintf("cannot read response from %s",url) )
	}
	if internal {
		var srvError Response
		if err := json.Unmarshal(body, &srvError); err != nil {
			return NewInternalServiceError(err).WithLogMessage(fmt.Sprintf("cannot read response from %s",url)).WithLogValues(ValuesMap{"rawResponse" : string(body)})
		}
		if srvError.Status != status.SuccessStatus {
			resType := status.GetTypeByStatus(srvError.GetStatus())
			msgValues := srvError.GetMessageValues()
			srvReply = NewServiceError( &resType, nil,srvError.GetMessageId(),1 )
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
			if dataJson  ,err := json.Marshal(srvError.Data);err != nil {
				return NewInternalServiceError(err).WithLogMessage(fmt.Sprintf("cannot marshal data from %s",url))
			}else {
				body = dataJson
			}
		} else {
			body = nil
		}
	}
	if err := unmarshalDataToStruct(body , target,url);err != nil {
		return err
	}
	return
}

func unmarshalDataToStruct(data []byte,target interface{},logStrings ...interface{}) ServiceReply  {
	if target == nil{
		return nil
	}
	if data == nil {
		return NewInternalServiceError(nil).WithLogMessage(fmt.Sprintf("Cannot unmarshal empty response from %s to target struct",logStrings...))
	}
	if err := json.Unmarshal(data, &target); err != nil {
		return NewInternalServiceError(err).WithLogMessage("cannot read response")
	}
	return nil
}
func getPayload(payload interface{},url string) (*bytes.Buffer,ServiceReply) {
	if payload != nil {
		request, err := json.Marshal(payload)
		if err != nil {
			return nil, NewInternalServiceError(err).WithLogMessage(fmt.Sprintf("cannot read response from %s",url)).WithLogValues(ValuesMap{"rawResponse" : payload})
		}
		return bytes.NewBuffer(request) , nil
	}
	return nil,nil
}

const urlKeyword = "Url"