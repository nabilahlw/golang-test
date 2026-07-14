package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"main/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DriveUpload(c *gin.Context) {
	fileName := c.PostForm("fileName")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"kode_error": "ERR-FORM",
			"pesan":      "File tidak ditemukan",
		})
		return
	}

	mimeType := file.Header.Get("Content-Type")
	fileOpen, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"kode_error": "ERR-OPEN",
			"pesan":      "Gagal membuka file",
		})
		return
	}
	defer fileOpen.Close()

	fileData, err := io.ReadAll(fileOpen)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"kode_error": "ERR-READ",
			"pesan":      "Gagal membaca file",
		})
		return
	}

	data := base64.StdEncoding.EncodeToString(fileData)

	postBody, err := json.Marshal(map[string]string{
		"fileName": fileName,
		"mimeType": mimeType,
		"data":     data,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"kode_error": "ERR-JSON",
			"pesan":      "Gagal membuat request body",
		})
		return
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
	res, err := client.Post(
		"https://script.google.com/macros/s/AKfycbxyukxFAYC0rs-8s8C0YOJqJVPVATFNP8GQ33Wq9sC-hUhwXPpTpq96mPsEnsesGqQBZA/exec",
		"application/json; charset=UTF-8",
		bytes.NewBuffer(postBody),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"kode_error": "ERR-DRIVE",
			"pesan":      "Gagal upload ke Drive",
		})
		return
	}
	defer res.Body.Close()

	hasilBody, err := io.ReadAll(res.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"kode_error": "ERR-RESPONSE",
			"pesan":      "Gagal membaca response",
		})
		return
	}

	var hasilJson map[string]interface{}
	if err := json.Unmarshal(hasilBody, &hasilJson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"kode_error": "ERR-PARSE",
			"pesan":      "Gagal parse response",
			"raw":        string(hasilBody),
		})
		return
	}

	// Simpan ke tabel dokumens
	db := c.MustGet("db").(*gorm.DB)
	dokumenBaru := models.Dokumen{
		NamaDokumen: hasilJson["filename"].(string),
		FileId:      hasilJson["fileId"].(string),
		FileUrl:     hasilJson["fileUrl"].(string),
	}
	hasilDokumen := db.Create(&dokumenBaru)

	c.JSON(http.StatusOK, gin.H{
		"status":    true,
		"pesan":     "Berhasil Upload",
		"data":      hasilJson,
		"tersimpan": hasilDokumen.RowsAffected,
	})
}

func DriveTampil(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var dokumen []models.Dokumen
	db.Find(&dokumen)
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"pesan":  "Berhasil Tampil",
		"data":   dokumen,
	})
}

func DriveUnduh(c *gin.Context) {
	id := c.Param("id")
	res, err := http.Get("https://script.google.com/macros/s/AKfycbxyukxFAYC0rs-8s8C0YOJqJVPVATFNP8GQ33Wq9sC-hUhwXPpTpq96mPsEnsesGqQBZA/exec?id=" + id)
	if err != nil {
		c.JSON(500, gin.H{
			"status": false,
			"pesan":  "Gagal Unduh",
		})
		return
	}
	defer res.Body.Close()

	hasilBody, _ := io.ReadAll(res.Body)
	var hasilJson map[string]interface{}
	json.Unmarshal(hasilBody, &hasilJson)

	fileBase64 := hasilJson["file"].(string)
	mimeType := hasilJson["mimeType"].(string)

	file, _ := base64.StdEncoding.DecodeString(fileBase64)
	c.Writer.Header().Set("Content-Type", mimeType)
	c.Writer.Write(file)
}
