package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

var sheetURL = "https://script.google.com/macros/s/AKfycbw9pq16kQqeksqDeED1JJAHnCM31uiOo270QQT3hbCj7kCEmS1qbMWmLyYAcJoGycX6/exec"

// BACA data dari Google Sheet
func SheetTampil(c *gin.Context) {
	res, err := http.Get(sheetURL)
	if err != nil {
		c.JSON(500, gin.H{"status": false, "pesan": "Gagal ambil data"})
		return
	}
	defer res.Body.Close()
	hasilBody, _ := io.ReadAll(res.Body)
	var hasilJson interface{}
	json.Unmarshal(hasilBody, &hasilJson)
	c.JSON(200, gin.H{"status": true, "data": hasilJson})
}

// TULIS data baru ke Google Sheet
func SheetTambah(c *gin.Context) {
	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"status": false, "pesan": "Data tidak valid"})
		return
	}
	postBody, _ := json.Marshal(input)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
	res, err := client.Post(sheetURL, "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		c.JSON(500, gin.H{"status": false, "pesan": "Gagal kirim data"})
		return
	}
	defer res.Body.Close()
	hasilBody, _ := io.ReadAll(res.Body)
	var hasilJson interface{}
	json.Unmarshal(hasilBody, &hasilJson)
	c.JSON(200, gin.H{"status": true, "data": hasilJson})
}
