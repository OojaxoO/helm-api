package util 

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

type respBody struct {
	Code  int         `json:"code"` // 0 or 1, 0 is ok, 1 is error
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func RespErr(c *gin.Context, err error) {
	glog.Warningln(err)

	c.JSON(http.StatusOK, &respBody{
		Code:  1,
		Error: err.Error(),
	})
}

func RespOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, &respBody{
		Code: 0,
		Data: data,
	})
}
