package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/proxy"

	"log"
)

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database("kicker").Collection("games")
	fmt.Println("Hello")
	dialer, err := proxy.SOCKS5("tcp", "bla", &proxy.Auth{User: "bla", Password: "bla"}, proxy.Direct)
	if err != nil {
		log.Fatal(err)
	}

	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial
	bot, err := tgbotapi.NewBotAPIWithClient("bla", httpClient)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	fmt.Printf("new version Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 5

	updates, err := bot.GetUpdatesChan(u)
	for {
		select {
		case update := <-updates:
			go func() {
				if update.CallbackQuery != nil {
					if update.CallbackQuery.Data == "pavel" {
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()
						_, err := collection.InsertOne(ctx, bson.M{"date": time.Now(), "win": "pavel", "lose": "sasha", "second_score": "0"})
						if err != nil {
							log.Fatal(err)
						}
					}
					if update.CallbackQuery.Data == "sasha" {
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()
						_, err := collection.InsertOne(ctx, bson.M{"date": time.Now(), "win": "sasha", "lose": "pavel", "second_score": "0"})
						if err != nil {
							log.Fatal(err)
						}
					}
					config := tgbotapi.CallbackConfig{}
					config.CallbackQueryID = update.CallbackQuery.ID
					config.Text = "Done"
					time.Sleep(time.Duration(500) * time.Millisecond)
					mrk := getInlineKeyboard(ctx, collection)
					edit := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, mrk)
					if _, err := bot.Send(edit); err != nil {
						log.Fatal(err)
					}
					if _, err := bot.AnswerCallbackQuery(config); err != nil {
						log.Fatal(err)
					}
				}
				if update.Message == nil {
					return
				}
				if cmd := update.Message.CommandWithAt(); cmd != "" {
					switch cmd {
					case "start":
						btn := tgbotapi.NewKeyboardButton("/show")
						kb := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{btn})
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "starting...")
						msg.ReplyMarkup = kb
						if _, err := bot.Send(msg); err != nil {
							log.Fatal(err)
						}
					case "show":
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "tags:")
						mrk := getInlineKeyboard(ctx, collection)
						msg.ReplyMarkup = mrk
						_, err := bot.Send(msg)
						if err != nil {
							log.Fatal(err)
							return
						}
					}
				}
			}()
		}
	}
}

func getInlineKeyboard(ctx context.Context, collection *mongo.Collection) tgbotapi.InlineKeyboardMarkup {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pavel, err := collection.CountDocuments(ctx, bson.M{"win": "pavel", "lose": "sasha"})
	if err != nil {
		log.Fatal(err)
	}
	buttons := make([][]tgbotapi.InlineKeyboardButton, 2)
	btn := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Pavel Wins: %d", pavel), "pavel")
	row := tgbotapi.NewInlineKeyboardRow(btn)
	buttons[0] = row

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sasha, err := collection.CountDocuments(ctx, bson.M{"win": "sasha", "lose": "pavel"})
	if err != nil {
		log.Fatal(err)
	}
	btn = tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Sasha Wins: %d", sasha), "sasha")
	row = tgbotapi.NewInlineKeyboardRow(btn)
	buttons[1] = row
	mrk := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	return mrk
}
