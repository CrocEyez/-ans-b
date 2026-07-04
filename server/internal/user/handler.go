package user

import (
	"errors"
	"net/http"
	"strconv"

	"ans-b/server/internal/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/register", h.Register)
}

func (h *Handler) RegisterProtectedRoutes(group *gin.RouterGroup) {
	group.GET("/me", h.Profile)
}

func (h *Handler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.GET("", h.List)
}

func (h *Handler) Register(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "message": "invalid request body", "data": nil})
		return
	}

	created, err := h.service.Register(c.Request.Context(), RegisterInput{
		Username: request.Username,
		Password: request.Password,
		Nickname: request.Nickname,
	})
	if err != nil {
		status := http.StatusBadRequest
		code := 40000
		if errors.Is(err, ErrUsernameTaken) {
			status = http.StatusConflict
			code = 40900
		}
		c.JSON(status, gin.H{"code": code, "message": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": created})
}

func (h *Handler) Profile(c *gin.Context) {
	claims, ok := auth.CurrentUser(c)
	if !ok || claims.Role != auth.RoleStudent {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40001, "message": "missing current user", "data": nil})
		return
	}
	profile, err := h.service.Profile(c.Request.Context(), claims.UserID)
	if err != nil {
		status := http.StatusInternalServerError
		code := 50000
		if err.Error() == "user not found" {
			status = http.StatusNotFound
			code = 40400
		}
		c.JSON(status, gin.H{"code": code, "message": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": profile})
}

func (h *Handler) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	result, err := h.service.List(c.Request.Context(), ListInput{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}
