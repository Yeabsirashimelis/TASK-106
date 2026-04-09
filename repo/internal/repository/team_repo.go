package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(ctx context.Context, team *models.Team) error {
	query := `INSERT INTO teams (id, name, season_id, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, team.ID, team.Name, team.SeasonID, team.CreatedAt)
	if err != nil {
		return fmt.Errorf("team_repo.Create: %w", err)
	}
	return nil
}

func (r *TeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	var team models.Team
	query := `SELECT id, name, season_id, created_at FROM teams WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&team); err != nil {
		return nil, fmt.Errorf("team_repo.GetByID: %w", err)
	}
	return &team, nil
}

func (r *TeamRepository) ListBySeason(ctx context.Context, seasonID uuid.UUID) ([]models.Team, error) {
	var teams []models.Team
	query := `SELECT id, name, season_id, created_at FROM teams WHERE season_id = $1 ORDER BY name`
	if err := r.db.SelectContext(ctx, &teams, query, seasonID); err != nil {
		return nil, fmt.Errorf("team_repo.ListBySeason: %w", err)
	}
	return teams, nil
}
