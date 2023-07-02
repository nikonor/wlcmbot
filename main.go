package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nikonor/wlcmbot/conf"
	"github.com/nikonor/wlcmbot/reader"
	"github.com/nikonor/wlcmbot/worker"
	"github.com/nikonor/wlcmbot/writer"

	log "github.com/sirupsen/logrus"
)

type ReaderI interface {
	Handler(idx int, wg *sync.WaitGroup, doneChan <-chan struct{}, updates <-chan tgbotapi.Update,
		wChan chan writer.Message, worker *worker.Worker)
}

type WriterI interface {
	Handler(idx int, wg *sync.WaitGroup, doneChan <-chan struct{}, tbot *tgbotapi.BotAPI,
		worker *worker.Worker)
}

func main() {
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	token, ok := os.LookupEnv("TLG_TOKEN")
	if !ok {
		panic("wrong token")
	}

	// TODO: конфиг из параметра
	cfg, err := conf.Load("./config.json")
	if err != nil {
		panic(err.Error())
	}

	tbot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err.Error())
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// TODO: webhook

	whInfo, err := tbot.GetWebhookInfo()
	if err != nil {
		panic(err.Error())
	}

	if whInfo.IsSet() {
		if _, err = tbot.Send(tgbotapi.DeleteWebhookConfig{DropPendingUpdates: true}); err != nil {
			panic(err.Error())
		}
	}
	doneChan := make(chan struct{})
	updates := tbot.GetUpdatesChan(u)
	wg := new(sync.WaitGroup)

	r := reader.NewReader()
	w := writer.NewWriter()
	wg.Add(1)
	ww, err := worker.New(wg, cfg, doneChan, w)
	if err != nil {
		panic(err.Error())
	}

	wg.Add(1)
	go sig(doneChan, wg)

	wg.Add(1)
	go w.Handler(1, wg, doneChan, tbot)
	wg.Add(1)
	go r.Handler(1, wg, doneChan, updates, ww)

	wg.Wait()
	time.Sleep(time.Second)
}

func sig(doneChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	for {
		s := <-sigChan
		switch s {
		case syscall.SIGINT, syscall.SIGTERM:
			close(doneChan)
			return
		}
	}
}
