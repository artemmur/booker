package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type App struct {
	client *telegram.Client
	token  string
	raw    *tg.Client
	sender *message.Sender

	stateStorage *BoltState
	gaps         *updates.Manager
	dispatcher   tg.UpdateDispatcher

	db      *pebble.DB
	storage storage.MsgID
	mux     dispatch.MessageMux
	bot     *dispatch.Bot

	http   *http.Client
	logger *zap.Logger
}

func InitApp(logger *zap.Logger) (_ *App, rerr error) {
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		return nil, errors.Wrapf(err, "APP_ID not set or invalid %q", os.Getenv("APP_ID"))
	}

	appHash := os.Getenv("APP_HASH")
	if appHash == "" {
		return nil, errors.New("no APP_HASH provided")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		return nil, errors.New("no BOT_TOKEN provided")
	}

	// Setting up session storage.
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, "get home")
	}
	sessionDir := filepath.Join(home, ".crm")
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return nil, errors.Wrap(err, "mkdir")
	}

}

func setupBot(app *App) error {
	app.mux.HandleFunc("/bot", "Ping bot", func(ctx context.Context, e dispatch.MessageEvent) error {
		_, err := e.Reply().Text(ctx, "What?")
		return err
	})
	app.mux.Handle("/bill", "Create new bill")

	app.mux.Handle("/pp", "Pretty print replied message", inspect.Pretty())
	app.mux.Handle("/json", "Print JSON of replied message", inspect.JSON())
	app.mux.Handle("/stat", "Metrics and version", metrics.NewHandler(app.mts))
	app.mux.Handle("/tts", "Text to speech", tts.New(app.http))
	app.mux.Handle("/gpt2", "Complete text with GPT2",
		gpt.New(gentext.NewGPT2().WithClient(app.http)))
	app.mux.Handle("/gpt3", "Complete text with GPT3",
		gpt.New(gentext.NewGPT3().WithClient(app.http)))
	return nil
}
