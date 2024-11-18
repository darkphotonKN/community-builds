package skill

import (
	"fmt"
	"net/http"

	"github.com/darkphotonKN/community-builds/internal/models"
	"github.com/gin-gonic/gin"
)

type SkillHandler struct {
	Service *SkillService
}

func NewSkillHandler(service *SkillService) *SkillHandler {
	return &SkillHandler{
		Service: service,
	}
}

// --- ADMIN HANDLERS ---
func (h *SkillHandler) CreateSkillHandler(c *gin.Context) {
	var createSkillReq CreateSkillRequest

	if err := c.ShouldBindJSON(&createSkillReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when parsing payload as JSON: %s", err)})
		return
	}

	err := h.Service.CreateSkillService(createSkillReq)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when attempting to create skill %s", err.Error())})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"statusCode": http.StatusCreated, "message": "Successfully created skill."})
}

func (h *SkillHandler) GetSkillsHandler(c *gin.Context) {
	skills, err := h.Service.GetSkills()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when attempting to get all skills %s", err.Error())})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"statusCode": http.StatusCreated, "message": "Successfully created skill.", "result": skills})
}