package build

import (
	"fmt"

	"github.com/darkphotonKN/community-builds/internal/models"
	"github.com/darkphotonKN/community-builds/internal/skill"
	"github.com/google/uuid"
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

func (s *BuildService) CreateBuildService(memberId uuid.UUID, createBuildRequest CreateBuildRequest) error {
	// confirm skill exists
	_, err := s.SkillService.GetSkillByIdService(createBuildRequest.SkillID)

	if err != nil {
		return fmt.Errorf("main skill id could not be found when attempting to create build for it.")
	}

	// check how many builds a user has, prevent creation if over the limit
	builds, err := s.GetBuildsForMemberService(memberId)

	if len(*builds) > maxBuildCount {
		return fmt.Errorf("Number of builds allowed have reached maximum capacity.")
	}

	if err != nil {
		return err
	}

	// create build with this skill and for this member
	return s.Repo.CreateBuild(memberId, createBuildRequest)
}

/**
* Get list builds available to a member.
**/
func (s *BuildService) GetBuildsForMemberService(memberId uuid.UUID) (*[]models.Build, error) {
	return s.Repo.GetBuildsByMemberId(memberId)
}

/**
* Get a single build for a member by Id.
**/
func (s *BuildService) GetBuildForMemberByIdService(memberId uuid.UUID, buildId uuid.UUID) (*models.Build, error) {
	return s.Repo.GetBuildForMemberById(memberId, buildId)
}

/**
* Get a single build with all information by ID for a member.
**/
func (s *BuildService) GetBuildInfoForMemberService(memberId uuid.UUID, buildId uuid.UUID) (*BuildInfoResponse, error) {
	return s.Repo.GetBuildInfo(memberId, buildId)
}

/**
* Adds primary and secondary skills and links to an existing build via a transanction,
* rolling back on error with any portion of the creation.
**/
func (s *BuildService) AddSkillLinksToBuildService(memberId uuid.UUID, buildId uuid.UUID, request AddSkillsToBuildRequest) error {
	// get build and check if it exists
	_, err := s.GetBuildForMemberByIdService(memberId, buildId)

	if err != nil {
		return err
	}

	// -- CREATE SKILL LINKS FOR BUILD --

	// --- transaction start ---
	tx, err := s.Repo.DB.Beginx() // NOTE: use "beginx" for the *sqlx* version

	if err != nil {
		return fmt.Errorf("failed to initiate transaction: %w", err)
	}

	// --- transaction commit or rollback ---
	defer func() {
		// roll back during crashes or unexpected events
		if p := recover(); p != nil {
			fmt.Println("Rolling back due to a crash or unexpected event.")
			tx.Rollback()

			// re-throw panic after rollback
			panic(p)
		}

		// Rollback in the case of any errors.
		if err != nil {
			fmt.Println("Error on DEFER, rolling back:", err)
			tx.Rollback()
		}
	}()

	// --- primary links ---

	// add main skill group
	mainSkillLinkId, err := s.Repo.CreateBuildSkillLinkTx(tx, buildId, request.MainSkillLinks.SkillLinkName, true)

	// add main skill relation to main skill link
	err = s.Repo.AddSkillToLinkTx(tx, mainSkillLinkId, request.MainSkillLinks.Skill)

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

	// commit but check for errors
	if err := tx.Commit(); err != nil {
		fmt.Printf("Failed to commit transaction: %v\n", err)
		tx.Rollback()
		return err
	}

	return nil
}
