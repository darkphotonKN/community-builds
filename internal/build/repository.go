package build

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/darkphotonKN/community-builds/internal/models"
	"github.com/darkphotonKN/community-builds/internal/types"
	"github.com/darkphotonKN/community-builds/internal/utils/errorutils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type BuildRepository struct {
	DB *sqlx.DB
}

func NewBuildRepository(db *sqlx.DB) *BuildRepository {
	return &BuildRepository{
		DB: db,
	}
}

func (r *BuildRepository) CreateBuild(memberId uuid.UUID, createBuildRequest CreateBuildRequest) error {
	query := `
	INSERT INTO builds(member_id, main_skill_id, title, description)
	VALUES($1, $2, $3, $4)
	RETURNING id
	`
	var buildId uuid.UUID

	err := r.DB.QueryRowx(query, memberId, createBuildRequest.SkillID, createBuildRequest.Title, createBuildRequest.Description).Scan(&buildId)

	if err != nil {
		return errorutils.AnalyzeDBErr(err)
	}

	buildTagQuery := `
	INSERT INTO build_tags(build_id, tag_id)
	VALUES($1, unnest($2::uuid[]))
	`

	_, buildTagsErr := r.DB.Exec(buildTagQuery, buildId, pq.Array(createBuildRequest.TagIDs))
	if buildTagsErr != nil {
		return errorutils.AnalyzeDBErr(buildTagsErr)
	}

	return nil
}

func (r *BuildRepository) GetBuildsByMemberId(memberId uuid.UUID) (*[]models.Build, error) {
	var builds []models.Build

	query := `
	SELECT * FROM builds
	WHERE member_id = $1
	`

	err := r.DB.Select(&builds, query, memberId)

	if err != nil {
		return nil, errorutils.AnalyzeDBErr(err)
	}

	return &builds, nil
}

func (r *BuildRepository) GetBuildForMemberById(memberId uuid.UUID, buildId uuid.UUID) (*models.Build, error) {
	var build models.Build

	query := `
	SELECT * FROM builds
	WHERE member_id = $1
	AND id = $2
	`

	err := r.DB.Get(&build, query, memberId, buildId)

	if err != nil {
		return nil, errorutils.AnalyzeDBErr(err)
	}

	return &build, nil
}

/**
* Getting all information related with builds via joins of
* join table build skills, builds, and skills.
**/
func (r *BuildRepository) GetBuildInfo(memberId uuid.UUID, buildId uuid.UUID) (*BuildInfoResponse, error) {
	var buildInfoRows []BuildInfoRows

	query := `
	SELECT 
		builds.id as id,
		builds.title as title,
		builds.description as description,
		build_skill_links.id as skill_link_id,
		build_skill_links.name as skill_link_name,
		build_skill_links.is_main as skill_link_is_main,
		skills.id as skill_id,
		skills.name as skill_name,
		skills.type as skill_type
	FROM builds
	JOIN build_skill_links ON build_skill_links.build_id = builds.id
	JOIN build_skill_link_skills ON build_skill_link_skills.build_skill_link_id = build_skill_links.id
	JOIN skills ON skills.id = build_skill_link_skills.skill_id
	WHERE builds.id = $1 AND builds.member_id = $2
	ORDER BY build_skill_links.id
	`

	err := r.DB.Select(&buildInfoRows, query, buildId, memberId)

	if err != nil {
		fmt.Printf("Error when querying for build info: %s\n", err)
		return nil, errorutils.AnalyzeDBErr(err)
	}

	if len(buildInfoRows) == 0 {
		fmt.Printf("No builds queried.")

		// no builds queried with skills or item joins
		return nil, fmt.Errorf("The build in relation with skills or items returned no data.")
	}

	// create the base of the response
	result := BuildInfoResponse{
		ID:          buildInfoRows[0].ID,
		Title:       buildInfoRows[0].Title,
		Description: buildInfoRows[0].Description,
	}

	var mainSkillLink SkillLinkResponse          // store primary skills
	var additionalSkillLinks []SkillLinkResponse // stores additional skills

	// group up all skill information
	for _, row := range buildInfoRows {

		// --- grouping primary skills ---

		// identify the "main skilllink" with the "skill_link_is_main" field
		if row.SkillLinkIsMain {
			mainSkillLink.SkillLinkName = row.SkillLinkName

			// match start of skill by active skill - match after casting
			if types.SkillType(row.SkillType) == types.Active {

				mainSkillLink.Skill = models.Skill{
					ID:   row.SkillID,
					Name: row.SkillName,
					Type: row.SkillType,
				}

			} else {
				// else its a support skill link
				mainSkillLink.Links = append(mainSkillLink.Links, models.Skill{
					ID:   row.SkillID,
					Name: row.SkillName,
					Type: row.SkillType,
				})
			}
		} else {
			// else we construct the secondary skills

			// --- grouping secondary skills ---

			// find existing skillLink via SkillLinkName
			skillLinkExists := false
			var existingSkillLink *SkillLinkResponse

			for index := range additionalSkillLinks {
				if additionalSkillLinks[index].SkillLinkName == row.SkillLinkName {
					skillLinkExists = true
					// save reference to original skill link slice
					existingSkillLink = &additionalSkillLinks[index]
					break
				}
			}

			// update existing link
			if skillLinkExists {
				fmt.Printf("\n UPDATE -----> Updating Skill Link with info: \n%v\n\n", map[string]interface{}{"skillLinkName": row.SkillLinkName, "name": row.SkillName})
				// starting link skill
				if types.SkillType(row.SkillType) == types.Active {

					existingSkillLink.Skill = models.Skill{
						ID:   row.SkillID,
						Name: row.SkillName,
						Type: row.SkillType,
					}

				} else {
					fmt.Printf("\n\nExisting Links update BEFORE: %+v\n\n", existingSkillLink.Links)
					// support link skill
					existingSkillLink.Links = append(existingSkillLink.Links, models.Skill{
						ID:   row.SkillID,
						Name: row.SkillName,
						Type: row.SkillType,
					})
				}

				fmt.Printf("\n\nExisting Links update AFTER: %+v\n\n", existingSkillLink.Links)
			} else {
				// creating new link

				var newAdditionalSkillLink SkillLinkResponse

				fmt.Printf("\n CREATE -----> Creating New Skill Link with info: \n%v\n\n", map[string]interface{}{"skillLinkName": row.SkillLinkName, "name": row.SkillName})

				// create new skill link name and the first skill
				newAdditionalSkillLink.SkillLinkName = row.SkillLinkName

				// starting link skill
				if types.SkillType(row.SkillType) == types.Active {
					newAdditionalSkillLink.Skill = models.Skill{
						ID:   row.SkillID,
						Name: row.SkillName,
						Type: row.SkillType,
					}

				} else {
					// supporting skill link
					newAdditionalSkillLink.Links = append(newAdditionalSkillLink.Links, models.Skill{
						ID:   row.SkillID,
						Name: row.SkillName,
						Type: row.SkillType,
					})
				}

				// add it to slice of additionalSkillLinks
				additionalSkillLinks = append(additionalSkillLinks, newAdditionalSkillLink)
			}

			fmt.Printf("\n END FN -----> Final additionalSkillLinks: %+v\n\n", additionalSkillLinks)
		}

	}

	// wrap them for response
	skills := SkillGroupResponse{
		MainSkillLinks:   mainSkillLink,
		AdditionalSkills: additionalSkillLinks,
	}

	result.Skills = skills

	return &result, nil
}

