package controllers

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// DriveUpload adalah fungsi untuk menangani upload via API/Postman
func DriveUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "pesan": "File tidak ditemukan"})
		return
	}

	filePath := "./temp/" + file.Filename
	c.SaveUploadedFile(file, filePath)

	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsFile("/root/PCC/service-account.json"))
	if err != nil {
		os.Stderr.WriteString("ERROR DRIVER: " + err.Error() + "\n") 
    c.JSON(http.StatusInternalServerError, gin.H{"status": false, "pesan": "Gagal inisialisasi: " + err.Error()})
    return
	}

	f, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "pesan": "Gagal membuka file"})
		return
	}
	defer f.Close()

	fileMetadata := &drive.File{Name: file.Filename}
	_, err = srv.Files.Create(fileMetadata).Media(f).Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "pesan": "Gagal upload ke Drive"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "pesan": "Berhasil upload ke Google Drive"})
}

// DriveTampil untuk melihat daftar file di Drive
func DriveTampil(c *gin.Context) {
	ctx := context.Background()
	srv, _ := drive.NewService(ctx, option.WithCredentialsFile("service-account.json"))

	r, _ := srv.Files.List().PageSize(10).Do()
	c.JSON(http.StatusOK, gin.H{"files": r.Files})
}

// DriveUnduh untuk mengunduh file berdasarkan ID
func DriveUnduh(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	srv, _ := drive.NewService(ctx, option.WithCredentialsFile("service-account.json"))

	resp, err := srv.Files.Get(id).Download()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "pesan": "Gagal mengunduh file"})
		return
	}
	defer resp.Body.Close()

	c.DataFromReader(http.StatusOK, resp.ContentLength, "application/octet-stream", resp.Body, nil)
}
