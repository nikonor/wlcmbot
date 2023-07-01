package writer

import (
	"errors"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nikonor/quickgobot/worker"
	log "github.com/sirupsen/logrus"
)

type Writer struct {
	ch chan Message
}

func NewWriter() *Writer {
	return &Writer{
		ch: make(chan Message),
	}
}

type Message struct {
	ChatId  int64
	Message string
}

func (w Writer) Handler(idx int, wg *sync.WaitGroup, doneChan <-chan struct{}, tbot *tgbotapi.BotAPI,
	worker *worker.Worker) {
	defer wg.Done()

	ld := func(s string) {
		log.WithFields(log.Fields{
			"file": "writer/writer.go",
			"func": "Handler",
		}).Debug(s)
	}
	le := func(s string) {
		log.WithFields(log.Fields{
			"file": "writer/writer.go",
			"func": "Handler",
		}).Error(s)
	}

	for {
		select {
		case <-doneChan:
			ld("done writer")
			return
		case msg := <-w.ch:
			m := tgbotapi.NewMessage(msg.ChatId, "<code>"+msg.Message+"</code>")
			m.ParseMode = tgbotapi.ModeHTML
			err := errors.New("tmp")
			for err != nil {
				ld("try to send")
				_, err = tbot.Send(m)
				if err == nil {
					break
				}
				for {
					select {
					case <-time.After(time.Second):
					case <-doneChan:
						le(err.Error())
						ld("done writer")
						return
					}
				}
			}
		}
	}
}

func (w Writer) Chan() chan Message {
	return w.ch
}
