package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"strconv"
	"time"
	"youtube-audio/pkg/reporter"
	"youtube-audio/pkg/util"
	"youtube-audio/pkg/util/env"
	"youtube-audio/pkg/util/log"
	"youtube-audio/pkg/util/myio"
	"youtube-audio/pkg/util/resource"
)

type Delivery struct {
	Parcel     Parcel
	PlaylistId string
	Done       bool
	Timestamp  int64
	Datetime   string
}

type Parcel struct {
	FilePath string
	Caption  string
	Url      string
}

type TelegramBot struct {
	//sync.Mutex
	Token         string //tg bot token, should be an admin in tg channel
	ChannelChatId string //tg channel's username only
	BotChatId     int64  //tg bot chat id
}

func SendAudio(delivery *Delivery) error {
	telegramBot, err := GenerateTelegramBot()
	if err != nil {
		return err
	}
	// Send an audio file
	err = telegramBot.Send(delivery.Parcel)
	if err == nil {
		markDelivered(delivery)
	}
	return err
}

func markDelivered(delivery *Delivery) {
	//only for testing
	//rand.Seed(time.Now().UnixNano())
	//delivery.Done = rand.Float32() < 0.5

	delivery.Done = true
	if delivery.Timestamp == 0 {
		now := time.Now()
		delivery.Timestamp = now.Unix()
		delivery.Datetime = now.Format(util.DateTimeFormat)
	}
}

func SendWarningMessage(template string, key ...any) {
	reporter.BriefSummary.FailedFetch++
	telegramBot, err := GenerateTelegramBot()
	if err != nil {
		log.Errorf("%s", err)
	}
	telegramBot.SendToBot(template, key...)
}

func SendSummary() {
	telegramBot, err := GenerateTelegramBot()
	if err != nil {
		log.Errorf("%s", err)
	}
	telegramBot.SendToBot(reporter.SummaryReportTemplate, reporter.BriefSummary.StartDatetime, reporter.BriefSummary.DurationSecs, reporter.BriefSummary.SuccessfulFetch, reporter.BriefSummary.FailedFetch)
}

func IsAudioValid(parcel Parcel) (bool, string) {
	if parcel.FilePath == "" {
		log.Warnf("file path EMPTY: %v", parcel)
		return false, util.EmptyFilePathWarningTemplate
	}
	// exists?
	audioExists, err := myio.FileExists(parcel.FilePath)
	if !audioExists {
		log.Warnf("downloaded file does NOT exist: %s, %v", parcel.FilePath, err)
		return false, util.FileNotExistWarningTemplate
	}
	// empty?
	audioFileInfo, _ := os.Stat(parcel.FilePath)
	log.Debugf("audioFileInfo size: %v", audioFileInfo.Size())
	if audioFileInfo.Size() < 1024 {
		log.Warnf("downloaded file size(%v) is not BIG enough(>= 1024B): %s", audioFileInfo.Size(), parcel.FilePath)
		return false, util.InvalidFileSizeWarningTemplate
	}
	return true, ""
}

func (t *TelegramBot) Send(parcel Parcel) error {
	//t.Lock()
	//defer t.Unlock()

	log.Infof("%s is going to be sent", parcel.FilePath)
	var err error

	bot, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		log.Errorf("new bot error, %s", err)
		return fmt.Errorf("building bot error")
	}

	log.Debugf("ready to new audio to channel")
	msg := tgbotapi.NewAudioToChannel(t.ChannelChatId, tgbotapi.FilePath(parcel.FilePath))
	msg.Caption = parcel.Caption
	log.Debugf("ready to send audio")

	_, err = bot.Send(msg)
	if err != nil {
		log.Errorf("bot send error, %s", err)
		return fmt.Errorf("sending audio error: %s", err)
	}
	log.Infof("audio %s has been sent", parcel.FilePath)

	return nil
}

