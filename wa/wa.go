package wa

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
var pesan string
var DB *gorm.DB

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("Received a message!", v.Message.GetConversation())
		fmt.Println(" => dari saya sendiri = ", v.Info.IsFromMe)
		fmt.Println(" => server = ", v.Info.MessageSource.Chat.Server)
		fmt.Println(" => apakah dari group = ", v.Info.IsGroup)
		fmt.Println(" => apakah dari broadcast = ", v.Info.IsIncomingBroadcast())

		if !v.Info.IsFromMe &&
			v.Info.MessageSource.Chat.Server == "lid" &&
			!v.Info.IsGroup &&
			!v.Info.IsIncomingBroadcast() {

			fmt.Println("PENGIRIM = ", v.Info.Sender.User)

			if v.Message.GetConversation() != "" {
				pesan = v.Message.GetConversation()
			} else if v.Message.ExtendedTextMessage != nil && v.Message.ExtendedTextMessage.GetText() != "" {
				pesan = v.Message.ExtendedTextMessage.GetText()
			}

			fmt.Println("PESAN = " + pesan)

			var id_pesan []string
			id_pesan = append(id_pesan, v.Info.ID)
			_ = id_pesan

			clientWa.MarkRead(context.Background(), []string{v.Info.ID}, time.Now(), v.Info.Chat, v.Info.Sender)
			clientWa.SubscribePresence(context.Background(), v.Info.Sender)
			clientWa.SendPresence(context.Background(), types.PresenceAvailable)
			time.Sleep(2 * time.Second)
			clientWa.SendChatPresence(context.Background(), v.Info.Sender, types.ChatPresenceComposing, types.ChatPresenceMediaText)
			time.Sleep(3 * time.Second)
			clientWa.SendChatPresence(context.Background(), v.Info.Sender, types.ChatPresencePaused, types.ChatPresenceMediaText)

			pesanAsli := pesan

			// convert ke huruf kecil semua
			pesan = strings.ToLower(pesan)

			if strings.HasPrefix(pesan, "[ai]") {

				pertanyaan := strings.TrimSpace(pesanAsli[4:])

				if pertanyaan != "" {

					jawabanAi := ai.TanyaAi(v.Info.Sender.User, pertanyaan)

					kirimPesanText(v.Info.Sender, jawabanAi)

				} else {

					kirimPesanText(
						v.Info.Sender,
						"Masukkan pertanyaan setelah prefiks [ai]. Contoh: [ai] Selamat pagi",
					)

				}

			} else if pesan == "tes" {

				kirimPesan(v.Info.Sender)

			} else {

				kirimPesanDatabase(v.Info.Sender, pesan)

			}
		}
	}
}

func kirimPesan(IDPenerima types.JID) {
	clientWa.SendMessage(
		context.Background(),
		IDPenerima,
		&waE2E.Message{
			Conversation: proto.String("[UJI COBA] \n PESAN OTOMATIS"),
		},
	)
}

func kirimPesanText(IDPenerima types.JID, text string) {
	clientWa.SendMessage(
		context.Background(),
		IDPenerima,
		&waE2E.Message{
			Conversation: proto.String(text),
		},
	)
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
	if deviceStore != nil {
		deviceStore.Platform = "macOS"
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	clientWa = client
	DB = db

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
