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

func (r *BuildRepository) GetAllBuilds(
	pageNo int,
	pageSize int,
	sortOrder string,
	sortBy string,
	search string,
	skillId uuid.UUID,
	minRating *int,
	ratingCategory types.RatingCategory) ([]BuildListQuery, error) {

	var builds []BuildListQuery

	// allowed columns and sort directions
	validSortColumns := map[string]bool{
		"created_at":           true,
		"avg_bossing_rating":   true,
		"avg_endgame_rating":   true,
		"avg_fun_rating":       true,
		"avg_creative_rating":  true,
		"avg_speedfarm_rating": true,
		"main_skill_id":        true,
	}

	validSortDirections := map[string]bool{
		"ASC":  true,
		"DESC": true,
	}

	// validate 'sortBy' and 'sortOrder' and set defaults
	if !validSortColumns[sortBy] {
		sortBy = "created_at"
	}
	if !validSortDirections[sortOrder] {
		sortOrder = "ASC"
	}

	query := `
		SELECT
			builds.id as id,
			title,
			builds.description as description,
			ascendancies.name as ascendancy_name,
			classes.name as class_name,
			skills.name as main_skill_name,
			avg_end_game_rating,
			avg_fun_rating,
			avg_creative_rating,
			avg_speed_farm_rating,
			avg_bossing_rating,
			views,
			status,
			builds.created_at as created_at
		FROM builds
		JOIN classes ON classes.id = builds.class_id
	  JOIN skills ON skills.id = builds.main_skill_id
	  LEFT JOIN ascendancies ON ascendancies.id = builds.ascendancy_id
	`

	// arguments for final query execution
	var queryArgs []interface{}

	// by default status is published when searching for community builds
	query += fmt.Sprintf("\nWHERE status = $1")
	queryArgs = append(queryArgs, types.IsPublished)

	// keyword search filter
	if search != "" {
		searchQuery := "%" + search + "%"
		queryArgs = append(queryArgs, searchQuery)
		query += fmt.Sprintf("\nAND title LIKE $1")
	}

	// skill filter
	if skillId != uuid.Nil {
		queryArgs = append(queryArgs, skillId)
		query += fmt.Sprintf("\nAND main_skill_id = $%d", len(queryArgs))
	}

	// TODO: WIP - rating filter
	// if minRating != nil {
	// 	queryArgs = append(queryArgs, minRating)
	// 	query += fmt.Sprintf("\nAND main_skill_id = $%d", len(queryArgs))
	// }

	// construct pagination and sorting
	queryArgs = append(queryArgs, pageSize, (pageNo-1)*pageSize)

	query += fmt.Sprintf(`
		ORDER BY %s %s
		LIMIT $%d
		OFFSET $%d
	`, sortBy, sortOrder,
		len(queryArgs)-1, // second last arg is LIMIT
		len(queryArgs),   // last arg is OFFSET
	)

	fmt.Printf("\n\nFinal Query: %s\n\n", query)
	fmt.Printf("\n\nFinal QueryArgs: %+v\n\n", queryArgs)

	err := r.DB.Select(&builds, query, queryArgs...)

	if err != nil {
		return nil, errorutils.AnalyzeDBErr(err)
	}

	buildList := make([]BuildListQuery, len(builds))

	fmt.Println("Builds after query:", builds)

	for index, build := range builds {
		buildList[index] = BuildListQuery{
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
			Views:              build.Views,
			Status:             build.Status,
			CreatedAt:          build.CreatedAt,
		}
	}

	fmt.Println("all community buildList:", buildList)

	return buildList, nil
}

func (r *BuildRepository) CreateBuild(memberId uuid.UUID, createBuildRequest CreateBuildRequest) (*uuid.UUID, error) {

	fmt.Println("createBuildRequest", createBuildRequest)
	fmt.Println("memberId", memberId)
	baseQuery := `
	INSERT INTO 
		builds(member_id, main_skill_id, class_id, title, description
	`

	if createBuildRequest.AscendancyID != uuid.Nil {
		baseQuery += ", ascendancy_id"

	}

	endQuery := `
	)
	VALUES($1, $2, $3, $4, $5`

	finalQuery := ""

	if createBuildRequest.AscendancyID != uuid.Nil {
		finalQuery = finalQuery + `, $6)
			RETURNING id
			`
	} else {
		finalQuery = finalQuery + `)
		RETURNING id
			`
	}

	query := baseQuery + endQuery + finalQuery

	fmt.Printf("\nFinal query: %+v\n\n", query)

	var buildId uuid.UUID

	var err error
	if createBuildRequest.AscendancyID != uuid.Nil {
		err = r.DB.QueryRowx(query, memberId, createBuildRequest.SkillID, createBuildRequest.ClassID, createBuildRequest.Title, createBuildRequest.Description, createBuildRequest.AscendancyID).Scan(&buildId)
	} else {
		err = r.DB.QueryRowx(query, memberId, createBuildRequest.SkillID, createBuildRequest.ClassID, createBuildRequest.Title, createBuildRequest.Description).Scan(&buildId)
	}

	if err != nil {
		return nil, errorutils.AnalyzeDBErr(err)
	}

	return &buildId, nil
}

