package store

import (
	"context"
	"strings"
	"time"
)

type WishlistItem struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	TitleTH    *string   `json:"titleTh"`
	Genre      *string   `json:"genre"`
	Year       *int      `json:"year"`
	Rating     *float64  `json:"rating"`
	Poster     *string   `json:"poster"`
	PosterPath *string   `json:"posterPath"`
	Note       *string   `json:"note"`
	AddedAt    time.Time `json:"addedAt"`
}

// VoteMovie บันทึกหรือเปลี่ยนคะแนนโหวต (PUT)
func (s *Store) VoteMovie(ctx context.Context, memberID, movieID int64, score int) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO votes (member_id, movie_id, score, created_at, updated_at)
		VALUES ($1, $2, $3, now(), now())
		ON CONFLICT (member_id, movie_id)
		DO UPDATE SET score = EXCLUDED.score, updated_at = now()
	`, memberID, movieID, score)
	return err
}

// DeleteVote ลบคะแนนโหวต (DELETE)
func (s *Store) DeleteVote(ctx context.Context, memberID, movieID int64) error {
	_, err := s.Pool.Exec(ctx, `
		DELETE FROM votes WHERE member_id = $1 AND movie_id = $2
	`, memberID, movieID)
	return err
}

// CreateReview บันทึกรีวิวหนัง (POST)
func (s *Store) CreateReview(ctx context.Context, memberID, movieID int64, text string) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO reviews (member_id, movie_id, text, created_at, updated_at)
		VALUES ($1, $2, $3, now(), now())
	`, memberID, movieID, text)
	return err
}

// ListMyWishlist ดึงรายการหนังที่อยากดูตาม Query ในสเปก
func (s *Store) ListMyWishlist(ctx context.Context, memberID int64) ([]WishlistItem, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT m.id, m.title, m.title_th, g.name_en AS genre,
		       EXTRACT(YEAR FROM m.release_date)::int AS year,
		       round(m.vote_average, 1) AS rating, m.poster_path,
		       w.note, w.added_at
		FROM wishlist_items w
		JOIN movies m             ON m.id = w.movie_id
		LEFT JOIN movie_genres mg ON mg.movie_id = m.id AND mg.position = 0
		LEFT JOIN genres g        ON g.id = mg.genre_id
		WHERE w.member_id = $1
		ORDER BY w.added_at DESC
	`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []WishlistItem
	for rows.Next() {
		var it WishlistItem
		var rawPoster *string
		if err := rows.Scan(&it.ID, &it.Title, &it.TitleTH, &it.Genre, &it.Year, &it.Rating, &rawPoster, &it.Note, &it.AddedAt); err != nil {
			return nil, err
		}

		if rawPoster != nil && *rawPoster != "" {
			p := *rawPoster
			it.PosterPath = &p

			// แปลงเป็น Full URL ของ TMDB หากเป็น relative path เริ่มต้นด้วย /
			if strings.HasPrefix(p, "/") {
				fullURL := "https://image.tmdb.org/t/p/w500" + p
				it.Poster = &fullURL
			} else {
				it.Poster = &p
			}
		}

		items = append(items, it)
	}
	if items == nil {
		items = []WishlistItem{}
	}
	return items, rows.Err()
}

// PutWishlist เพิ่มหรืออัปเดต note ใน wishlist
func (s *Store) PutWishlist(ctx context.Context, memberID, movieID int64, note *string) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO wishlist_items (member_id, movie_id, note)
		VALUES ($1, $2, $3)
		ON CONFLICT (member_id, movie_id)
		DO UPDATE SET note = EXCLUDED.note
	`, memberID, movieID, note)
	return err
}

// DeleteWishlist ลบหนังออกจาก wishlist
func (s *Store) DeleteWishlist(ctx context.Context, memberID, movieID int64) error {
	_, err := s.Pool.Exec(ctx, `
		DELETE FROM wishlist_items WHERE member_id = $1 AND movie_id = $2
	`, memberID, movieID)
	return err
}