/**
* Creates a skill link group for a build.
**/
func (r *BuildRepository) CreateBuildSkillLinkTx(tx *sqlx.Tx, buildId uuid.UUID, name string, isMain bool) (uuid.UUID, error) {

	// validate that skill doesn't already exists first
	var existsId uuid.UUID

	query := `
	SELECT id FROM build_skill_links
	WHERE build_id = $1 AND name = $2
	`

	err := tx.Get(&existsId, query, buildId, name)

	if !errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, errorutils.ErrDuplicateResource
	}

	var newId uuid.UUID

	query = `
	INSERT INTO build_skill_links(build_id, name, is_main)
	VALUES($1, $2, $3)
	RETURNING id
	`

	err = tx.QueryRowx(query, buildId, name, isMain).Scan(&newId)

	if err != nil {
		fmt.Printf("Error when attempting to insert into build_skill_links: %s\n", err)
		return uuid.Nil, errorutils.AnalyzeDBErr(err)
	}

	return newId, nil
}

/**
* Adds a skill to a existing skill link.
**/
func (r *BuildRepository) AddSkillToLinkTx(tx *sqlx.Tx, buildSkillLinkId uuid.UUID, skillId uuid.UUID) error {

	// validate that skill doesn't already exists first
	var existsId uuid.UUID

	query := `
	SELECT id FROM build_skill_link_skills 
	WHERE build_skill_link_id = $1 AND skill_id = $2
	`

	err := tx.Get(&existsId, query, buildSkillLinkId, skillId)

	// if resource IS found, don't create duplicate skill-link to skill relation insert
	if !errors.Is(err, sql.ErrNoRows) {
		fmt.Println("Rows found, duplicate.")
		return errorutils.ErrDuplicateResource
	}

	query = `
	INSERT INTO build_skill_link_skills(build_skill_link_id, skill_id)
	VALUES(:build_skill_link_id, :skill_id)
	`

	params := map[string]interface{}{
		"build_skill_link_id": buildSkillLinkId,
		"skill_id":            skillId,
	}

	_, err = tx.NamedExec(query, params)

	if err != nil {
		fmt.Printf("DEBUG AddSkillToLinkTx: Error when attempting to insert into join table build_skill_link_skills: %s\n", err)
		return errorutils.AnalyzeDBErr(err)
	}

	return nil
}
