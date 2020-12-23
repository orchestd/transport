package http

import (
	. "bitbucket.org/HeilaSystems/servicereply"
	"bitbucket.org/HeilaSystems/servicereply/status"
	"bitbucket.org/HeilaSystems/transport/client"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type httpClientWrapper struct {
	client *http.Client
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

func NewHttpClientWrapper(client *http.Client) client.HttpClient {
	return &httpClientWrapper{client: client}
}

func (h *httpClientWrapper) do(c context.Context,httpMethod string ,payload interface{},host, handler string,target interface{},headers map[string]string,internal bool) ServiceReply  {
	b , sErr := getPayload(payload)
	if sErr != nil{
		return sErr
	}
	req, err := http.NewRequest(httpMethod,fmt.Sprintf("http://%s/%s", host, handler),  b)
	if err != nil {
		return NewInternalServiceError( err).WithLogMessage("Cannot marshal request")
	}
	req = req.WithContext(c)
	req.Header.Add("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return NewIoError(err).WithLogMessage("cannot emmit post request")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return NewInternalServiceError( err).WithLogMessage("cannot read response")
	}
	if internal {
		var srvError Response
		if err := json.Unmarshal(body, &srvError); err != nil {
			return NewInternalServiceError(err).WithLogMessage("cannot read response")
		}
		if srvError.Status != status.SuccessStatus {
			resType := status.GetTypeByStatus(srvError.GetStatus())
			msg := srvError.GetMessageId()
			return NewServiceError( &resType, nil,msg,1 )
		}
		if srvError.Data != nil {
			if dataJson  ,err := json.Marshal(srvError.Data);err != nil {
				return NewInternalServiceError(err)
			}else {
				body = dataJson
			}
		}
	}
	if err := unmarshalDataToStruct(body , target);err != nil {
		return err
	}
	return nil
}
func unmarshalDataToStruct(data []byte,target interface{}) ServiceReply  {
	if err := json.Unmarshal(data, &target); err != nil {
		return NewInternalServiceError( err).WithLogMessage("cannot read response")
	}
	return nil
}
func getPayload(payload interface{}) (*bytes.Buffer,ServiceReply) {
	if payload != nil {
		request, err := json.Marshal(payload)
		if err != nil {
			return nil, NewInternalServiceError( err).WithLogMessage("Cannot marshal request")
		}
		return bytes.NewBuffer(request) , nil
	}
	return nil,nil
}

