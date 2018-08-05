package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/translate"
	"golang.org/x/net/context"
	"golang.org/x/text/language"
	"gopkg.in/telegram-bot-api.v4"
)

type TransactionJSON struct {
	Result TransactionResultJSON
}

type TransactionResultJSON struct {
	Input string
}

func hex2int(hexStr string) int {
	result, _ := strconv.ParseInt(hexStr, 16, 64)
	return int(result)
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TGAPIKEY"))
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	baseurl := "https://api.etherscan.io/api?module=proxy&action=eth_getTransactionByHash&txhash="
	txhash := "0xb1ed364e4333aae1da4a901d5231244ba6a35f9421d4607f7cb90d60bf45578a"
	ethscanapitoken := os.Getenv("ETHSCANAPIKEY")
	url := baseurl + txhash + "&apikey=" + ethscanapitoken

	resp, err := http.Get(url)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var tx TransactionJSON
	err = json.Unmarshal(body, &tx)
	if err != nil {
		log.Panic(err)
	}
	inputFromTransaction := strings.Replace(tx.Result.Input, "0x", "", 1)
	decoded, err := hex.DecodeString(inputFromTransaction)
	if err != nil {
		log.Panic(err)
	}

	inputString := string(decoded)

	messages := strings.Split(inputString, "\n")
	var filtered []string
	for _, message := range messages {
		if strings.TrimSpace(message) != "" {
			filtered = append(filtered, strings.TrimSpace(message))
		}
	}

	messagesCount := len(filtered)

	ctx := context.Background()
	client, err := translate.NewClient(ctx)
	if err != nil {
		log.Panic(err)
	}
	target, err := language.Parse("pt")
	if err != nil {
		log.Panic(err)
	}

	var channelID int64 = -1001360104969
	for i, message := range filtered {
		translations, err := client.Translate(ctx, []string{message}, target,
			&translate.Options{
				Format: translate.Text,
			})
		if err != nil {
			log.Panic(err)
		}
		finalMessage := fmt.Sprintf("(%d/%d) %s", i+1, messagesCount, translations[0].Text)
		msg := tgbotapi.NewMessage(channelID, finalMessage)
		fmt.Println(finalMessage)
		bot.Send(msg)
		time.Sleep(time.Millisecond * 3000)
	}
}
