package handler

import (
	"net/http"

	"mahjong-backend/internal/domain/entity"
	"mahjong-backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

// YakuHandler handles yaku evaluation requests
type YakuHandler struct {
	yakuUC *usecase.YakuUsecase
}

func NewYakuHandler(yu *usecase.YakuUsecase) *YakuHandler {
	return &YakuHandler{yakuUC: yu}
}

// HandleEvaluateYaku handles POST /api/v1/yaku
func (h *YakuHandler) HandleEvaluateYaku(c *gin.Context) {
	var req entity.YakuEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	resp := h.yakuUC.EvaluateYaku(&req)
	c.JSON(http.StatusOK, resp)
}



