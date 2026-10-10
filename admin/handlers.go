package admin

import (
	"errors"
	"net/http"
	"strconv"

	"aimodelcore/internal/dto"
	"aimodelcore/internal/pkg/authcontext"
	"aimodelcore/internal/pkg/response"
	"aimodelcore/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/providers", h.Providers)
	r.GET("/dashboard", h.Dashboard)
	r.GET("/usage", h.Usage)
	r.GET("/connections", h.ListConnections)
	r.POST("/connections", h.CreateConnection)
	r.PUT("/connections/:id", h.UpdateConnection)
	r.DELETE("/connections/:id", h.DeleteConnection)
	r.POST("/connections/:id/test", h.TestConnection)
	r.POST("/connections/:id/suggest-models", h.SuggestModels)
	r.GET("/models", h.ListModels)
	r.POST("/models", h.CreateModel)
	r.PUT("/models/:id", h.UpdateModel)
	r.DELETE("/models/:id", h.DeleteModel)
}

func (h *Handler) Providers(c *gin.Context) {
	response.OK(c, h.svc.Providers())
}

func (h *Handler) Dashboard(c *gin.Context) {
	data, err := h.svc.Dashboard(authcontext.TenantID(c))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Usage(c *gin.Context) {
	list, err := h.svc.ListUsage(authcontext.TenantID(c))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, list)
}

func (h *Handler) ListConnections(c *gin.Context) {
	list, err := h.svc.ListConnections(authcontext.TenantID(c))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, list)
}

func (h *Handler) CreateConnection(c *gin.Context) {
	var in dto.ConnectionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数不正确")
		return
	}
	view, err := h.svc.CreateConnection(authcontext.TenantID(c), in)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, view)
}

func (h *Handler) UpdateConnection(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var in dto.ConnectionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数不正确")
		return
	}
	view, err := h.svc.UpdateConnection(authcontext.TenantID(c), id, in)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, view)
}

func (h *Handler) DeleteConnection(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteConnection(authcontext.TenantID(c), id); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (h *Handler) TestConnection(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	result, err := h.svc.TestConnection(c.Request.Context(), authcontext.TenantID(c), id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) SuggestModels(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	n, err := h.svc.SuggestModels(authcontext.TenantID(c), id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, gin.H{"added": n})
}

func (h *Handler) ListModels(c *gin.Context) {
	var connectionID uint64
	if raw := c.Query("connectionId"); raw != "" {
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "参数不正确")
			return
		}
		connectionID = n
	}
	list, err := h.svc.ListModels(authcontext.TenantID(c), connectionID)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, list)
}

func (h *Handler) CreateModel(c *gin.Context) {
	var in dto.ModelInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数不正确")
		return
	}
	view, err := h.svc.CreateModel(authcontext.TenantID(c), in)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, view)
}

func (h *Handler) UpdateModel(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var in dto.ModelInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数不正确")
		return
	}
	view, err := h.svc.UpdateModel(authcontext.TenantID(c), id, in)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, view)
}

func (h *Handler) DeleteModel(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteModel(authcontext.TenantID(c), id); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func pathID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, http.StatusBadRequest, "参数不正确")
		return 0, false
	}
	return id, true
}

func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.Fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrBadRequest), errors.Is(err, service.ErrNoModel), errors.Is(err, service.ErrNoKey):
		response.Fail(c, http.StatusBadRequest, err.Error())
	default:
		response.Fail(c, http.StatusBadRequest, err.Error())
	}
}
