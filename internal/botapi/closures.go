package botapi

import (
	"context"

	telegramBot "github.com/go-telegram/bot"
	telegramBotModels "github.com/go-telegram/bot/models"
)

func TextHandlerClosure(text string) telegramBot.HandlerFunc {
	return func(ctx context.Context, b *telegramBot.Bot, update *telegramBotModels.Update) {
		b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      text,
			ParseMode: telegramBotModels.ParseModeMarkdown,
		})
	}
}
