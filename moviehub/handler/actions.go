package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"moviehub/mw"
)

type voteReq struct {
	Score int `json:"score"`
}

type reviewReq struct {
	Text string `json:"text"`
}

type wishlistReq struct {
	Note *string `json:"note"`
}

// PUT /api/movies/{id}/vote
func (h *Handler) VoteMovie(w http.ResponseWriter, r *http.Request) {
	memberID, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}
	movieID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_ID", "รหัสหนังไม่ถูกต้อง")
		return
	}

	var req voteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
		return
	}
	if req.Score < 1 || req.Score > 10 {
		writeError(w, http.StatusBadRequest, "BAD_SCORE", "คะแนนต้องอยู่ระหว่าง 1 ถึง 10")
		return
	}

	if err := h.Store.VoteMovie(r.Context(), memberID, movieID, req.Score); err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// DELETE /api/movies/{id}/vote
func (h *Handler) DeleteVote(w http.ResponseWriter, r *http.Request) {
	memberID, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}
	movieID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_ID", "รหัสหนังไม่ถูกต้อง")
		return
	}

	if err := h.Store.DeleteVote(r.Context(), memberID, movieID); err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// POST /api/movies/{id}/reviews
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	memberID, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}
	movieID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_ID", "รหัสหนังไม่ถูกต้อง")
		return
	}

	var req reviewReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "รูปแบบ JSON ไม่ถูกต้อง")
		return
	}
	trimmed := strings.TrimSpace(req.Text)
	if utf8.RuneCountInString(trimmed) < 10 {
		writeError(w, http.StatusBadRequest, "BAD_REVIEW", "รีวิวต้องยาวอย่างน้อย 10 ตัวอักษร")
		return
	}

	if err := h.Store.CreateReview(r.Context(), memberID, movieID, trimmed); err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true})
}

// GET /api/me/wishlist
func (h *Handler) GetMyWishlist(w http.ResponseWriter, r *http.Request) {
	memberID, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}

	items, err := h.Store.ListMyWishlist(r.Context(), memberID)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// PUT /api/me/wishlist/{movieId}
func (h *Handler) PutWishlist(w http.ResponseWriter, r *http.Request) {
	memberID, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}
	movieID, err := strconv.ParseInt(r.PathValue("movieId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_ID", "รหัสหนังไม่ถูกต้อง")
		return
	}

	var req wishlistReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.Store.PutWishlist(r.Context(), memberID, movieID, req.Note); err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// DELETE /api/me/wishlist/{movieId}
func (h *Handler) DeleteWishlist(w http.ResponseWriter, r *http.Request) {
	memberID, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}
	movieID, err := strconv.ParseInt(r.PathValue("movieId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_ID", "รหัสหนังไม่ถูกต้อง")
		return
	}

	if err := h.Store.DeleteWishlist(r.Context(), memberID, movieID); err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}