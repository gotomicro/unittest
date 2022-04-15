package api

import (
	"github.com/gin-gonic/gin"
)

type HelloRequest struct {
	Name string `form:"name"`
}

func Hello(ctx *gin.Context) {
	req := HelloRequest{}
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(401, err.Error())
		return
	}

	ctx.JSON(200, "Hello client: "+req.Name)
}

type PostHelloRequest struct {
	Name string `json:"name"`
}

func PostHello(ctx *gin.Context) {
	req := HelloRequest{}
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(401, err.Error())
		return
	}

	ctx.JSON(200, "Hello client: "+req.Name)
}
