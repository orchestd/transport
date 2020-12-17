package http

import (
	. "bitbucket.org/HeilaSystems/servicereply"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/fx"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)


type TestInterface struct {
}

func NewTestInterface() TestInterface {
	return TestInterface{}
}

type TestReq struct {

}

type TestRes struct {
	Hello string `json:"hello"`
}

func (i TestInterface) Test(c context.Context,req TestReq)(TestRes, ServiceReply){
	return TestRes{Hello: "world"},nil
}

func Test_Main(t *testing.T) {
	testHandler :=  func(router *gin.Engine,m TestInterface) {
		router.GET("/", HandleFunc(m.Test))
	}

	convey.Convey("Given a test handler with empty request ",t, func() {
		app := fx.New(
			fx.Provide(
				NewGinServer,
				NewTestInterface,
			),
			fx.Invoke(testHandler),
		)
		startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := app.Start(startCtx); err != nil {
			log.Fatal(err)
		}
		convey.Convey("Get response from the handler ", func() {
			resp, err := http.Get("http://localhost:8080/")
			convey.Convey("http shouldn't return error", func() {
				convey.So(err,convey.ShouldBeNil)
			})
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			convey.Convey("body read shouldnt return error", func() {
				convey.So(err,convey.ShouldBeNil)
			})
			var res TestRes
			fmt.Println(string(body))
			err = json.Unmarshal(body, &res)
			convey.Convey("Error should be nil", func() {
				convey.So(err,convey.ShouldBeNil)
			})
		})
	})

}