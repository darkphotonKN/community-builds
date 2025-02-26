package build

import (
	"errors"
	"fmt"

	"github.com/darkphotonKN/community-builds/internal/models"
	"github.com/darkphotonKN/community-builds/internal/skill"
	"github.com/darkphotonKN/community-builds/internal/types"
	"github.com/darkphotonKN/community-builds/internal/utils/dbutils"
	"github.com/darkphotonKN/community-builds/internal/utils/errorutils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type BuildService struct {
	Repo         *BuildRepository
	SkillService *skill.SkillService
}

func NewBuildService(repo *BuildRepository, skillService *skill.SkillService) *BuildService {
	return &BuildService{
		Repo:         repo,
		SkillService: skillService,
	}
}

const (
	maxBuildCount = 10
)

/**
* Gets list of public community builds.
**/
func (s *BuildService) GetCommunityBuildsService(pageNo int, pageSize int, sortOrder string, sortBy string, search string, skillId uuid.UUID, minRating *int, ratingCategory types.RatingCategory) ([]BuildListResponse, error) {

	builds, err := s.Repo.GetAllBuilds(pageNo, pageSize, sortOrder, sortBy, search, skillId, minRating, ratingCategory)

	if err != nil {
		return nil, err
	}

	// get tags and add to build
	buildList := make([]BuildListResponse, len(builds))

	for index, build := range builds {
		tags, err := s.Repo.GetBuildTagsForMemberById(build.ID)

		// exit prematurely with error if any tags returned an error
		if err != nil {
			return nil, err
		}

		buildList[index] = BuildListResponse{
			ID:                 build.ID,
			Title:              build.Title,
			Description:        build.Description,
			Class:              build.Class,
			Ascendancy:         build.Ascendancy,
			MainSkillName:      build.MainSkillName,
			AvgEndGameRating:   build.AvgEndGameRating,
			AvgFunRating:       build.AvgFunRating,
			AvgCreativeRating:  build.AvgCreativeRating,
			AvgSpeedFarmRating: build.AvgSpeedFarmRating,
			AvgBossingRating:   build.AvgBossingRating,
			Tags:               *tags,
			Views:              build.Views,
			Status:             build.Status,
			CreatedAt:          build.CreatedAt,
		}
	}

	return buildList, nil
}

/**
* Create build for a signed-in member.
**/
func (s *BuildService) CreateBuildService(memberId uuid.UUID, createBuildRequest CreateBuildRequest) error {
	// confirm skill exists
	_, err := s.SkillService.GetSkillByIdService(createBuildRequest.SkillID)

	if err != nil {
		return fmt.Errorf("main skill id could not be found when attempting to create build for it.")
	}

	// check how many builds a user has, prevent creation if over the limit
	builds, err := s.GetBuildsForMemberService(memberId)

	if err != nil {
		return err
	}

	if len(*builds) > maxBuildCount {
		return fmt.Errorf("Number of builds allowed have reached maximum capacity.")
	}

	// create build with this skill and for this member
	buildId, err := s.Repo.CreateBuild(memberId, createBuildRequest)

	if err != nil {
		return err
	}

	// create build tags
	err = s.Repo.CreateBuildTags(*buildId, createBuildRequest.TagIDs)

	if err != nil {
		return err
	}

	// create build default set
	err = s.CreateDefaultItemSetsToBuildService(memberId, *buildId)

	if err != nil {
		return err
	}

	return nil
}

/**
* Create build for a signed-in member.
**/
func (s *BuildService) UpdateBuildService(memberId uuid.UUID, buildId uuid.UUID, request UpdateBuildRequest) error {
	// TODO: confirm skill exists
	// _, err := s.SkillService.GetSkillByIdService(request.SkillID)

	// if err != nil {
	// 	return fmt.Errorf("main skill id could not be found when attempting to update build for it.")
	// }

	// check build exists before updating
	_, err := s.Repo.GetBuildForMemberById(memberId, buildId)

	if err != nil {
		return fmt.Errorf("Can only update an existing build. %s", err)
	}

	err = s.Repo.UpdateBuild(memberId, buildId, request)

	if err != nil {
		return err
	}

	// TODO: update build tags
	// err = s.Repo.CreateBuildTags(*buildId, request.TagIDs)

	if err != nil {
		return err
	}

	return nil
}

/**
* Get list of builds available to a member.
**/
func (s *BuildService) GetBuildsForMemberService(memberId uuid.UUID) (*[]BuildListResponse, error) {
	baseBuilds, err := s.Repo.GetBuildsByMemberId(memberId)

	if err != nil {
		return nil, err
	}

	buildListResponse := make([]BuildListResponse, len(*baseBuilds))

	// query and add each builds tag's
	for index, build := range *baseBuilds {
		tags, err := s.Repo.GetBuildTagsForMemberById(build.ID)

		// stop query pre-maturely if errored on query
		if err != nil {
			return nil, err
		}

		buildListResponse[index] = BuildListResponse{
			ID:                 build.ID,
			Title:              build.Title,
			Description:        build.Description,
			Class:              build.Class,
			Ascendancy:         build.Ascendancy,
			MainSkillName:      build.MainSkillName,
			AvgEndGameRating:   build.AvgEndGameRating,
			AvgBossingRating:   build.AvgBossingRating,
			AvgCreativeRating:  build.AvgCreativeRating,
			AvgFunRating:       build.AvgFunRating,
			AvgSpeedFarmRating: build.AvgSpeedFarmRating,
			Tags:               *tags,
			Views:              build.Views,
			Status:             build.Status,
			CreatedAt:          build.CreatedAt,
		}
	}

	return &buildListResponse, nil
}

