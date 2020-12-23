package http

import (
	"bitbucket.org/HeilaSystems/helpers"
	"bitbucket.org/HeilaSystems/servicehelpers"
	"bitbucket.org/HeilaSystems/servicereply"
	httpError "bitbucket.org/HeilaSystems/servicereply/http"
	"bitbucket.org/HeilaSystems/servicereply/status"
	"context"
	"fmt"
	"github.com/gin-gonic/contrib/gzip"
	"go.uber.org/fx"
	"time"

	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)


func HandleFunc( mFunction interface{}) func(context *gin.Context) {
	return func(context *gin.Context) {
		newH := createInnerHandlers(reflect.ValueOf(getHandlerRequestStruct(mFunction)))
		if context.Request.Method != "GET" && context.Request.Method != "DELETE" {
			if err := context.ShouldBindJSON(&newH); err != nil {
				internalError := servicereply.NewBadRequestError("invalidJson").WithError(err).WithLogMessage("Cannot parse request to struct")
				GinErrorReply(context, internalError)
				return
			}
		}

		exec := func() (interface{}, servicereply.ServiceReply) {
			c := reflect.ValueOf(context)
			req := reflect.Indirect(reflect.ValueOf(newH))
			if !IsFunc(mFunction) {
				return nil, servicereply.NewInternalServiceError(nil).WithLogMessage("mFunction must be a function")
			}
			reflect.ValueOf(mFunction)
			responseArr := reflect.ValueOf(mFunction).Call([]reflect.Value{c, req})
			if len(responseArr) == 2 {
				if !responseArr[1].IsNil() {
					return responseArr[0].Interface(), responseArr[1].Interface().(servicereply.ServiceReply)
				}
				return responseArr[0].Interface(), nil
			} else {
				err := fmt.Errorf("invalid response")
				return nil, servicereply.NewInternalServiceError(err)
			}
		}

		if response, err := exec(); err != nil {
			GinErrorReply(context, err)
		}else {
			GinSuccessReply(context, response)
		}
		//else if hc.RedirectUrl != "" {
		//	GinRedirectReply(hc)
		//} else if !hc.ReplyOverridden {
		//	GinSuccessReply(hc, response)
		//}

	}
}


func createInnerHandlers(v reflect.Value) interface{} {
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	} //else - an error ??
	n := reflect.New(v.Type())
	return n.Interface()
}

func IsFunc(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Func
}

func GinErrorReply(c *gin.Context, err servicereply.ServiceReply) {

	c.Errors = append(c.Errors, &gin.Error{Err: err.GetError(), Type: gin.ErrorTypePrivate})
	Response := servicereply.Response{}
	Response.Status = status.GetStatus(err.GetErrorType())
	Response.Message = &servicereply.Message{
		Id:  err.GetUserError(),
		Values: err.GetReplyValues(),
	}
	c.JSON(httpError.GetHttpCode(err.GetErrorType()), Response)
}

func GinSuccessReply(c *gin.Context, reply interface{}) {
	serviceReply := servicereply.Response{}
	serviceReply.Status = status.SuccessStatus
	serviceReply.Data = reply
	c.JSON(http.StatusOK, serviceReply)
}

func NewGinRouter() (*gin.Engine, error) {
	router := gin.New()
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	// Very important !
	router.Use(servicehelpers.RequestId())
	// Recovery middleware
	router.Use(gin.Recovery())
	//router.Use(GinLog())
	//isAlive check
	router.GET("/isAlive", helpers.IsAliveGinHandler)
	//router.GET("/version", GetVersion)
	return router,nil
}

const (
	defaultPort = "8080"
	defaultTimeout =  30 * time.Second
)

func NewGinServer(lc fx.Lifecycle,port  *string, readTimeout ,WriteTimeout *time.Duration) *gin.Engine {

	if port == nil {
		p := defaultPort
		port = &p
	}
	if readTimeout == nil {
		t := defaultTimeout
		readTimeout = &t
	}
	if WriteTimeout == nil {
		t := defaultTimeout
		WriteTimeout = &t
	}
	h ,_ := NewGinRouter()

	s := &http.Server{

		Addr:         ":" + *port,//appConf.ListenOnPort,
		Handler:      h,
		ReadTimeout:  *readTimeout,
		WriteTimeout: *WriteTimeout,//getHttpRespTimeoutSeconds(appConf),
	}

	lc.Append(fx.Hook{
		// To mitigate the impact of deadlocks in application startup and
		// shutdown, Fx imposes a time limit on OnStart and OnStop hooks. By
		// default, hooks have a total of 15 seconds to complete. Timeouts are
		// passed via Go's usual context.Context.
		OnStart: func(ctx context.Context) error {
			// In production, we'd want to separate the Listen and Serve phases for
			// better error-handling.
			s.ListenAndServe()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return s.Shutdown(ctx)
		},
	})
	//if len(appConf.Domains) > 0 {
	//	//domains := []string{"students-aid.org", "stud-aid.com"}
	//	return certmagic.HTTPS(appConf.Domains, h.Engine)
	//}

	return h
}
//func GinRedirectReply(c *gin.Context) {
//	c.Redirect(http.StatusFound, c.RedirectUrl)
//}

func getHandlerRequestStruct(f interface{}) interface{} {
	fType := reflect.TypeOf(f)
	argType := fType.In(1)
	return reflect.New(argType).Interface()
}