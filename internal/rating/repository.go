package rating

import (
	"fmt"

	"github.com/darkphotonKN/community-builds/internal/models"
	"github.com/darkphotonKN/community-builds/internal/utils/errorutils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RatingRepository struct {
	DB *sqlx.DB
}

func NewRatingRepository(db *sqlx.DB) *RatingRepository {
	return &RatingRepository{
		DB: db,
	}
}

func (r *RatingRepository) CreateRatingForBuildById(memberId uuid.UUID, request CreateRatingRequest) error {

	// add a new rating under this member id and build id
	query := `
	INSERT INTO ratings (build_id, member_id, value, category)
	VALUES (:build_id, :member_id, :value, :category)
	`

	params := map[string]interface{}{
		"build_id":  request.BuildId,
		"member_id": memberId,
		"value":     request.Value,
		"category":  request.Category,
	}

	_, err := r.DB.NamedExec(query, params)

	if err != nil {
		return errorutils.AnalyzeDBErr(err)
	}

	return nil
}

func (r *RatingRepository) GetAllRatingsByMemberId(memberId uuid.UUID) (*[]models.Rating, error) {
	var ratings []models.Rating

	query := `
	SELECT * FROM ratings
	WHERE ratings.member_id = $1
	`

	err := r.DB.Select(&ratings, query, memberId)

	fmt.Println("ratings:", ratings)

	if err != nil {
		return nil, err
	}

	return &ratings, nil
}
