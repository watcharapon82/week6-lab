package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"moviehub/auth"
	"moviehub/handler"
	"moviehub/mw"
	"moviehub/store"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secretFile, err := os.ReadFile(".jwt_secret")
		if err != nil {
			log.Fatal("JWT_SECRET ไม่ได้ตั้งค่า และอ่านไฟล์ .jwt_secret ไม่ได้: ", err)
		}
		secret = strings.TrimSpace(string(secretFile))
	}
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET ต้องยาวอย่างน้อย 32 ตัวอักษร")
	}
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("ต่อฐานข้อมูลไม่ติด ", err)
	}

	tickets := &auth.Issuer{Secret: []byte(secret), TTL: time.Hour}
	h := &handler.Handler{Store: &store.Store{Pool: pool}, Tickets: tickets}

	// login = ต้องมีตั๋ว, admin = ต้องมีตั๋วและบทบาทเป็น admin (ตรวจตั๋วก่อนเสมอ ค่อยดูบทบาท)
	login := func(next http.HandlerFunc) http.HandlerFunc { return mw.RequireLogin(tickets, next) }
	admin := func(next http.HandlerFunc) http.HandlerFunc { return login(mw.RequireRole("admin", next)) }

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)

	mux.HandleFunc("GET /api/me", login(h.Me))
	mux.HandleFunc("GET /api/genres", h.Genres)

	// หมวดภาพยนตร์
	mux.HandleFunc("GET /api/movies", h.SearchMovies)
	mux.HandleFunc("GET /api/movies/{id}", h.GetMovie)

	// หมวดรีวิวหนัง (ดูรีวิวไม่ต้องล็อกอิน / ส่งรีวิวต้องล็อกอิน)
	mux.HandleFunc("GET /api/movies/{id}/reviews", h.ListReviews)
	mux.HandleFunc("POST /api/movies/{id}/reviews", login(h.CreateReview))

	// ให้คะแนนหนัง (Vote)
	mux.HandleFunc("POST /api/movies/{id}/vote", login(h.VoteMovie))

	// รายการที่อยากดู (Wishlist)
	mux.HandleFunc("GET /api/me/wishlist", login(h.ListWishlist))
	mux.HandleFunc("POST /api/me/wishlist/{id}", login(h.AddToWishlist))
	mux.HandleFunc("DELETE /api/me/wishlist/{id}", login(h.RemoveFromWishlist))

	// หลังบ้าน: ทุกเส้นทางใต้ /api/admin ผ่าน admin(...) ไม่มีข้อยกเว้น
	mux.HandleFunc("GET /api/admin/stats", admin(h.AdminStats))
	mux.HandleFunc("GET /api/admin/members", admin(h.AdminMembers))
	mux.HandleFunc("PUT /api/admin/members/{id}/role", admin(h.AdminSetRole))
	mux.HandleFunc("GET /api/admin/movies/{id}", admin(h.AdminGetMovie))
	mux.HandleFunc("POST /api/admin/movies", admin(h.AdminCreateMovie))
	mux.HandleFunc("PUT /api/admin/movies/{id}", admin(h.AdminUpdateMovie))
	mux.HandleFunc("DELETE /api/admin/movies/{id}", admin(h.AdminDeleteMovie))

	origins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	port := os.Getenv("PORT") // Render ส่งมาให้ ในเครื่องไม่มีก็ใช้ 8080
	if port == "" {
		port = "8080"
	}

	log.Println("MovieHub API ฟังอยู่ที่ :"+port, "ต้อนรับ", origins)
	log.Fatal(http.ListenAndServe(":"+port, mw.Logging(mw.CORS(origins, mux))))
}