/**
* Get a single build for a member by Id, without extra joined information.
**/
func (s *BuildService) GetBuildForMemberByIdService(memberId uuid.UUID, buildId uuid.UUID) (*models.Build, error) {
	return s.Repo.GetBuildForMemberById(memberId, buildId)
}

/**
* Get a single build with all join table information by ID for a member.
**/
func (s *BuildService) GetBuildInfoForMemberService(memberId uuid.UUID, buildId uuid.UUID) (*BuildInfoResponse, error) {
	// base build
	build, err := s.GetBuildForMemberByIdService(memberId, buildId)

	if err != nil {
		return nil, err
	}

	// tags
	tags, err := s.Repo.GetBuildTagsForMemberById(buildId)

	if err != nil {
		return nil, err
	}

	// retrieve build skills

	// get all skill and link info that exist in a build
	skillRows, err := s.SkillService.GetSkillsByBuildIdService(buildId)

	if err != nil {
		return nil, err
	}

	fmt.Printf("\nSkillrows: %+v\n\n", skillRows)

	// group them into the skill group response, if there are returned data
	var skills *SkillGroupResponse

	if len(*skillRows) > 0 {
		formedSkills := s.Repo.GetAndFormSkillLinks(*skillRows)
		skills = &formedSkills
	}

	fmt.Printf("\nFormed Skills: %+v\n\n", skills)

	if err != nil {
		return nil, err
	}

	// get class (name only)
	class, err := s.Repo.GetBuildClassById(buildId)

	if err != nil {
		return nil, err
	}

	// get ascendancy (name only)
	ascendancy, err := s.Repo.GetBuildAscendancyById(buildId)

	if err != nil {
		return nil, err
	}

	// TODO: retrieve build items
	set, err := s.Repo.GetBuildItemSetById(buildId)
	if err != nil {
		fmt.Println("set err", err)
		return nil, err
	}
	fmt.Println("set", set)

	// return all join information (base, class, ascendancy, skills and items)
	buildInfo := BuildInfoResponse{
		ID:          build.ID,
		Title:       build.Title,
		Description: build.Description,
		// TODO: add ascendancy and class
		Class:      *class,
		Ascendancy: ascendancy,
		Skills:     skills,
		Tags:       *tags,
		Sets:       set,
	}

	fmt.Printf("Constructed build info: %+v\n", buildInfo)

	return &buildInfo, nil
}

/**
* Get a single build with all join table information at once.
* Public version.
**/
func (s *BuildService) GetBuildInfoService(buildId uuid.UUID) (*BuildInfoResponse, error) {

	// return all join information (base, class, ascendancy, skills and items)
	return s.Repo.GetBuildInfo(buildId)
}

