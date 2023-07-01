package reader

import (
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nikonor/quickgobot/worker"
	"github.com/nikonor/quickgobot/writer"
	log "github.com/sirupsen/logrus"
)

type Reader struct{}

func NewReader() *Reader {
	return &Reader{}
}

func (r Reader) Handler(idx int, wg *sync.WaitGroup, doneChan <-chan struct{}, updates <-chan tgbotapi.Update,
	wChan chan writer.Message, worker *worker.Worker) {
	defer wg.Done()

	ld := func(s string) {
		log.WithFields(log.Fields{
			"file": "reader/reader.go",
			"func": "Handler",
		}).Debug(s)
	}

	for {
		select {
		case <-doneChan:
			ld("done reader")
			return
		case u := <-updates:
			wChan <- writer.Message{
				ChatId:  u.Message.Chat.ID,
				Message: u.Message.Text,
			}
		}
	}
}
