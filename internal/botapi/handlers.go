package botapi

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	telegramBot "github.com/go-telegram/bot"
	telegramBotModels "github.com/go-telegram/bot/models"
)

// /////////////////////////////////////////////////////////////////////////////
// Raw handlers
// /////////////////////////////////////////////////////////////////////////////

func defaultHandler(ctx context.Context, b *telegramBot.Bot, update *telegramBotModels.Update) {
	b.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Type /help to get a list of available commands",
	})
}

func helpHandler(ctx context.Context, b *telegramBot.Bot, update *telegramBotModels.Update) {
	// TODO: update this message with the actual available commands
	b.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: "Available commands: " +
			"\n/help - get a list of available commands" +
			"\n/getid - get your user ID" +
			"\n/weather <city> <N> days|hours - get weather forecast for the city for N days or hours (PROMOTED USER)" +
			"\n/chat <prompt> - get a chatgpt response to the prompt (PROMOTED USER)",
	})
}

func getIDHandler(ctx context.Context, b *telegramBot.Bot, update *telegramBotModels.Update) {
	b.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Your ID is: %d", update.Message.From.ID),
	})
}

// /////////////////////////////////////////////////////////////////////////////
// Custom closures
// /////////////////////////////////////////////////////////////////////////////

func weatherHandlerClosure(b *Bot) telegramBot.HandlerFunc {
	return func(ctx context.Context, _ *telegramBot.Bot, update *telegramBotModels.Update) {
		wf := b.fetchers["weather"]
		if wf == nil {
			b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Weather fetcher is not available",
			})
			return
		}

		err := b.db.LogUserActivity(update.Message.From.ID, update.Message.Text)
		if err != nil {
			b.logger.Error().Err(err).Msg("Failed to log user activity")
		}

		// example: "/weather london (any case) 5 days" or "/weather london 12 hours"
		messageParts := strings.Fields(update.Message.Text)

		if len(messageParts) < 3 {
			b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Please provide a city and forecast type (days or hours). Example: /weather london 5 days",
			})
			return
		}

		city := messageParts[1]
		periodType := messageParts[3]

		qParams := map[string]interface{}{
			"city": city,
		}

		if strings.Contains(periodType, "day") {
			days, err := strconv.Atoi(messageParts[2])
			if err != nil || days < 1 {
				b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Invalid number of days. Please provide a valid number from 1 to 5 (more is truncated)",
				})
				return
			}
			qParams["days"] = days
		} else if strings.Contains(periodType, "hour") {
			hours, err := strconv.Atoi(messageParts[2])
			if err != nil || hours < 1 {
				b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Invalid number of hours. Please provide a valid number from 1 to 12 (more is truncated)",
				})
				return
			}
			qParams["hours"] = hours
		} else {
			b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Please specify either 'days' or 'hours'. Example: /weather london 3 days",
			})
			return
		}

		r, err := wf.Fetch(qParams)
		if err != nil {
			b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("Failed to fetch weather: %v", err),
			})
			return
		}

		b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   r,
		})
	}
}

func allowHandlerClosure(b *Bot) telegramBot.HandlerFunc {
	return func(ctx context.Context, _ *telegramBot.Bot, update *telegramBotModels.Update) {
		err := b.db.LogUserActivity(update.Message.From.ID, update.Message.Text)
		if err != nil {
			b.logger.Error().Err(err).Msg("Failed to log user activity")
		}

		if b.getUserRole(update.Message.From.ID) != AdminUser {
			b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "You are not authorized to use this command",
			})
			return
		}

		args := strings.Fields(update.Message.Text)
		if len(args) != 2 {
			b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Usage: /allow <user_id>",
			})
			return
		}

		userID, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Invalid user ID. Please provide a valid numeric ID",
			})
			return
		}

		err = b.db.AllowUser(userID)
		if err != nil {
			b.logger.Error().Err(err).Msg("Failed to allow user")
			b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Failed to allow user. Please try again later",
			})
			return
		}

		b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("User with ID %d has been allowed to use admin commands", userID),
		})
	}
}

func chatHandlerClosure(b *Bot) telegramBot.HandlerFunc {
	return func(ctx context.Context, _ *telegramBot.Bot, update *telegramBotModels.Update) {
		cf := b.fetchers["chat"]
		if cf == nil {
			b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Chat fetcher is not available",
			})
			return
		}

		err := b.db.LogUserActivity(update.Message.From.ID, update.Message.Text)
		if err != nil {
			b.logger.Error().Err(err).Msg("Failed to log user activity")
		}

		prompt := strings.TrimPrefix(update.Message.Text, "/chat ")
		if prompt == "" {
			b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Please provide a message after /chat command.",
			})
			return
		}

		response, err := cf.Fetch(map[string]interface{}{"prompt": prompt})
		if err != nil {
			b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("Error: %v", err),
			})
			return
		}

		b.bot.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   response,
		})
	}
}
