package main

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/bwmarrin/discordgo"
)

func main() {
	a := app.NewWithID("botchat")
	w := a.NewWindow("Discord Bot Chat Helper")
	w.Resize(fyne.NewSize(640, 480))
	exePath, err := os.Executable()
	if err != nil {
		dialog.ShowError(err, w)
	}
	exeDir := filepath.Dir(exePath)
	tokenPath := filepath.Join(exeDir, "token.txt")
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		os.Create(tokenPath)
	}
	token := widget.NewEntry()
	token.SetPlaceHolder("Enter Bot Token Here")
	submitBtn := widget.NewButton("Submit", func() {
		chatPage(a, w, token.Text)
	})
	loadTokenBtn := widget.NewButton("Load From File", func() {
		tokenFile, err := os.ReadFile(tokenPath)
		if err != nil {
			dialog.ShowError(err, w)
		}
		token.SetText(string(tokenFile))
	})
	tokenPage := container.NewVBox(
		token,
		submitBtn,
		loadTokenBtn,
	)
	w.SetContent(tokenPage)
	w.ShowAndRun()
}

func chatPage(a fyne.App, w fyne.Window, token string) {
	var guildNames []string
	var guildIDs []string
	var channelNames []string
	var channelIDs []string
	var channel *discordgo.Channel
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	discord.Identify.Intents |= discordgo.IntentGuilds
	discord.Identify.Intents |= discordgo.IntentGuildMessages
	discord.Identify.Intents |= discordgo.IntentGuildMembers
	err = discord.Open()
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	guilds, err := discord.UserGuilds(100, "", "", false)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	for _, g := range guilds {
		guildNames = append(guildNames, g.Name)
		guildIDs = append(guildIDs, g.ID)
	}
	messagesBtn := widget.NewButton("Show Messages", func() {
		if channel == nil {
			return
		}
		showMessages(a, discord, channel.ID)
	})
	messagesBtn.Hide()
	channelList := widget.NewSelect(channelNames, func(s string) {
		selectedChannelID := ""
		for i, name := range channelNames {
			if name == s {
				selectedChannelID = channelIDs[i]
				break
			}
		}
		channel, _ = discord.Channel(selectedChannelID)
		messagesBtn.Show()
	})
	channelList.PlaceHolder = "Select a Channel"
	guildList := widget.NewSelect(guildNames, func(s string) {
		if s == "" {
			channelList.Hide()
		} else {
			selectedGuildID := ""
			for i, name := range guildNames {
				if name == s {
					selectedGuildID = guildIDs[i]
					break
				}
			}
			if selectedGuildID == "" {
				channelList.Hide()
				return
			}

			channels, err := discord.GuildChannels(selectedGuildID)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			channelIDs = nil
			channelNames = nil
			for _, ch := range channels {
				if ch.Type == discordgo.ChannelTypeGuildText {
					channelNames = append(channelNames, ch.Name)
					channelIDs = append(channelIDs, ch.ID)
				}
			}
			channelList.SetOptions(channelNames)
			channelList.SetSelected("")
			channelList.Show()
		}
	})
	guildList.PlaceHolder = "Select a Guild"
	channelList.Hide()
	messageEntry := widget.NewMultiLineEntry()
	messageEntry.SetPlaceHolder("Type your message here...")
	submitMessageBtn := widget.NewButton("Send Message", func() {
		if channel == nil {
			return
		}
		_, err := discord.ChannelMessageSend(channel.ID, messageEntry.Text)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		messageEntry.SetText("")
	})

	content := container.NewVBox(
		guildList,
		channelList,
		messageEntry,
		submitMessageBtn,
		messagesBtn,
	)
	w.SetContent(content)
}

func showMessages(a fyne.App,discord *discordgo.Session, channelID string) {
	w := a.NewWindow("Messages")
	w.Resize(fyne.NewSize(640, 480))
	messages, err := discord.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	var cards []fyne.CanvasObject
	for _, m := range messages {
		card := widget.NewCard(m.Author.Username, m.Content, nil)
		cards = append(cards, card)
	}
	content := container.NewVBox(cards...)
	w.SetContent(content)
	w.Show()
}