func (t *TelegramBot) SendToBot(template string, key ...any) {
	//t.Lock()
	//defer t.Unlock()

	log.Warnf("Ready to send message about %v to telegram bot", key)
	var err error

	bot, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		log.Errorf("building msg bot error %s", err)
		return
	}

	msg := tgbotapi.NewMessage(t.BotChatId, fmt.Sprintf(template, key...))

	_, err = bot.Send(msg)
	if err != nil {
		log.Errorf("sending message error: %s", err)
		return
	}
	log.Infof("message %s has been sent", msg.Text)

}

func AppendDeliveries(deliveries *[]Delivery, fetchItems resource.FetchItems, playlistId string, done bool) {
	// remain time from FetchItems
	fetchTimestamp := fetchItems.Timestamp
	fetchDatetime := fetchItems.Datetime
	if done {
		// apply current time to last_fetch block
		now := time.Now()
		fetchTimestamp = now.Unix()
		fetchDatetime = now.Format(util.DateTimeFormat)
	} else {
		if len(fetchItems.Urls) > 0 {
			fetchTime := time.Unix(fetchTimestamp, 0)
			durationTillNow := time.Since(fetchTime)
			log.Debugf("playlistId: %s, including urls: %v at fetchDatetime: %s, fetch time till now hours: %v", playlistId, fetchItems.Urls, fetchDatetime, durationTillNow.Hours())
			if durationTillNow.Hours() > 48 {
				log.Warnf("fetch block time has EXPIRED: %s, playlistId: %s, urls: %v, drop it", fetchDatetime, playlistId, fetchItems.Urls)
				return
			}
		} else {
			log.Debugf("EMPTY fetch items urls, playlistId: %s, urls: %v, ignore it", playlistId, fetchItems.Urls)
			return
		}
	}
	// always keep the fetch block, but under maximum count of urls, drop random(?) ones
	for len(fetchItems.Urls) > util.FetchBlockMaxUrlsLimit {
		fetchItems.Urls = append(fetchItems.Urls[:0], fetchItems.Urls[1:]...)
	}
	for _, fetchUrl := range fetchItems.Urls {
		historyFetch := Delivery{
			Parcel:     GenerateParcel("", "", fetchUrl),
			PlaylistId: playlistId,
			Done:       done,
			Timestamp:  fetchTimestamp,
			Datetime:   fetchDatetime,
		}
		*deliveries = append(*deliveries, historyFetch)
	}
}

// RemoveDuplicatedUrlsByLoop 通过两重循环过滤重复元素 ref: https://blog.csdn.net/qq_27068845/article/details/77407358
func RemoveDuplicatedUrlsByLoop(slc []Delivery) []Delivery {
	var result []Delivery
	for i := range slc {
		flag := true
		for j := range result {
			if slc[i].Parcel.Url == result[j].Parcel.Url {
				flag = false // 存在重复元素，标识为false
				break
			}
		}
		if flag { // 标识为false，不添加进结果
			result = append(result, slc[i])
		}
	}
	return result
}

func GenerateParcel(filePath string, caption string, url string) Parcel {
	parcel := Parcel{
		FilePath: filePath,
		Caption:  caption,
		Url:      url,
	}
	return parcel
}

func GenerateTelegramBot() (TelegramBot, error) {
	var err error
	var telegramBot TelegramBot

	// Get the TOKEN and the CHAT_ID
	botToken, err := env.GetEnvVariable(util.EnvTokenName)
	if err != nil {
		log.Errorf("%s", err)
		return telegramBot, fmt.Errorf("reading env %s vars error", util.EnvTokenName)
	}
	telegramBot.Token = botToken

	channelChatId, err := env.GetEnvVariable(util.EnvChatIdName)
	if err != nil {
		log.Errorf("%s", err)
		return telegramBot, fmt.Errorf("reading env %s vars error", util.EnvChatIdName)
	}
	telegramBot.ChannelChatId = channelChatId

	botChatId, err := env.GetEnvVariable(util.EnvBotChatIdName)
	if err != nil {
		log.Errorf("%s", err)
		return telegramBot, fmt.Errorf("reading env %s vars error", util.EnvBotChatIdName)
	}
	intBotChatId, _ := strconv.ParseInt(botChatId, 10, 64)
	telegramBot.BotChatId = intBotChatId

	return telegramBot, nil
}
