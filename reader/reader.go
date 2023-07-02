package reader

import (
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nikonor/wlcmbot/worker"
	log "github.com/sirupsen/logrus"
)

const (
	NewUser = 1
)

type Reader struct{}

func NewReader() *Reader {
	return &Reader{}
}

func (r Reader) Handler(idx int, wg *sync.WaitGroup, doneChan <-chan struct{},
	updates <-chan tgbotapi.Update, worker *worker.Worker) {
	defer wg.Done()

	for {
		select {
		case <-doneChan:
			log.Debug("done reader")
			return
		case u := <-updates:
			typ, users := typeOfMessage(u)
			switch typ { // nolint:
			case NewUser:
				if len(users.NewUsers) == 0 {
					log.Error("has not new users")
					continue
				}
				worker.Chan() <- users
			}
		}
	}
}

func typeOfMessage(u tgbotapi.Update) (uint, *worker.MainData) {
	chatID := getChatId(u)

	if chatID == 0 {
		return 0, nil
	}

	if u.Message == nil || len(u.Message.NewChatMembers) == 0 {
		return 0, nil
	}

	return NewUser, &worker.MainData{
		ChatId:   chatID,
		NewUsers: u.Message.NewChatMembers,
	}
}

func getChatId(u tgbotapi.Update) int64 {
	if u.Message == nil || u.Message.Chat == nil {
		return 0
	}
	return u.Message.Chat.ID
}
