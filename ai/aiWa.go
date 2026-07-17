package ai

import (
	"context"
	"fmt"
	"sync"
	"os"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var client openai.Client
var userHistories = make(map[string][]openai.ChatCompletionMessageParamUnion)
var mu sync.Mutex

func InitAi() {
	apiKey := os.Getenv("OPENAI_API_KEY")
baseURL := os.Getenv("AI_BASE_URL")
    client = openai.NewClient(
        option.WithAPIKey(apiKey),
option.WithBaseURL(baseURL),
    )
	fmt.Println("AI Engine berhasil diinisialisasi.")
}

func TanyaAi(userID string, userInput string) string {
	if len(userInput) < 3 {
		return "Maaf, input terlalu pendek."
	}

	ctx := context.Background()
	instruksiSistem := "Anda adalah FikomBot, asisten virtual Fakultas Ilmu Komputer UDB Surakarta. Berikan jawaban yang singkat dan sopan."

	mu.Lock()
	chatHistory := userHistories[userID]
	mu.Unlock()

	var currentPayload []openai.ChatCompletionMessageParamUnion
	currentPayload = append(currentPayload, openai.SystemMessage(instruksiSistem))
	currentPayload = append(currentPayload, chatHistory...)
	currentPayload = append(currentPayload, openai.UserMessage(userInput))

	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModel("openai/gpt-oss-20b:free"),
		Messages: currentPayload,
	})
	fmt.Printf("RESP AI = %+v\n", resp)

	if err != nil {
		fmt.Printf("Error detail dari AI: %v\n", err)
		return "Mohon maaf, terjadi gangguan saat memproses jawaban."
	}

	var jawabanAi string
	if len(resp.Choices) > 0 {
		jawabanAi = resp.Choices[0].Message.Content
		fmt.Println("JAWABAN AI =", jawabanAi)
	} else {
		jawabanAi = "Mohon maaf, AI tidak memberikan respon"
	}

	mu.Lock()
	if len(userHistories[userID]) > 8 {
		userHistories[userID] = []openai.ChatCompletionMessageParamUnion{}
	}
	userHistories[userID] = append(
		userHistories[userID],
		openai.UserMessage(userInput),
		openai.AssistantMessage(jawabanAi),
	)
	mu.Unlock()

	return jawabanAi
}