/**
* Adds primary and secondary skills and links to an existing build via a transanction,
* rolling back on error.
**/
func (s *BuildService) AddSkillLinksToBuildService(memberId uuid.UUID, buildId uuid.UUID, request AddSkillsToBuildRequest) error {

	return dbutils.ExecTx(s.Repo.DB, func(tx *sqlx.Tx) error {
		// get build and check if it exists
		_, err := s.GetBuildForMemberByIdService(memberId, buildId)

		if err != nil {
			return err
		}

		// -- CREATE SKILL LINKS FOR BUILD --

		// --- primary links ---

		// add main skill group
		mainSkillLinkId, err := s.Repo.CreateBuildSkillLinkTx(tx, buildId, request.MainSkillLinks.SkillLinkName, true)

		// add main skill relation to main skill link
		err = s.Repo.AddSkillToLinkTx(tx, mainSkillLinkId, request.MainSkillLinks.Skill)
		// create item mod

		if err != nil {
			return err
		}

		fmt.Printf("mainSkillLink created, id: %s\n", mainSkillLinkId)

		// create skill relations under this main skill link, one skill at a time
		for _, skillId := range request.MainSkillLinks.Links {
			err := s.Repo.AddSkillToLinkTx(tx, mainSkillLinkId, skillId)
			if err != nil {
				return err
			}
		}

		// --- other links --
		for _, skillLinks := range request.AdditionalSkills {

			// add secondary skill group
			secondarySkillLinkId, err := s.Repo.CreateBuildSkillLinkTx(tx, buildId, skillLinks.SkillLinkName, false)

			// add main skill relation to secondary link
			err = s.Repo.AddSkillToLinkTx(tx, secondarySkillLinkId, skillLinks.Skill)
			if err != nil {
				return err
			}

			fmt.Printf("secondarySkillLinkId created, id: %s\n", secondarySkillLinkId)

			// create skill relations under this secondary skill link, one skill at a time
			for _, skillId := range skillLinks.Links {
				err := s.Repo.AddSkillToLinkTx(tx, secondarySkillLinkId, skillId)

				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

/**
* Updates a specific member's build's skills and linksvia a transaction, rolling back on error.
**/
func (s *BuildService) UpdateBuildSkillLinksService(memberId uuid.UUID, buildId uuid.UUID, request UpdateSkillsToBuildRequest) error {
	return dbutils.ExecTx(s.Repo.DB, func(tx *sqlx.Tx) error {

		return fmt.Errorf("")
	})
}

/**
* Updates an average rating of a specific category for a build.
**/
func (s *BuildService) UpdateAvgRatingForBuildService(buildId string, category types.RatingCategory, avgRating float32) error {
	return s.Repo.UpdateAvgRatingForBuild(buildId, category, avgRating)
}

/**
* Create default set,
* rolling back on error.
**/
func (s *BuildService) CreateDefaultItemSetsToBuildService(memberId uuid.UUID, buildId uuid.UUID) error {

	return dbutils.ExecTx(s.Repo.DB, func(tx *sqlx.Tx) error {
		// get build and check if it exists
		_, err := s.GetBuildForMemberByIdService(memberId, buildId)

		if err != nil {
			return err
		}

		itemSetId, err := s.Repo.CreateBuildItemSetTx(tx, buildId)

		if err != nil {
			return err
		}

		itemSetsMap := map[string]interface{}{
			"weapon":     "",
			"shield":     "",
			"helmet":     "",
			"bodyArmour": "",
			"gloves":     "",
			"belt":       "",
			"boots":      "",
			"amulet":     "",
			"leftRing":   "",
			"rightRing":  "",
		}
		// create item relations under this item set, one item at a time
		for key, value := range itemSetsMap {
			if value == "" {
				value = nil
			}
			err = s.Repo.CreateItemToSetTx(tx, itemSetId, key, value)

			if err != nil {
				return err
			}
		}

		return nil
	})
}

/**
* Adds primary and secondary items and links to an existing build via a transanction,
* rolling back on error.
**/
func (s *BuildService) UpdateItemSetsToBuildService(memberId uuid.UUID, buildId uuid.UUID, request AddItemsToBuildRequest) error {

	return dbutils.ExecTx(s.Repo.DB, func(tx *sqlx.Tx) error {
		// get build and check if it exists
		_, err := s.GetBuildForMemberByIdService(memberId, buildId)

		if err != nil {
			return err
		}

		itemSetId, err := s.Repo.GetBuildItemSetIdTx(tx, buildId)

		if err != nil {
			return err
		}
		fmt.Println("itemSetId", itemSetId)
		itemSetsMap := map[string]string{
			"weapon":     request.Weapon,
			"shield":     request.Shield,
			"helmet":     request.Helmet,
			"bodyArmour": request.BodyArmour,
			"gloves":     request.Gloves,
			"belt":       request.Belt,
			"boots":      request.Boots,
			"amulet":     request.Amulet,
			"leftRing":   request.LeftRing,
			"rightRing":  request.RightRing,
		}
		// create item relations under this item set, one item at a time
		for key, value := range itemSetsMap {
			// 表示空值
			var itemId interface{}
			if value == "" {
				itemId = nil
			} else {
				itemId = value
			}
			// not empty
			fmt.Println("key", key)
			fmt.Println("value", value)
			s.Repo.UpdateItemToSetTx(tx, itemSetId, key, itemId)
		}

		return nil
	})
}

/**
* Deletes a build for a member by its build id.
**/
func (s *BuildService) DeleteBuildByMemberService(memberId uuid.UUID, buildId uuid.UUID) error {
	// check if build is member's
	_, err := s.GetBuildForMemberByIdService(memberId, buildId)

	if err != nil {
		return fmt.Errorf("The build does not belong to this member or does not exist.")
	}

	// delete build from db
	err = s.Repo.DeleteBuildByIdForMember(memberId, buildId)

	if err != nil {
		return err
	}

	return nil
}

/**
* Publish a build for a member by Id service.
**/
func (s *BuildService) PublishBuildService(id uuid.UUID, memberId uuid.UUID) error {

	// check if build belongs to member
	build, err := s.Repo.GetBasicBuildInfoByIdForMember(id, memberId)

	fmt.Printf("Build retrieved for member: %+v", build)

	if err != nil {
		return errorutils.AnalyzeDBErr(err)
	}

	if build == nil {
		return errors.New(fmt.Sprintf("No build with the id %s exists.\n", id))
	}

	publishBuild := models.Build{
		Status: int(draft),
	}

	return s.Repo.UpdateBuildByIdForMemberService(id, memberId, publishBuild)
}
