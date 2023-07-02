package worker

import (
	"sync"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nikonor/wlcmbot/writer"
	log "github.com/sirupsen/logrus"
)

type MainData struct {
	ChatId   int64
	NewUsers []tgbotapi.User
}

type Worker struct {
	workDir  string
	ch       chan *MainData
	doneChan chan struct{}
	writer   *writer.Writer
	wg       *sync.WaitGroup
}

func New(wg *sync.WaitGroup, workDir string, doneChan chan struct{}, writer *writer.Writer) *Worker {
	w := Worker{
		wg:       wg,
		workDir:  workDir,
		ch:       make(chan *MainData),
		doneChan: doneChan,
		writer:   writer,
	}

	go w.do()

	return &w
}

func (w Worker) Chan() chan *MainData {
	return w.ch
}

func (w Worker) do() {
	for {
		select {
		case <-w.doneChan:
			w.wg.Done()
			log.Debug("done worker")
			return
		case u := <-w.ch:
			if u == nil {
				continue
			}
			switch {
			case len(u.NewUsers) > 0:
				w.writer.Chan() <- writer.Message{
					ChatId:  u.ChatId,
					Message: "new user=" + u.NewUsers[0].UserName,
				}
			}
		}
	}
}
