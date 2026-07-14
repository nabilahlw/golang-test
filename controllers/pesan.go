package controllers

import (
	"net/http"

	"main/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PesanTampil(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var pesan []models.Pesan
	db.Find(&pesan)
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"pesan":  "Berhasil Tampil",
		"data":   pesan,
	})
}

func PesanTambah(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var input models.Pesan
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"pesan":  "Gagal membaca data",
		})
		return
	}
	hasil := db.Create(&input)
	if hasil.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"pesan":  "Gagal tambah data",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"pesan":  "Berhasil tambah data",
		"data":   input,
	})
}

func PesanUbah(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var input models.Pesan
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"pesan":  "Gagal membaca data",
		})
		return
	}
	var pesan models.Pesan
	cek := db.Where("kode = ?", input.Kode).First(&pesan)
	if cek.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"pesan":  "Data tidak ditemukan",
		})
		return
	}
	pesan.Balasan = input.Balasan
	db.Save(&pesan)
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"pesan":  "Berhasil ubah data",
		"data":   pesan,
	})
}

func PesanHapus(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var input struct {
		Kode string
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"pesan":  "Gagal membaca data",
		})
		return
	}
	db.Delete(&models.Pesan{}, "kode = ?", input.Kode)
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"pesan":  "Berhasil hapus data",
	})
}
