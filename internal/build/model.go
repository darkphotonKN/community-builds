package build

import (
	"time"

	"github.com/darkphotonKN/community-builds/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// --- Request ---

type CreateBuildRequest struct {
	SkillID      uuid.UUID   `json:"skillId" binding:"required" db:"main_skill_id"`
	TagIDs       []uuid.UUID `json:"tagIds" binding:"required" db:"tag_ids"`
	Title        string      `json:"title" binding:"required,min=6" db:"title"`
	Description  string      `json:"description" binding:"required,min=1" db:"description"`
	ClassID      uuid.UUID   `json:"classId" binding:"required" db:"class_id"`
	AscendancyID uuid.UUID   `json:"ascendancyId" db:"ascendancy_id"`
}

type UpdateBuildRequest struct {
	SkillID      *uuid.UUID  `json:"skillId" binding:"omitempty" db:"main_skill_id"`
	TagIDs       []uuid.UUID `json:"tagIds" binding:"omitempty" db:"tag_ids"`
	Title        *string     `json:"title" binding:"omitempty,min=6" db:"title"`
	Description  *string     `json:"description"  binding:"omitempty,min=10" db:"description"`
	ClassID      *uuid.UUID  `json:"classId" binding:"omitempty" db:"class_id"`
	AscendancyID *uuid.UUID  `json:"ascendancyId" binding:"omitempty" db:"ascendancy_id"`
}

type SkillLinks struct {
	SkillLinkName string      `json:"skillLinkName" binding:"required"`
	Skill         uuid.UUID   `json:"skill" binding:"required,uuid"`
	Links         []uuid.UUID `json:"links" binding:"required,max=6,dive,uuid"`
}

type AddSkillsToBuildRequest struct {
	MainSkillLinks   SkillLinks   `json:"mainSkillLinks" binding:"required"`
	AdditionalSkills []SkillLinks `json:"additionalSkills" binding:"required"`
}

type UpdateSkillsToBuildRequest struct {
	MainSkillLinks   SkillLinks   `json:"mainSkillLinks" binding:"required"`
	AdditionalSkills []SkillLinks `json:"additionalSkills" binding:"required"`
}

type AddItemsToBuildRequest struct {
	Weapon     string `json:"weapon"`
	Shield     string `json:"shield"`
	Helmet     string `json:"helmet"`
	BodyArmour string `json:"bodyArmour"`
	Boots      string `json:"boots"`
	Gloves     string `json:"gloves"`
	Belt       string `json:"Belt"`
	Amulet     string `json:"amulet"`
	LeftRing   string `json:"leftRing"`
	RightRing  string `json:"rightRing"`
}

// --- Response ---

// TEMP CONTAINER for builds JOIN build_skill_links JOIN skills data
type BuildInfoRows struct {
	ID              uuid.UUID `db:"id"`
	Title           string    `db:"title"`
	Description     string    `db:"description"`
	SkillLinkID     string    `db:"skill_link_id"`
	SkillLinkName   string    `db:"skill_link_name"`
	SkillLinkIsMain bool      `db:"skill_link_is_main"`
	SkillID         uuid.UUID `db:"skill_id"`
	SkillName       string    `db:"skill_name"`
	SkillType       string    `db:"skill_type"`
}
type BuildItemRows struct {
	BuildItemSetResponse
}

type TempSkillLink struct {
	Name   string `json:"name"`
	IsMain bool   `json:"isMain"`
}

// Data Structure of Skill Links + Information for Response
type SkillLinkResponse struct {
	SkillLinkName string         `json:"skillLinkName"`
	Skill         models.Skill   `json:"skill"`
	Links         []models.Skill `json:"links"`
}

type SkillGroupResponse struct {
	MainSkillLinks   SkillLinkResponse   `json:"mainSkillLinks"`
	AdditionalSkills []SkillLinkResponse `json:"additionalSkills"`
}

// All Build Information
type BuildInfoResponse struct {
	ID          uuid.UUID              `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Class       string                 `json:"class"`
	Ascendancy  *string                `json:"ascendancy,omitempty"`
	Tags        []models.Tag           `json:"tags"`
	Skills      *SkillGroupResponse    `json:"skills"`
	Sets        []BuildItemSetResponse `json:"sets"`
}

// Partial, basic build information
type BasicBuildInfoResponse struct {
	ID                     uuid.UUID `json:"id" db:"id"`
	Title                  string    `json:"title" db:"title"`
	Description            string    `json:"description" db:"description"`
	Class                  string    `json:"class" db:"class"`
	MainSkill              string    `json:"mainSkill" db:"main_skill"`
	Views                  string    `json:"views" db:"views"`
	AverageEndGameRating   string    `json:"averageEndGameRating" db:"avg_end_game_rating"`
	AverageFunRating       string    `json:"averageFunRating" db:"avg_fun_rating"`
	AverageCreativeRating  string    `json:"averageCreativeRating" db:"avg_creative_rating"`
	AverageSpeedFarmRating string    `json:"averageSpeedFarmRating" db:"avg_speed_farm_rating"`
	Status                 string    `json:"status" db:"status"`
	CreatedAt              time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt              time.Time `json:"updatedAt" db:"updated_at"`
	Ascendancy             *string   `json:"ascendancy,omitempty" db:"ascendancy"`
}

// Build List
type BuildListQuery struct {
	ID                 uuid.UUID `json:"id"`
	Title              string    `db:"title" json:"title"`
	Description        string    `db:"description" json:"description"`
	Class              string    `db:"class_name" json:"class"`
	Ascendancy         *string   `db:"ascendancy_name" json:"ascendancy"`
	MainSkillName      string    `db:"main_skill_name" json:"mainSkill"`
	AvgEndGameRating   *float32  `db:"avg_end_game_rating" json:"avgEndGameRating,omitempty"`
	AvgFunRating       *float32  `db:"avg_fun_rating" json:"avgFunRating,omitempty"`
	AvgCreativeRating  *float32  `db:"avg_creative_rating" json:"avgCreativeRating,omitempty"`
	AvgSpeedFarmRating *float32  `db:"avg_speed_farm_rating" json:"avgSpeedFarmRating,omitempty"`
	AvgBossingRating   *float32  `db:"avg_bossing_rating" json:"avgBossingRating,omitempty"`
	Views              int       `db:"views" json:"views"`
	Status             int       `db:"status" json:"status"`
	CreatedAt          string    `db:"created_at" json:"createdAt"`
}

type BuildListResponse struct {
	ID                 uuid.UUID    `json:"id"`
	Title              string       `json:"title"`
	Description        string       `json:"description"`
	Class              string       `json:"class"`
	Ascendancy         *string      `json:"ascendancy"`
	MainSkillName      string       `json:"mainSkill"`
	AvgEndGameRating   *float32     `json:"avgEndGameRating,omitempty"`
	AvgFunRating       *float32     `json:"avgFunRating,omitempty"`
	AvgCreativeRating  *float32     `json:"avgCreativeRating,omitempty"`
	AvgSpeedFarmRating *float32     `json:"avgSpeedFarmRating,omitempty"`
	AvgBossingRating   *float32     `json:"avgBossingRating,omitempty"`
	Views              int          `json:"views"`
	Tags               []models.Tag `json:"tags"`
	Status             int          `json:"status"`
	CreatedAt          string       `json:"createdAt"`
}

type BuildItemSetResponse struct {
	BuildId uuid.UUID `db:"build_id"`
	SetId   uuid.UUID `db:"set_id"`
	ItemId  uuid.UUID `db:"item_id"`
	SetSlot string    `db:"set_slot"`
	// MemberID    uuid.UUID `json:"memberId" db:"member_id"`
	// BaseItemId  uuid.UUID `json:"baseItemId,omitempty" db:"base_item_id"`
	ImageUrl    string `json:"imageUrl" db:"image_url"`
	Category    string `json:"category" db:"category"`
	Class       string `json:"class" db:"class"`
	Name        string `json:"name" db:"name"`
	Type        string `json:"type" db:"type"`
	Description string `json:"description" db:"description"`
	UniqueItem  bool   `json:"uniqueItem" db:"unique_item"`
	Slot        string `json:"slot" db:"slot"`
	// armor
	RequiredLevel        *string `json:"requiredLevel,omitempty" db:"required_level"`
	RequiredStrength     *string `json:"requiredStrength,omitempty" db:"required_strength"`
	RequiredDexterity    *string `json:"requiredDexterity,omitempty" db:"required_dexterity"`
	RequiredIntelligence *string `json:"requiredIntelligence,omitempty" db:"required_intelligence"`
	Armour               *string `json:"armour,omitempty" db:"armour"`
	EnergyShield         *string `json:"energyShield,omitempty" db:"energy_shield"`
	Evasion              *string `json:"evasion,omitempty" db:"evasion"`
	Block                *string `json:"block,omitempty" db:"block"`
	Ward                 *string `json:"ward,omitempty" db:"ward"`
	// weapon
	Damage *string `json:"damage,omitempty" db:"damage"`
	APS    *string `json:"aps,omitempty" db:"aps"`
	Crit   *string `json:"crit,omitempty" db:"crit"`
	PDPS   *string `json:"pdps,omitempty" db:"pdps"`
	EDPS   *string `json:"edps,omitempty" db:"edps"`
	DPS    *string `json:"dps,omitempty" db:"dps"`
	// poison
	Life     *string `json:"life,omitempty" db:"life"`
	Mana     *string `json:"mana,omitempty" db:"mana"`
	Duration *string `json:"duration,omitempty" db:"duration"`
	Usage    *string `json:"usage,omitempty" db:"usage"`
	Capacity *string `json:"capacity,omitempty" db:"capacity"`
	// common
	Additional *string         `json:"additional,omitempty" db:"additional"`
	Stats      pq.StringArray  `json:"stats,omitempty" db:"stats"`
	Implicit   *pq.StringArray `json:"implicit,omitempty" db:"implicit"`
}

type BuildStatus int

const (
	draft     BuildStatus = 0
	published BuildStatus = 1
	archived  BuildStatus = 2
)
