package front

import (
	"fmt"
	"strconv"

	"github.com/Gealber/calvopro-botv2/scrapper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

//formListMarkup form the markup reply to send in the search
func formListMarkup(videos []*scrapper.Video) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for i, video := range videos {
		title := func() string {
			if len(video.Title) > 30 {
				return video.Title[:30]
			}
			return video.Title
		}()
		text := fmt.Sprintf("🍆🍑%s... [%d MB]", title, video.Size >> 20)
		button := tgbotapi.NewInlineKeyboardButtonData(text, strconv.Itoa(i+1))
		row := tgbotapi.NewInlineKeyboardRow(button)
		rows = append(rows, row)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return keyboard
}

//createBtn create a button
func createBtn(text string, data string) tgbotapi.InlineKeyboardMarkup {
	btn := tgbotapi.NewInlineKeyboardButtonData(text, data)
	row := tgbotapi.NewInlineKeyboardRow(btn)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	return keyboard
}
