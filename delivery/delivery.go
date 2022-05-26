package delivery

import (
	"jekabot/models"
	"log"
	"math/rand"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type myDelivery struct {
	TextUs   models.TextMessageUsecases
	VoiceUs  models.VoiceMessageUsecases
	TaksaUs  models.TaksaUsecases
	YtUs     models.YoutubeUsecases
	Config   models.TelegramConfig
	Bot      *tgbotapi.BotAPI
}

const TAKSA_CAPTION = "Собака умная может и самоутилизироваться )\n😍😍😍😍"
const YT_LINK_CAPTION = "Взгляните на это видео:\n\n"

func NewDelivery(
	textUs models.TextMessageUsecases,
	voiceUs models.VoiceMessageUsecases,
	taksaUs models.TaksaUsecases,
	ytUs models.YoutubeUsecases,
	c models.TelegramConfig,
	bot *tgbotapi.BotAPI) *myDelivery {

	textMsgs, err := textUs.GetTextMessagesCount()
	voiceMsgs, err := voiceUs.GetVoiceMessagesCount()
	textMsgsStr := strconv.Itoa(int(textMsgs))
	voiceMsgsStr := strconv.Itoa(int(voiceMsgs))

	if err != nil {
		log.Printf("total count messages err: %v", err)
	}

	log.Printf("total text messages: %s   total voices: %s", textMsgsStr, voiceMsgsStr)

	return &myDelivery{
		TextUs:   textUs,
		VoiceUs:  voiceUs,
		TaksaUs:  taksaUs,
		YtUs:     ytUs,
		Config:   c,
		Bot:      bot,
	}

}

func (d *myDelivery) Router(update tgbotapi.Update) {
	chatId := update.FromChat().ID
	strChattId := strconv.Itoa(int(chatId))

	if strChattId != d.Config.SouceChatID {
		d.respRouter(update)
	} else {
		d.storeRouter(update)
	}

}

func (t *myDelivery) respRouter(update tgbotapi.Update) {

	textMsg := update.Message.Text
	author := update.Message.From.UserName

	log.Printf("[%s] %s \n", author, textMsg)

	if strings.Contains(strings.ToLower(textMsg), "jeka_taksa") {
		t.RespondWithTaksa(update)
		return
	}

	if strings.Contains(strings.ToLower(textMsg), "jeka_video") {
		t.RespondWithYtUrl(update)
		return
	}

	isReply := update.Message.ReplyToMessage

	var isReplyToBot bool

	if isReply != nil {
		replyTo := update.Message.ReplyToMessage.From.UserName
		isReplyToBot = replyTo == t.Config.BotSign
	}

	isTriggerWords := strings.Contains(strings.ToLower(textMsg), "jeka")
	isAuthorJeka := author == t.Config.JekaRealid
	trigger := isTriggerWords || isAuthorJeka || isReplyToBot

	//make rundomize for text messages more
	//get rid of this piece of shit
	var fns []func(update tgbotapi.Update)
	fns = append(fns, t.RespondWithText)
	fns = append(fns, t.RespondWithText)
	fns = append(fns, t.RespondWithText)
	fns = append(fns, t.RespondWithText)
	fns = append(fns, t.RespondWithText)
	fns = append(fns, t.RespondWithText)
	fns = append(fns, t.RespondWithVoice)

	randFunc := fns[rand.Intn(len(fns))]

	if trigger {
		randFunc(update)
	}
}

func (d *myDelivery) storeRouter(update tgbotapi.Update) {
	if update.Message.Voice != nil {
		voiceId := update.Message.Voice.FileID
		d.VoiceUs.AddVoiceId(voiceId)
	} else {
		d.TextUs.AddTextMessage(update.Message.Text)
	}
}

func (d *myDelivery) RespondWithTaksa(update tgbotapi.Update) {

	bytes, id, err := d.TaksaUs.GetRandomTaksa()
	if err != nil {
		log.Printf("rand taksa error: %v", err)
	}

	msg := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FileBytes{Name: id, Bytes: bytes})
	msg.ReplyToMessageID = update.Message.MessageID
	msg.Caption = TAKSA_CAPTION
	d.Bot.Send(msg)
}

func (d *myDelivery) RespondWithYtUrl(update tgbotapi.Update) {
	ytUrl, err := d.YtUs.GetRandomVideoUrl()
	if err != nil {
		log.Printf("yt url error: %v", err)
	}

	msgText := YT_LINK_CAPTION + ytUrl

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
	msg.ReplyToMessageID = update.Message.MessageID
	d.Bot.Send(msg)
}

func (d *myDelivery) RespondWithText(update tgbotapi.Update) {
	randMsg, err := d.TextUs.GetRandTextMessage()
	if err != nil {
		log.Printf("rand text error: %v", err)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, randMsg)
	msg.ReplyToMessageID = update.Message.MessageID
	d.Bot.Send(msg)
}

func (d *myDelivery) RespondWithVoice(update tgbotapi.Update) {
	voiceId, err := d.VoiceUs.GetRandVoiceMessage()
	if err != nil {
		log.Printf("rand voice error: %v", err)
	}
	voice := tgbotapi.NewVoice(update.Message.Chat.ID, tgbotapi.FileID(voiceId))
	voice.ReplyToMessageID = update.Message.MessageID
	d.Bot.Send(voice)
}
