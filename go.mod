module bitbucket.org/HeilaSystems/transport

go 1.14

require (
	bitbucket.org/HeilaSystems/helpers v1.21.1
	bitbucket.org/HeilaSystems/serviceerror v0.3.0
	bitbucket.org/HeilaSystems/servicehelpers v1.100.2
	github.com/gin-gonic/contrib v0.0.0-20201101042839-6a891bf89f19
	github.com/gin-gonic/gin v1.6.3
	github.com/golang/protobuf v1.4.3
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.0.1
	github.com/micro/micro/v3 v3.0.2 // indirect
	github.com/smartystreets/goconvey v1.6.4
	github.com/soheilhy/cmux v0.1.4 // indirect
	go.uber.org/fx v1.13.1
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9 // indirect
	google.golang.org/genproto v0.0.0-20201019141844-1ed22bb0c154
	google.golang.org/grpc v1.33.1
	google.golang.org/protobuf v1.25.0

)

replace (
		bitbucket.org/HeilaSystems/serviceerror v0.3.0 => ../serviceerror
)