func (r *BuildRepository) UpdateBuild(memberId uuid.UUID, buildId uuid.UUID, request UpdateBuildRequest) error {

	fmt.Printf("request.SkillID :%s\n", request.SkillID)

	query := `
	UPDATE builds
	SET 
		title = COALESCE(:title, title),
		description = COALESCE(:description, description),
		main_skill_id = COALESCE(:main_skill_id, main_skill_id),
		class_id = COALESCE(:class_id, class_id),
		ascendancy_id = COALESCE(:ascendancy_id, ascendancy_id),
		updated_at = CURRENT_TIMESTAMP
	WHERE id = :build_id AND member_id = :member_id
	`

	params := map[string]interface{}{
		"title":         request.Title,
		"description":   request.Description,
		"main_skill_id": request.SkillID,
		"class_id":      request.ClassID,
		"ascendancy_id": request.AscendancyID,
		"build_id":      buildId,
		"member_id":     memberId,
	}

	fmt.Printf("Final Query: %s\n", query)

	_, err := r.DB.NamedExec(query, params)

	if err != nil {
		return errorutils.AnalyzeDBErr(err)
	}

	return nil
}

func (r *BuildRepository) CreateBuildTags(buildId uuid.UUID, tagIds []uuid.UUID) error {
	buildTagQuery := `
	INSERT INTO build_tags(build_id, tag_id)
	VALUES($1, unnest($2::uuid[]))
	`
	_, buildTagsErr := r.DB.Exec(buildTagQuery, buildId, pq.Array(tagIds))
	if buildTagsErr != nil {
		return errorutils.AnalyzeDBErr(buildTagsErr)
	}

	return nil
}

/**
* Multi build query By Id, for a specific member. Used for member personal viewing and edit.
**/
func (r *BuildRepository) GetBuildsByMemberId(memberId uuid.UUID) (*[]BuildListQuery, error) {
	var builds []BuildListQuery

	query := `
	SELECT
		builds.id as id,
		title,
		builds.description as description,
		ascendancies.name as ascendancy_name,
		classes.name as class_name,
		skills.name as main_skill_name,
		avg_end_game_rating,
		avg_fun_rating,
		avg_creative_rating,
		avg_speed_farm_rating,
		avg_bossing_rating,
		views,
		status,
		builds.created_at as created_at
 	FROM builds
  JOIN classes ON classes.id = builds.class_id
  JOIN skills ON skills.id = builds.main_skill_id
  LEFT JOIN ascendancies ON ascendancies.id = builds.ascendancy_id

	WHERE member_id = $1
	`

	err := r.DB.Select(&builds, query, memberId)

	if err != nil {
		return nil, errorutils.AnalyzeDBErr(err)
	}

	return &builds, nil
}

/**
* Single build query By Id, for a specific member.
**/
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
* Queries for a corresponding builds tags.
**/
func (r *BuildRepository) GetBuildTagsForMemberById(buildId uuid.UUID) (*[]models.Tag, error) {

	var tags []models.Tag

	query := `
	SELECT
		tags.id as id,
		tags.created_at as created_at,
		tags.name as name
	FROM tags
	JOIN build_tags ON build_tags.tag_id = tags.id
	JOIN builds ON builds.id = build_tags.build_id
	WHERE builds.id = $1
	`

	err := r.DB.Select(&tags, query, buildId)

	if err != nil {
		fmt.Println("Errored when querying tags.")
		return nil, errorutils.AnalyzeDBErr(err)
	}

	return &tags, nil
}

