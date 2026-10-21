package store

import (
	"context"
)

// VoteMovie บันทึกหรืออัปเดตคะแนนโหวต (Upsert)
func (s *Store) VoteMovie(ctx context.Context, memberID, movieID int64, score int) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO votes (member_id, movie_id, score, created_at, updated_at)
		VALUES ($1, $2, $3, now(), now())
		ON CONFLICT (member_id, movie_id)
		DO UPDATE SET score = EXCLUDED.score, updated_at = now()
	`, memberID, movieID, score)
	return err
}

// CreateReview บันทึกรีวิวหนัง
func (s *Store) CreateReview(ctx context.Context, memberID, movieID int64, text string) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO reviews (member_id, movie_id, text, created_at, updated_at)
		VALUES ($1, $2, $3, now(), now())
	`, memberID, movieID, text)
	return err
}

// AddWishlist เพิ่มหนังเข้า wishlist
func (s *Store) AddWishlist(ctx context.Context, memberID, movieID int64) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO wishlist_items (member_id, movie_id, added_at)
		VALUES ($1, $2, now())
		ON CONFLICT DO NOTHING
	`, memberID, movieID)
	return err
}

// RemoveWishlist ลบหนังออกจาก wishlist
func (s *Store) RemoveWishlist(ctx context.Context, memberID, movieID int64) error {
	_, err := s.Pool.Exec(ctx, `
		DELETE FROM wishlist_items
		WHERE member_id = $1 AND movie_id = $2
	`, memberID, movieID)
	return err
}
