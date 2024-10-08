package botapi

import (
	"context"

	telegramBot "github.com/go-telegram/bot"
	telegramBotModels "github.com/go-telegram/bot/models"
)

func authorizationMiddleware(b *Bot, handler telegramBot.HandlerFunc, minRole UserRole) telegramBot.HandlerFunc {
	return func(ctx context.Context, bot *telegramBot.Bot, update *telegramBotModels.Update) {
		userRole := b.getUserRole(update.Message.From.ID)
		if userRole >= minRole {
			handler(ctx, bot, update)
		} else {
			bot.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "You are not authorized to use this command.",
			})
		}
	}
}
