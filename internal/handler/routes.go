package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pardnchiu/PodRun/internal/database"
	"github.com/pardnchiu/PodRun/internal/utils"
)

var (
	DB *database.SQLite
)

func NewRoutes(db *database.SQLite) error {
	if DB == nil {
		DB = db
	}

	r := gin.Default()

	ip, err := utils.GetLocalIP()
	if err != nil {
		return err
	}

	r.SetTrustedProxies([]string{
		"127.0.0.1",
		ip,
	})

	// * Pod > GET
	r.GET("/api/pod/list", getAPIPodList)

	// * Pod > POST
	r.POST("/api/pod/upsert", postAPIPodUpsert)
	r.POST("/api/pod/update/:uid", postAPIPodRecordUpdate)
	r.POST("/api/pod/record/insert", postAPIPodRecordInsert)

	// # NOT THIS PROJECT POINT, REMOVE IT FOR NOW
	// // * User > POST
	// r.POST("/api/user/upsert", PostAPIUserUpsert)

	// * Other
	r.GET("/api/health", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})
	r.NoRoute(func(c *gin.Context) {
		select {}
	})

	log.Println("start on :8080")
	if err := r.Run(":8080"); err != nil {
		return err
	}

	return nil
}
