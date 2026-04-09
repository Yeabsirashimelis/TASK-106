package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type VenueRepository struct {
	db *sqlx.DB
}

func NewVenueRepository(db *sqlx.DB) *VenueRepository {
	return &VenueRepository{db: db}
}

func (r *VenueRepository) Create(ctx context.Context, venue *models.Venue) error {
	query := `INSERT INTO venues (id, name, location, capacity, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, venue.ID, venue.Name, venue.Location, venue.Capacity, venue.CreatedAt)
	if err != nil {
		return fmt.Errorf("venue_repo.Create: %w", err)
	}
	return nil
}

func (r *VenueRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Venue, error) {
	var venue models.Venue
	query := `SELECT id, name, location, capacity, created_at FROM venues WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&venue); err != nil {
		return nil, fmt.Errorf("venue_repo.GetByID: %w", err)
	}
	return &venue, nil
}

func (r *VenueRepository) List(ctx context.Context, offset, limit int) ([]models.Venue, error) {
	var venues []models.Venue
	query := `SELECT id, name, location, capacity, created_at FROM venues ORDER BY name LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &venues, query, limit, offset); err != nil {
		return nil, fmt.Errorf("venue_repo.List: %w", err)
	}
	return venues, nil
}
