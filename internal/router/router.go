package router

import (
	"aimodelcore/admin"
	adminmw "aimodelcore/admin/middleware"
	"aimodelcore/internal/config"
	"aimodelcore/internal/internalapi"
	jwtmgr "aimodelcore/internal/pkg/jwt"
	"aimodelcore/internal/repo"
	"aimodelcore/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), corsMiddleware(cfg))

	svc := service.New(repo.New(db), cfg.Auth.SecretKey)
	h := admin.NewHandler(svc)
	internalH := internalapi.NewHandler(svc, &cfg.Auth)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "aimodelcore"})
	})

	v1 := r.Group("/api/v1")
	adminGroup := v1.Group("/admin")
	adminGroup.Use(adminmw.AdminAuth(&cfg.Auth, jwtmgr.NewManager(cfg.Auth.JWTSecret)))
	admin.RegisterRoutes(adminGroup, h)
	internalapi.Register(v1.Group("/internal"), internalH)
	return r
}

func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	origins := cfg.CORS.AllowOrigins
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := origin == ""
		for _, o := range origins {
			if o == origin || o == "*" {
				allowed = true
				break
			}
		}
		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Internal-Token")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
