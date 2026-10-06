package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pardnchiu/PodRun/internal/model"
)

func getAPIPodList(ctx *gin.Context) {
	containers, err := DB.ListPods(ctx.Request.Context())
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": containers})
}

func postAPIPodUpsert(ctx *gin.Context) {
	var pod model.Pod
	if err := ctx.ShouldBindJSON(&pod); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	if err := DB.UpsertPod(ctx.Request.Context(), &pod); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.String(http.StatusOK, "ok")
}

func postAPIPodRecordUpdate(ctx *gin.Context) {
	var pod model.Pod
	if err := ctx.ShouldBindJSON(&pod); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	pod.UID = ctx.Param("uid")

	if pod.UID == "" {
		ctx.String(http.StatusBadRequest, "uid is required")
		return
	}

	if err := DB.UpdatePod(ctx.Request.Context(), &pod); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.String(http.StatusOK, "ok")
}

func postAPIPodRecordInsert(ctx *gin.Context) {
	var record model.Record
	if err := ctx.ShouldBindJSON(&record); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	if err := DB.InsertRecord(ctx.Request.Context(), &record); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.String(http.StatusOK, "ok")
}
