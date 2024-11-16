package models

import (
	"github.com/google/uuid"
	"time"
)

/**
* Types here are shared model entities that are imported by more than one package.
**/

/**
* Member
**/
type Member struct {
	BaseDBDateModel
	Email         string  `db:"email" json:"email"`
	Name          string  `db:"name" json:"name"`
	Password      string  `db:"password" json:"password,omitempty"`
	Status        string  `db:"status" json:"status"`
	AverageRating float64 `db:"average_rating"`
	ResponseTime  int     `db:"response_time"`
	TotalTrades   int     `db:"total_trades"`
}

type Rating struct {
	BaseIDModel
	MemberID uuid.UUID `db:"member_id" json:"memberId"`
	Rating   int       `db:"rating" json:"rating"`
}

/**
* Items
**/

type Item struct {
	BaseDBDateModel
	MemberID      uuid.UUID `db:"member_id" json:"memberId"`
	ProductID     uuid.UUID `db:"product_id" json:"productId"`
	Category      string    `db:"category" json:"category"`
	Type          string    `db:"type" json:"type"`
	Name          string    `db:"name" json:"name"`
	Description   string    `db:"description" json:"description"`
	PricePerUnit  float64   `db:"price_per_unit" json:"pricePerUnit"`
	StockQuantity int       `db:"stock_quantity" json:"stockQuantity"`
}

/**
* Skills
**/

type Skill struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Type      string    `db:"type" json:"type"` // "active" or "support"
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

/**
* Build
**/
type Build struct {
	BaseDBDateModel
	MemberID    uuid.UUID `db:"member_id" json:"memberId"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	MainSkill   string    `db:"main_skill" json:"mainSkill"`
}

type BuildItem struct {
	ID      uuid.UUID `db:"id" json:"id"`
	BuildID uuid.UUID `db:"build_id" json:"buildId"`
	ItemID  uuid.UUID `db:"item_id" json:"itemId"`
	Slot    string    `db:"slot" json:"slot"`
}

type BuildSkill struct {
	ID      uuid.UUID `db:"id" json:"id"`
	BuildID uuid.UUID `db:"build_id" json:"buildId"`
	SkillID uuid.UUID `db:"skill_id" json:"skillId"`
}

/**
* Base models for default table columns.
**/

type BaseIDModel struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type BaseDBMemberModel struct {
	ID            uuid.UUID `db:"id" json:"id"`
	UpdatedMember uuid.UUID `db:"updated_member" json:"updatedMember"`
	CreatedMember uuid.UUID `db:"created_member" json:"createdMember"`
}

type BaseDBDateModel struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type BaseDBMemberDateModel struct {
	ID            uuid.UUID `db:"id" json:"id"`
	UpdatedMember uuid.UUID `db:"updated_member" json:"updatedMember"`
	CreatedMember uuid.UUID `db:"created_member" json:"createdMember"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}