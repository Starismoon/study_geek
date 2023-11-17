package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	//wrk -t4 -d10s -c20 -s ./scripts/wrk/signup.lua http://localhost:8080/users/signup
	server := InitWebServer()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello，启动成功了！")
	})
	server.Run(":8080")
}
