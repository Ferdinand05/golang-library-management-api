package member

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {

	members := api.Group("/members")

	members.GET("", handler.GetMembers)
	members.GET("/:id", handler.GetMember)
	members.POST("", handler.CreateMember)
	members.PUT("/:id", handler.UpdateMember)
	members.DELETE("/:id", handler.DeleteMember)
}
