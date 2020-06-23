package main

import "github.com/gin-gonic/gin"

import "helm-api/views/charts"
import "helm-api/setting"
import "helm-api/models"

func init() {
	setting.Setup()
	models.Setup()
}

func main() {
	r := gin.Default()
	r.POST("/charts/:cluster/:namespace/:name", charts.Update)
	r.GET("/charts/:cluster/:namespace/", charts.List)
	r.DELETE("/charts/:cluster/:namespace/:name", charts.Delete)
	r.GET("/charts/:cluster/:namespace/:name", charts.Retrieve)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	port := setting.HttpSetting.Port
	r.Run(":" + port)
}