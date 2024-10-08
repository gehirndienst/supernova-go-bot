package botapi

import (
	"context"
	"net/http"
	"os"
	"strconv"

	telegramBot "github.com/go-telegram/bot"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/gehirndienst/supernova-go-bot/internal/database"
	"github.com/gehirndienst/supernova-go-bot/internal/fetch"
)

type BotTokensConfig struct {
	TelegramAPIKey    string
	OpenAIAPIKey      string
	AccuWeatherAPIKey string
}

type BotWebhookConfig struct {
	URL  string
	Port string
}

type Bot struct {
	bot           *telegramBot.Bot
	tokensConfig  *BotTokensConfig
	webhookConfig *BotWebhookConfig
	adminID       int64
	handlers      map[string]string
	fetchers      map[string]fetch.Fetchable
	logger        *zerolog.Logger
	db            *database.Database
}

func InitBot(envfile string) (*Bot, error) {
	err := godotenv.Load(envfile)

	// can be instantiated without env params
	logger := GetLogger()

	if err != nil {
		logger.Fatal().Err(err).Msg("error loading .env file")
		return nil, err
	}

	if os.Getenv("TELEGRAM_API_KEY") == "" {
		logger.Fatal().Msg("error TELEGRAM_API_KEY is not set")
		return nil, errors.New("TELEGRAM_API_KEY is not set")
	}

	tokensConfig := &BotTokensConfig{
		TelegramAPIKey:    os.Getenv("TELEGRAM_API_KEY"),
		OpenAIAPIKey:      os.Getenv("OPEN_AI_API_KEY"),
		AccuWeatherAPIKey: os.Getenv("ACCU_WEATHER_API_KEY"),
	}

	adminID, err := strconv.ParseInt(os.Getenv("ADMIN_ID"), 10, 64)
	if err != nil {
		logger.Fatal().Err(err).Msg("error parsing ADMIN_ID")
		return nil, err
	}

	var webhookCfg *BotWebhookConfig
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL != "" {
		webhookCfg = &BotWebhookConfig{
			URL:  webhookURL,
			Port: os.Getenv("WEBHOOK_PORT"),
		}
	}

	tOpts := []telegramBot.Option{
		telegramBot.WithDefaultHandler(defaultHandler),
	}

	tBot, err := telegramBot.New(tokensConfig.TelegramAPIKey, tOpts...)
	if err != nil {
		logger.Fatal().Err(err).Msg("error creating telegram bot")
		return nil, err
	}

	db, err := database.NewDatabase()
	if err != nil {
		logger.Fatal().Err(err).Msg("error initializing database")
		return nil, err
	}

	bot := &Bot{
		bot:           tBot,
		tokensConfig:  tokensConfig,
		webhookConfig: webhookCfg,
		adminID:       adminID,
		handlers:      make(map[string]string),
		fetchers:      make(map[string]fetch.Fetchable),
		logger:        &logger,
		db:            db,
	}

	if err := bot.setFetchers(); err != nil {
		logger.Fatal().Err(err).Msg("error setting fetchers")
		return nil, err
	}

	bot.setHandlers()

	return bot, nil
}

func (b *Bot) Run(ctx context.Context) {
	if b.webhookConfig != nil {
		if err := b.runWebhook(ctx); err == nil {
			return
		}
		b.logger.Warn().Msg("falling back to long polling")
	}
	b.logger.Info().Msg("running telegram bot with long polling")
	b.bot.Start(ctx)
}

func (b *Bot) runWebhook(ctx context.Context) error {
	if _, err := b.bot.SetWebhook(ctx, &telegramBot.SetWebhookParams{
		URL: b.webhookConfig.URL,
	}); err != nil {
		b.logger.Fatal().Err(err).Msg("error setting webhook")
		return err
	}

	port := b.webhookConfig.Port
	if port == "" {
		port = "2000"
	}

	go func() {
		if err := http.ListenAndServe(":"+port, b.bot.WebhookHandler()); err != nil {
			b.logger.Fatal().Err(err).Msg("webhook server error")
		}
	}()

	b.logger.Info().Msg("running telegram bot with a webhook")
	b.bot.StartWebhook(ctx)

	return nil
}

func (b *Bot) setFetchers() error {
	// TODO: add more fetchers later
	if b.tokensConfig.AccuWeatherAPIKey != "" {
		weatherFetcher := &fetch.WeatherFetcher{}
		if err := weatherFetcher.Set(b.tokensConfig.AccuWeatherAPIKey, b.logger); err != nil {
			return err
		}
		b.fetchers["weather"] = weatherFetcher
	}

	if b.tokensConfig.OpenAIAPIKey != "" {
		chatFetcher := &fetch.ChatFetcher{}
		if err := chatFetcher.Set(b.tokensConfig.OpenAIAPIKey, b.logger); err != nil {
			return err
		}
		b.fetchers["chat"] = chatFetcher
	}
	return nil
}

func (b *Bot) setHandlers() {
	b.handlers["help"] = b.bot.RegisterHandler(telegramBot.HandlerTypeMessageText, "/help", telegramBot.MatchTypeExact, authorizationMiddleware(b, helpHandler, RegularUser))
	b.handlers["weather"] = b.bot.RegisterHandler(telegramBot.HandlerTypeMessageText, "/weather", telegramBot.MatchTypePrefix, authorizationMiddleware(b, weatherHandlerClosure(b), RegularUser))
	b.handlers["chat"] = b.bot.RegisterHandler(telegramBot.HandlerTypeMessageText, "/chat", telegramBot.MatchTypePrefix, authorizationMiddleware(b, chatHandlerClosure(b), RegularUser))
	b.handlers["getid"] = b.bot.RegisterHandler(telegramBot.HandlerTypeMessageText, "/getid", telegramBot.MatchTypeExact, authorizationMiddleware(b, getIDHandler, RegularUser))
	b.handlers["allow"] = b.bot.RegisterHandler(telegramBot.HandlerTypeMessageText, "/allow", telegramBot.MatchTypePrefix, authorizationMiddleware(b, allowHandlerClosure(b), AdminUser))
}

func (b *Bot) getUserRole(userID int64) UserRole {
	if userID == b.adminID {
		return AdminUser
	}
	if b.db != nil && b.db.IsUserAllowed(userID) {
		return PromotedUser
	}
	return RegularUser
}
