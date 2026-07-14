// ID FOLDER = 1vgGzfojOmhMGw6PCGEUaIdRqsMR_KFsVpWviLxkBKVkAUSjqoyjdONOb
// URL APK WEB = https://script.google.com/macros/s/AKfycbzqiD0BGQg3S3kNPQAmDkhN-SwK_NTxoOEyy2vMIFmDMfEhItUbuM48Ol4hqUXEtDRRUQ/exec
// id PENERAPAN = AKfycbzqiD0BGQg3S3kNPQAmDkhN-SwK_NTxoOEyy2vMIFmDMfEhItUbuM48Ol4hqUXEtDRRUQ
// KALO AMBBIL JWT TOKEN, POST LOGIN http://localhost:8111/login = "username": "admin", "password": "123456"
// ACCESS ID = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODA2NjM1MjUsImlkIjoxLCJuYW1hIjoiQWRtaW4iLCJvcmlnX2lhdCI6MTc4MDY1OTkyNX0.UTN5sQCxRQKFNSjvqqOKfm4Qv3fP33frDkEuQgBPmdA
// url drive pcc = 1jEX5u_O4z9POMEFRX4Aa2nEYGybiKUqm

package main

import (
	"log"
	"os"
	"time"

	"main/ai"
	"main/controllers"
	"main/fungsi"
	"main/models"
	"main/wa"

	jwtV3 "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	db := koneksi()
	db.AutoMigrate(&models.Suhu{})
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Dokumen{})
	db.AutoMigrate(&models.Pesan{})

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	key_jwt := os.Getenv("KEY_JWT")
	authMiddleware, err := jwtV3.New(&jwtV3.GinJWTMiddleware{
		Realm:       "fikom UDB",
		Key:         []byte(key_jwt),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour * 24,
		IdentityKey: "id",
		PayloadFunc: func(data any) jwt.MapClaims {
			value, ok := data.(models.User)
			if ok {
				return jwt.MapClaims{
					"id":   value.ID,
					"nama": value.Nama,
				}
			}
			return jwt.MapClaims{}
		},
		Authenticator: controllers.UserLogin,
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	errInit := authMiddleware.MiddlewareInit()
	if errInit != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}

	// route tanpa middleware
	r.POST("/login", authMiddleware.LoginHandler)

	// route group dengan middleware jwt
	auth := r.Group("/backend", authMiddleware.MiddlewareFunc())
	{
		auth.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": true,
				"pesan":  "Berhasil tampil",
			})
		})

		auth.POST("/programstudi", fungsi.BacaDataProdi)

		auth.GET("/suhu", controllers.Tampil)
		auth.POST("/suhu", controllers.Tambah)
		auth.PUT("/suhu", controllers.Ubah)
		auth.DELETE("/suhu", controllers.Hapus)
		auth.POST("/drive", controllers.DriveUpload)
		auth.GET("/user", controllers.UserTampil)
		auth.POST("/user", controllers.UserTambah)
		auth.PUT("/user", controllers.UserUbah)
		auth.DELETE("/user", controllers.UserHapus)
		auth.GET("/drive", controllers.DriveTampil)
		auth.GET("/drive/:id", controllers.DriveUnduh)
		auth.GET("/sheet", controllers.SheetTampil)
		auth.POST("/sheet", controllers.SheetTambah)
		auth.GET("/pesan", controllers.PesanTampil)
		auth.POST("/pesan", controllers.PesanTambah)
		auth.PUT("/pesan", controllers.PesanUbah)
		auth.DELETE("/pesan", controllers.PesanHapus)
	}

	port := os.Getenv("PORT")
	go r.Run(":" + port)
	ai.InitAi()
	go wa.KonekWa(db)
	ai.MulaiChatAi()
}

// cara jalanin ilama, di folder ilama =  ./llama-server -m qwen2.5-0.5b-instruct-q4_k_m.gguf -c 2048 --port 8080
//trs ke terminal projek PCC llau = go run main.go koneksi.go
