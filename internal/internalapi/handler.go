package internalapi

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"aimodelcore/internal/config"
	"aimodelcore/internal/dto"
	"aimodelcore/internal/pkg/response"
	"aimodelcore/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc   *service.Service
	token string
}

func NewHandler(svc *service.Service, cfg *config.AuthConfig) *Handler {
	return &Handler{svc: svc, token: cfg.InternalToken}
}

func Register(r *gin.RouterGroup, h *Handler) {
	r.GET("/models", h.ListModels)
	r.POST("/chat", h.Chat)
	r.POST("/images", h.Images)
}

func (h *Handler) ListModels(c *gin.Context) {
	if !h.authorized(c) {
		return
	}
	tenantID, err := strconv.ParseUint(strings.TrimSpace(c.Query("tenantId")), 10, 64)
	if err != nil || tenantID == 0 {
		response.Fail(c, http.StatusBadRequest, "参数不正确")
		return
	}
	var connectionID uint64
	if raw := strings.TrimSpace(c.Query("connectionId")); raw != "" {
		connectionID, err = strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "参数不正确")
			return
		}
	}
	capability := strings.TrimSpace(c.Query("capability"))
	if capability == "" {
		capability = "image"
	}
	list, err := h.svc.ListEnabledModels(tenantID, connectionID, capability)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, list)
}

func (h *Handler) Chat(c *gin.Context) {
	if !h.authorized(c) {
		return
	}
	var in dto.ChatRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数不正确")
		return
	}
	out, err := h.svc.Chat(c.Request.Context(), in)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) Images(c *gin.Context) {
	if !h.authorized(c) {
		return
	}
	var in dto.ImageRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数不正确")
		return
	}
	out, err := h.svc.Image(c.Request.Context(), in)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, out)
}

func (h *Handler) authorized(c *gin.Context) bool {
	got := strings.TrimSpace(c.GetHeader("X-Internal-Token"))
	if h.token == "" || subtle.ConstantTimeCompare([]byte(got), []byte(h.token)) != 1 {
		response.Fail(c, http.StatusUnauthorized, "内部调用未授权")
		return false
	}
	return true
}

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound), errors.Is(err, service.ErrNoModel):
		response.Fail(c, http.StatusNotFound, service.ErrNoModel.Error())
	case errors.Is(err, service.ErrBadRequest), errors.Is(err, service.ErrNoKey):
		response.Fail(c, http.StatusBadRequest, err.Error())
	default:
		response.Fail(c, http.StatusBadGateway, err.Error())
	}
}
