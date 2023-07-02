package worker

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nikonor/wlcmbot/conf"
	"github.com/nikonor/wlcmbot/writer"
	log "github.com/sirupsen/logrus"
)

type MainData struct {
	ChatId   int64
	NewUsers []tgbotapi.User
}

type AdditionalDataCfg struct {
	NewUserMessage []byte
}

type Worker struct {
	cfg            *conf.Conf
	ch             chan *MainData
	doneChan       chan struct{}
	writer         *writer.Writer
	wg             *sync.WaitGroup
	additionalData AdditionalDataCfg
}

func New(wg *sync.WaitGroup, cfg *conf.Conf, doneChan chan struct{},
	writer *writer.Writer) (*Worker, error) {
	nUsrMsg, err := readFile(cfg.WorkDir, cfg.Files.NewUserTemplate)
	if err != nil {
		return nil, err
	}

	w := Worker{
		wg:       wg,
		cfg:      cfg,
		ch:       make(chan *MainData),
		doneChan: doneChan,
		writer:   writer,
		additionalData: AdditionalDataCfg{
			NewUserMessage: nUsrMsg,
		},
	}

	go w.do()

	return &w, nil
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
				var users []string
				for _, uName := range u.NewUsers {
					users = append(users, "@"+uName.UserName)
				}

				msgBody := bytes.ReplaceAll(w.additionalData.NewUserMessage, []byte("<!-- @@username -->"),
					[]byte(strings.Join(users, ", ")))

				w.writer.Chan() <- writer.Message{
					ChatId:  u.ChatId,
					Message: "new user message",
				}
				log.Debug("----" + string(msgBody) + "----")
				w.writer.Chan() <- writer.Message{
					ChatId:  u.ChatId,
					Message: string(msgBody),
				}
			}
		}
	}
}

func readFile(path, file string) ([]byte, error) {
	f, err := os.Open(path + file)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(f)
}
