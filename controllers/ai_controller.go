package controllers

import (
    "github.com/gin-gonic/gin"
    "main/ai"
)

func HandleAi(c *gin.Context) {
    var input struct {
        Pertanyaan string `json:"pertanyaan"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": "Input tidak valid"})
        return
    }

    // Memanggil fungsi AI yang sudah ada di proyekmu
    jawaban := ai.TanyaAi("admin", input.Pertanyaan)

    c.JSON(200, gin.H{
        "status": "success",
        "jawaban": jawaban,
    })
}
