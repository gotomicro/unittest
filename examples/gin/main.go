package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ego"
	"github.com/gotomicro/ego/core/elog"
	"github.com/gotomicro/ego/server/egin"
)

//  export EGO_DEBUG=true && go run main.go --config=config.toml
func main() {
	if err := ego.New().Serve(func() *egin.Component {
		server := egin.Load("server.http").Build()
		server.GET("/hello", hello)
		return server
	}()).Run(); err != nil {
		elog.Panic("startup", elog.FieldErr(err))
	}
}

type helloRequest struct {
	Name string `query:"name"`
}

func hello(ctx *gin.Context) {
	req := helloRequest{}
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(401, err.Error())
		return
	}

	ctx.JSON(200, "Hello client: "+req.Name)
}
