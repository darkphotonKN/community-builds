package build

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BuildHandler struct {
	Service *BuildService
}

func NewBuildHandler(service *BuildService) *BuildHandler {
	return &BuildHandler{
		Service: service,
	}
}

/**
* Get all builds for community viewing.
**/

func (h *BuildHandler) GetCommunityBuildsHandler(c *gin.Context) {
	// defaults
	pageNo := 1
	pageSize := 20

	// parse query pagination querystrings to ints
	if pageNoQuery := c.Query("page_no"); pageNoQuery != "" {
		pageNo, _ = strconv.Atoi(pageNoQuery)
	}

	if pageSizeQuery := c.Query("page_size"); pageSizeQuery != "" {
		pageSize, _ = strconv.Atoi(pageSizeQuery)
	}

	// query strings
	sortBy := c.Query("sort_by")
	sortOrder := c.Query("sort_order")
	search := c.Query("search")

	skillQuery := c.Query("skill")

	// validate querystrings
	skillId, err := uuid.Parse(skillQuery)
	if skillQuery != "" && err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Skill in querystring was not a valid uuid, error: %s", err.Error())})
		return
	}

	builds, err := h.Service.GetCommunityBuildsService(pageNo, pageSize, sortOrder, sortBy, search, skillId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when attempting to get all community builds: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statusCode": http.StatusOK, "message": "Successfully retrieved all builds for member.", "result": builds})

}

/**
* Create build for a signed-in member.
**/
func (h *BuildHandler) CreateBuildHandler(c *gin.Context) {
	memberId, _ := c.Get("userId")
	var createBuildReq CreateBuildRequest

	if err := c.ShouldBindJSON(&createBuildReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when parsing payload as JSON: %s", err)})
		return
	}

	err := h.Service.CreateBuildService(memberId.(uuid.UUID), createBuildReq)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when attempting to create a build: %s", err.Error())})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"statusCode": http.StatusCreated, "message": "Successfully created the build."})
}

/**
* Get list of builds by a signed-in member's ID.
**/
func (h *BuildHandler) GetBuildsForMemberHandler(c *gin.Context) {
	memberId, _ := c.Get("userId")

	builds, err := h.Service.GetBuildsForMemberService(memberId.(uuid.UUID))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when attempting to get all builds for memberId %s: %s", memberId, err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statusCode": http.StatusOK, "message": "Successfully retrieved all builds for member.", "result": builds})
}

/**
* Get all information for a build by ID for a particular member.
**/
func (h *BuildHandler) GetBuildInfoByIdHandler(c *gin.Context) {
	memberId, _ := c.Get("userId")

	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error with id %d, not a valid uuid.", id)})
		return
	}

	fmt.Printf("memberId: %s, id: %s\n", memberId, id)

	build, err := h.Service.GetBuildInfoByIdService(memberId.(uuid.UUID), id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when attempting to get all build information for memberId %s: %s", memberId, err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statusCode": http.StatusOK, "message": "Successfully retrieved build for member.", "result": build})
}

/**
* Adds primary, secondary, and other skills and links to an existing build.
**/
func (h *BuildHandler) AddSkillLinksToBuildHandler(c *gin.Context) {
	memberId, _ := c.Get("userId")

	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error with id %d, not a valid uuid.", id)})
		return
	}

	var request AddSkillsToBuildRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("Failed to bind JSON payload: %+v, Error: %s", request, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when parsing payload as JSON: %s", err)})
		return
	}

	err = h.Service.AddSkillLinksToBuildService(memberId.(uuid.UUID), id, request)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when attempting add skills to builds, buildId %s: memberId: %s, error: %s", id, memberId, err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statusCode": http.StatusOK, "message": "Successfully added skills to build for member."})
}

/**
* Updates a specific build's skill links.
**/
func (h *BuildHandler) UpdateBuildSkillLinksHandler(c *gin.Context) {
	memberId, _ := c.Get("userId")

	idParam := c.Param("id")

	id, err := uuid.Parse(idParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error with id %d, not a valid uuid.", id)})
		return
	}

	var request UpdateSkillsToBuildRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("Failed to bind JSON payload: %+v, Error: %s", request, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when parsing payload as JSON: %s", err)})
		return
	}

	err = h.Service.UpdateBuildSkillLinksService(memberId.(uuid.UUID), id, request)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when attempting to get all builds for memberId %s: %s", memberId, err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statusCode": http.StatusOK, "message": "Successfully retrieved all builds for member."})
}

/**
* Quick example setup for quick creation of extra handlers.
**/
func (h *BuildHandler) GetBuildsTemplate(c *gin.Context) {
	memberId, _ := c.Get("userId")

	builds, err := h.Service.GetBuildsForMemberService(memberId.(uuid.UUID))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"statusCode": http.StatusBadRequest, "message": fmt.Sprintf("Error when attempting to get all builds for memberId %s: %s", memberId, err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statusCode": http.StatusOK, "message": "Successfully retrieved all builds for member.", "result": builds})
}
