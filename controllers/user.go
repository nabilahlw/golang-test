package controllers

import (
	"crypto/sha1"
	"fmt"
	"log"
	"net/http"

	"main/models"

	jwtV3 "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StrukturUserTambah struct {
	Nama     string `binding:"required"`
	Username string `binding:"required"`
	Password string `binding:"required"`
}

type StrukturUserUbah struct {
	Id       uint
	Nama     string `binding:"required"`
	Username string `binding:"required"`
	Password string `binding:"required"`
}

type StrukturUserHapus struct {
	Id uint
}

type StrukturLogin struct {
	Username string `binding:"required"`
	Password string `binding:"required"`
}

func UserTampil(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var modelUser []models.User
	hasil := db.Find(&modelUser)
	kesalahan := hasil.Error
	if hasil.Error == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":    true,
			"pesan":     "Berhasil Tampil data",
			"kesalahan": nil,
			"data":      modelUser,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":    false,
			"pesan":     "Gagal Tampil Data",
			"kesalahan": kesalahan.Error(),
			"data":      nil,
		})
	}
}

func UserTambah(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var dataUser StrukturUserTambah
	if err := c.ShouldBindJSON(&dataUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    false,
			"pesan":     "Gagal membaca Data",
			"kesalahan": err.Error(),
		})
		return
	}
	var sha = sha1.New()
	sha.Write([]byte(dataUser.Password))
	var encrypted = sha.Sum(nil)
	var encryptedString = fmt.Sprintf("%x", encrypted)

	modelUser := models.User{
		Nama:     dataUser.Nama,
		Username: dataUser.Username,
		Password: encryptedString,
	}
	hasil := db.Create(&modelUser)
	kesalahan := hasil.Error
	if hasil.Error == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":    true,
			"pesan":     "Berhasil tambah data",
			"kesalahan": nil,
			"data":      modelUser,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":    false,
			"pesan":     "Gagal Tambah Data",
			"kesalahan": kesalahan.Error(),
			"data":      modelUser,
		})
	}
}

func UserUbah(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var dataUser StrukturUserUbah
	if err := c.ShouldBindJSON(&dataUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    false,
			"pesan":     "Gagal membaca Data",
			"kesalahan": err.Error(),
		})
		return
	}
	var modelUser models.User
	cekUser := db.First(&modelUser, dataUser.Id)
	if cekUser.Error == nil {
		var sha = sha1.New()
		sha.Write([]byte(dataUser.Password))
		var encrypted = sha.Sum(nil)
		var encryptedString = fmt.Sprintf("%x", encrypted)

		modelUser.Nama = dataUser.Nama
		modelUser.Username = dataUser.Username
		modelUser.Password = encryptedString
		hasil := db.Save(&modelUser)
		kesalahan := hasil.Error
		if hasil.Error == nil {
			c.JSON(http.StatusOK, gin.H{
				"status":    true,
				"pesan":     "Berhasil ubah data",
				"kesalahan": nil,
				"data":      modelUser,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":    false,
				"pesan":     "Gagal ubah Data",
				"kesalahan": kesalahan.Error(),
				"data":      modelUser,
			})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":    false,
			"pesan":     "Data Tidak ditemukan",
			"kesalahan": cekUser.Error.Error(),
			"data":      modelUser,
		})
	}
}

func UserHapus(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var dataUser StrukturUserHapus
	if err := c.ShouldBindJSON(&dataUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    false,
			"pesan":     "Gagal membaca Data",
			"kesalahan": err.Error(),
		})
		return
	}
	var modelUser models.User
	hasil := db.Delete(&modelUser, dataUser.Id)
	kesalahan := hasil.Error
	if hasil.Error == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":    true,
			"pesan":     "Berhasil hapus data",
			"kesalahan": nil,
			"data":      dataUser,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":    false,
			"pesan":     "Gagal hapus Data",
			"kesalahan": kesalahan.Error(),
			"data":      dataUser,
		})
	}
}

func UserLogin(c *gin.Context) (any, error) {
	db := c.MustGet("db").(*gorm.DB)
	var dataUser StrukturLogin
	if err := c.ShouldBindJSON(&dataUser); err != nil {
		return nil, jwtV3.ErrMissingLoginValues
	}
	var sha = sha1.New()
	sha.Write([]byte(dataUser.Password))
	var encrypted = sha.Sum(nil)
	var encryptedString = fmt.Sprintf("%x", encrypted)

	var modelUser models.User
	cekUser := db.Where("username = ?", dataUser.Username).Where("password = ?", encryptedString).First(&modelUser)
	if cekUser.Error == nil {
		return modelUser, nil
	} else {
		return nil, jwtV3.ErrFailedAuthentication
	}
}

// suppress unused import warning
var _ = log.Fatal
