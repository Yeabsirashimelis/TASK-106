package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SeasonRepository struct {
	db *sqlx.DB
}

func NewSeasonRepository(db *sqlx.DB) *SeasonRepository {
	return &SeasonRepository{db: db}
}

func (r *SeasonRepository) Create(ctx context.Context, season *models.Season) error {
	query := `INSERT INTO seasons (id, name, start_date, end_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		season.ID, season.Name, season.StartDate, season.EndDate,
		season.Status, season.CreatedAt, season.UpdatedAt)
	if err != nil {
		return fmt.Errorf("season_repo.Create: %w", err)
	}
	return nil
}

func (r *SeasonRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Season, error) {
	var season models.Season
	query := `SELECT id, name, start_date, end_date, status, created_at, updated_at FROM seasons WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&season); err != nil {
		return nil, fmt.Errorf("season_repo.GetByID: %w", err)
	}
	return &season, nil
}

func (r *SeasonRepository) List(ctx context.Context, offset, limit int) ([]models.Season, error) {
	var seasons []models.Season
	query := `SELECT id, name, start_date, end_date, status, created_at, updated_at
		FROM seasons ORDER BY start_date DESC LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &seasons, query, limit, offset); err != nil {
		return nil, fmt.Errorf("season_repo.List: %w", err)
	}
	return seasons, nil
}

func (r *SeasonRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.SeasonStatus) error {
	query := `UPDATE seasons SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("season_repo.UpdateStatus: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("season_repo.UpdateStatus: season not found")
	}
	return nil
}
