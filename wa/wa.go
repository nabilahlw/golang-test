package wa

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"

	"main/ai"
	"main/models"
)

var clientWa *whatsmeow.Client
var DB *gorm.DB

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		// Mengambil isi teks dengan menangani beberapa kemungkinan tipe pesan
		var pesan string
		if v.Message.GetConversation() != "" {
			pesan = v.Message.GetConversation()
		} else if v.Message.ExtendedTextMessage != nil {
			pesan = v.Message.ExtendedTextMessage.GetText()
		}

		// Jika pesan kosong, abaikan (mungkin pesan gambar/stiker)
		if pesan == "" {
			return
		}

		fmt.Println("DEBUG: Pesan ditangkap:", pesan)
		fmt.Println("SERVER:", v.Info.MessageSource.Chat.Server)
fmt.Println("IS GROUP:", v.Info.IsGroup)
fmt.Println("IS BROADCAST:", v.Info.IsIncomingBroadcast())
		// Filter chat pribadi saja
		if !v.Info.IsGroup && !v.Info.IsIncomingBroadcast() {
			
            fmt.Println("DEBUG: Masuk ke dalam filter chat") // TAMBAHKAN INI

			// Cek apakah pesan mengandung [ai] dengan cara yang lebih toleran
			if strings.Contains(strings.ToLower(pesan), "[ai]") {
				fmt.Println("DEBUG: AI terdeteksi!")
				
				// Paksa ambil teks setelah [ai]
				parts := strings.SplitN(strings.ToLower(pesan), "[ai]", 2)
				pertanyaan := strings.TrimSpace(parts[1])

				if pertanyaan == "" {
					kirimPesanText(v.Info.Sender, "Ya, ada yang bisa saya bantu?")
				} else {
					jawaban := ai.TanyaAi(v.Info.Sender.User, pertanyaan)

fmt.Printf("SENDER     = %+v\n", v.Info.Sender)
fmt.Printf("CHAT       = %+v\n", v.Info.Chat)
fmt.Printf("SOURCECHAT = %+v\n", v.Info.MessageSource.Chat)

kirimPesanText(v.Info.Sender, jawaban)
				}
            }
		}
	}
}

func kirimPesan(IDPenerima types.JID) {
	clientWa.SendMessage(context.Background(), IDPenerima, &waE2E.Message{
		Conversation: proto.String("[UJI COBA] \n PESAN OTOMATIS BERHASIL"),
	})
}

func kirimPesanText(IDPenerima types.JID, text string) {
	if text == "" {
		text = "Maaf, AI tidak memberikan jawaban."
	}
	// Hilangkan device part (:35)
    IDPenerima.Device = 0

    fmt.Println("Mengirim WA:", text)
    fmt.Println("JID:", IDPenerima.String())

    _, err := clientWa.SendMessage(
        context.Background(),
        IDPenerima,
        &waE2E.Message{
            Conversation: proto.String(text),
        },
    )

    if err != nil {
        fmt.Println("ERROR SEND WA:", err)
    }
}
func kirimPesanDatabase(IDPenerima types.JID, kode string) {
	var pesanDB models.Pesan
	result := DB.Where("kode = ?", kode).First(&pesanDB)
	if result.Error == nil {
		kirimPesanText(IDPenerima, pesanDB.Balasan)
	}
}

func KonekWa(db *gorm.DB) {
	DB = db
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		client.Connect()
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			}
		}
	} else {
		client.Connect()
	}

	clientWa = client
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	client.Disconnect()
}