/**
* Community Version for public viewing when the build is published.
*
* Getting all information related with builds via joins of
* join table build skills, builds, and skills, in order to reduce DB
* round trips.
**/
func (r *BuildRepository) GetBuildInfo(buildId uuid.UUID) (*BuildInfoResponse, error) {
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
	WHERE builds.id = $1
	ORDER BY build_skill_links.id
	`

	fmt.Println("buildId", buildId)
	err := r.DB.Select(&buildInfoRows, query, buildId)

	fmt.Println("buildInfoRows", buildInfoRows)

	if err != nil {
		fmt.Printf("Error when querying for build info: %s\n", err)
		return nil, errorutils.AnalyzeDBErr(err)
	}

	var buildItemRows []BuildItemSetResponse
	itemQuery := `
	SELECT
		builds.id AS build_id,
		build_item_sets.id AS set_id,
		build_item_set_items.slot AS set_slot,
		COALESCE(items.unique_item, false) AS unique_item,
		COALESCE(items.id, null) AS item_id,
		COALESCE(items.name, '') AS name,
		COALESCE(items.description, '') AS description,
		COALESCE(items.required_level, null) AS required_level,
		COALESCE(items.required_strength, null) AS required_strength,
		COALESCE(items.required_dexterity, null) AS required_dexterity,
		COALESCE(items.required_intelligence, null) AS required_intelligence,
		COALESCE(items.damage, null) AS damage,
		COALESCE(items.aps, null) AS aps,
		COALESCE(items.crit, null) AS crit,
		COALESCE(items.pdps, null) AS pdps,
		COALESCE(items.edps, null) AS edps,
		COALESCE(items.dps, null) AS dps,
		COALESCE(items.additional, null) AS additional,
		COALESCE(items.stats, null) AS stats,
		COALESCE(items.implicit, null) AS implicit,
		COALESCE(items.slot, '') AS slot,
		COALESCE(items.armour, null) AS armour,
		COALESCE(items.energy_shield, null) AS energy_shield,
		COALESCE(items.evasion, null) AS evasion,
		COALESCE(items.block, null) AS block,
		COALESCE(items.ward, null) AS ward
	FROM builds
	JOIN build_item_sets ON build_item_sets.build_id = builds.id
	JOIN build_item_set_items ON build_item_set_items.build_item_set_id = build_item_sets.id
	LEFT JOIN items ON items.id = build_item_set_items.item_id
	WHERE builds.id = $1
	`
	err = r.DB.Select(&buildItemRows, itemQuery, buildId)

	if err != nil {
		fmt.Printf("Error when querying for build items: %s\n", err)
		return nil, errorutils.AnalyzeDBErr(err)
	}

	fmt.Println("buildItemRows", &buildItemRows)
	if len(buildInfoRows) == 0 {
		fmt.Println("No builds queried.")

		// no builds queried with skills or item joins
		return nil, nil
	}

	// create the base of the response
	result := BuildInfoResponse{
		ID:          buildInfoRows[0].ID,
		Title:       buildInfoRows[0].Title,
		Description: buildInfoRows[0].Description,
	}

	var skillRows []models.SkillRow

	for _, buildInfoRow := range buildInfoRows {
		skillRows = append(skillRows, models.SkillRow{
			SkillLinkID:     buildInfoRow.SkillLinkID,
			SkillLinkName:   buildInfoRow.SkillLinkName,
			SkillLinkIsMain: buildInfoRow.SkillLinkIsMain,
			SkillID:         buildInfoRow.SkillID,
			SkillType:       buildInfoRow.SkillType,
		})

	}

	skills := r.GetAndFormSkillLinks(skillRows)

	result.Skills = &skills

	result.Sets = buildItemRows

	return &result, nil
}

/**
* Retrieves and organizes all skills and skill links.
**/

func (r *BuildRepository) GetAndFormSkillLinks(skillData []models.SkillRow) SkillGroupResponse {
	var mainSkillLink SkillLinkResponse          // store primary skills
	var additionalSkillLinks []SkillLinkResponse // stores additional skills

	// group up all skill information
	for _, row := range skillData {

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
				// starting link skill
				if types.SkillType(row.SkillType) == types.Active {

					existingSkillLink.Skill = models.Skill{
						ID:   row.SkillID,
						Name: row.SkillName,
						Type: row.SkillType,
					}

				} else {
					existingSkillLink.Links = append(existingSkillLink.Links, models.Skill{
						ID:   row.SkillID,
						Name: row.SkillName,
						Type: row.SkillType,
					})
				}

			} else {
				// creating new link

				var newAdditionalSkillLink SkillLinkResponse

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

		}

	}

	// wrap them for response
	skills := SkillGroupResponse{
		MainSkillLinks:   mainSkillLink,
		AdditionalSkills: additionalSkillLinks,
	}

	return skills
}

/**
* Creates a skill link group for a build.
**/
func (r *BuildRepository) CreateBuildSkillLinkTx(tx *sqlx.Tx, buildId uuid.UUID, name string, isMain bool) (uuid.UUID, error) {

	// validate that skill link and name combination doesn't already exists for the build
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

/**
* Updates the average rating of a specific category for a build.
**/
func (r *BuildRepository) UpdateAvgRatingForBuild(buildId string, category types.RatingCategory, avgRating float32) error {

	// mapping rating category to build category column name
	categoryColumn := map[types.RatingCategory]string{
		types.Endgame:   "avg_end_game_rating",
		types.Fun:       "avg_fun_rating",
		types.Creative:  "avg_creative_rating",
		types.Speedfarm: "avg_speedfarm_rating",
		types.Bossing:   "avg_bossing_rating",
	}

	// package parameters
	params := map[string]interface{}{
		"build_id":   buildId,
		"avg_rating": avgRating,
	}

	// construct query with mapped category column name
	query := fmt.Sprintf(`
	UPDATE builds
	SET %s = :avg_rating
	WHERE id = :build_id
	`, categoryColumn[category])

	_, err := r.DB.NamedExec(query, params)

	if err != nil {
		return errorutils.AnalyzeDBErr(err)
	}

	return nil
}

/**
* Get's a class belonging to a build.
**/
func (r *BuildRepository) GetBuildClassById(buildId uuid.UUID) (*string, error) {
	var className string

	query := `
	SELECT
		classes.name as name
	FROM builds
	JOIN classes on classes.id = builds.class_id
	WHERE builds.id = $1
	`

	err := r.DB.Get(&className, query, buildId)

	if err != nil {
		return nil, errorutils.AnalyzeDBErr(err)
	}

	return &className, nil

}

/**
* Get's an ascendancy belonging to a build.
**/
func (r *BuildRepository) GetBuildAscendancyById(buildId uuid.UUID) (*string, error) {
	var ascendancyName string

	query := `
	SELECT
		ascendancies.name as name
	FROM builds
	JOIN ascendancies on ascendancies.id = builds.ascendancy_id
	WHERE builds.id = $1
	`

	err := r.DB.Get(&ascendancyName, query, buildId)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, errorutils.AnalyzeDBErr(err)
	}

	return &ascendancyName, nil
}

/**
* Creates a skill link group for a build.
**/
func (r *BuildRepository) CreateBuildItemSetTx(tx *sqlx.Tx, buildId uuid.UUID) (uuid.UUID, error) {

	var newId uuid.UUID

	query := `
	INSERT INTO build_item_sets(build_id)
	VALUES($1)
	RETURNING id
	`

	err := tx.QueryRowx(query, buildId).Scan(&newId)

	if err != nil {
		fmt.Printf("Error when attempting to insert into build_item_links: %s\n", err)
		return uuid.Nil, errorutils.AnalyzeDBErr(err)
	}

	return newId, nil
}

/**
* Adds a item to a existing item set.
**/
func (r *BuildRepository) CreateItemToSetTx(tx *sqlx.Tx, buildItemSetId uuid.UUID, slot string, itemId interface{}) error {

	// create set item
	query := `
	INSERT INTO build_item_set_items(build_item_set_id, item_id, slot)
	VALUES(:build_item_set_id, :item_id, :slot)
	`

	params := map[string]interface{}{
		"build_item_set_id": buildItemSetId,
		"item_id":           itemId,
		"slot":              slot,
	}

	_, err := tx.NamedExec(query, params)

	if err != nil {
		fmt.Printf("DEBUG AddItemToSetTx: Error when attempting to insert into join table build_item_set_items: %s\n", err)
		return errorutils.AnalyzeDBErr(err)
	}

	return nil
}

/**
* Creates a skill link group for a build.
**/
func (r *BuildRepository) GetBuildItemSetIdTx(tx *sqlx.Tx, buildId uuid.UUID) (uuid.UUID, error) {
	fmt.Println("buildId", buildId)
	var setId uuid.UUID

	query := `
	SELECT id
	FROM build_item_sets
	WHERE build_id = $1
	`

	err := tx.Get(&setId, query, buildId)

	if err != nil {
		fmt.Printf("Error when attempting to insert into build_item_links: %s\n", err)
		return uuid.Nil, errorutils.AnalyzeDBErr(err)
	}

	return setId, nil
}

/**
* Adds a item to a existing item set.
**/
func (r *BuildRepository) UpdateItemToSetTx(tx *sqlx.Tx, buildItemSetId uuid.UUID, slot string, itemId interface{}) error {

	var itemSetSlotItemId uuid.UUID
	// get item of update
	query := `
	SELECT id
	FROM build_item_set_items
	WHERE build_item_set_id = $1 AND slot = $2
	`
	err := tx.Get(&itemSetSlotItemId, query, buildItemSetId, slot)
	if err != nil {
		fmt.Printf("Error when attempting to insert into build_item_set_item: %s\n", err)
		return errorutils.AnalyzeDBErr(err)
	}
	fmt.Println("itemSetSlotItemId", itemSetSlotItemId)
	// update set item
	query = `
	UPDATE build_item_set_items
	SET item_id = :item_id
	WHERE id = :id
	`

	params := map[string]interface{}{
		"id":      itemSetSlotItemId,
		"item_id": itemId,
	}

	_, err = tx.NamedExec(query, params)

	if err != nil {
		fmt.Printf("DEBUG AddItemToSetTx: Error when attempting to insert into join table build_item_set_items: %s\n", err)
		return errorutils.AnalyzeDBErr(err)
	}

	return nil
}

func (r *BuildRepository) GetBuildItemSetById(buildId uuid.UUID) ([]BuildItemSetResponse, error) {
	var buildItemSetItems []BuildItemSetResponse
	query := `
	SELECT
	build_item_sets.id AS set_id,
	build_item_sets.build_id AS build_id,
	build_item_set_items.slot AS set_slot,
	COALESCE(items.unique_item, false) AS unique_item,
	COALESCE(items.id, null) AS item_id,
	COALESCE(items.name, '') AS name,
	COALESCE(items.description, '') AS description,
	COALESCE(items.required_level, null) AS required_level,
	COALESCE(items.required_strength, null) AS required_strength,
	COALESCE(items.required_dexterity, null) AS required_dexterity,
	COALESCE(items.required_intelligence, null) AS required_intelligence,
	COALESCE(items.damage, null) AS damage,
	COALESCE(items.aps, null) AS aps,
	COALESCE(items.crit, null) AS crit,
	COALESCE(items.pdps, null) AS pdps,
	COALESCE(items.edps, null) AS edps,
	COALESCE(items.dps, null) AS dps,
	COALESCE(items.additional, null) AS additional,
	COALESCE(items.stats, null) AS stats,
	COALESCE(items.implicit, null) AS implicit,
	COALESCE(items.slot, '') AS slot,
	COALESCE(items.armour, null) AS armour,
	COALESCE(items.energy_shield, null) AS energy_shield,
	COALESCE(items.evasion, null) AS evasion,
	COALESCE(items.block, null) AS block,
	COALESCE(items.ward, null) AS ward
	FROM build_item_sets
	JOIN build_item_set_items ON build_item_set_items.build_item_set_id = build_item_sets.id
	LEFT JOIN items ON items.id = build_item_set_items.item_id
	WHERE build_id = $1
	`

	err := r.DB.Select(&buildItemSetItems, query, buildId)
	if err != nil {
		return nil, err
	}

	return buildItemSetItems, nil
}

func (r *BuildRepository) DeleteBuildByIdForMember(memberId uuid.UUID, buildId uuid.UUID) error {
	query := `
	DELETE FROM builds 
	WHERE id = $1 AND member_id = $2
	`

	_, err := r.DB.Exec(query, buildId, memberId)
	if err != nil {
		return errorutils.AnalyzeDBErr(err)
	}

	return nil
}

func (r *BuildRepository) GetBasicBuildInfoByIdForMember(id uuid.UUID, memberId uuid.UUID) (*BasicBuildInfoResponse, error) {
	var build BasicBuildInfoResponse

	query := `
	SELECT 
		builds.id as id,
		title,
		builds.description as description,
		skills.name as main_skill,
		classes.name as class,
		ascendancies.name as ascendancy,
		avg_end_game_rating,
		avg_fun_rating,
		avg_creative_rating,
		avg_speed_farm_rating,
		views,
		status,
		builds.created_at as created_at,
		builds.updated_at as updated_at
	FROM builds
	JOIN classes ON classes.id = builds.class_id
	JOIN ascendancies ON ascendancies.id = builds.ascendancy_id
	JOIN skills ON skills.id = builds.main_skill_id
	WHERE builds.member_id = $1
	AND builds.id = $2
	`

	err := r.DB.Get(&build, query, memberId, id)

	if err != nil {
		return nil, errorutils.AnalyzeDBErr(err)
	}

	return &build, nil
}

func (r *BuildRepository) UpdateBuildByIdForMemberService(id uuid.UUID, memberId uuid.UUID, updateFields models.Build) error {

	return nil
}
