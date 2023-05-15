package main

import (
	"testing"

	"app/config"
	"app/test"
	"github.com/matryer/is"
	"github.com/playwright-community/playwright-go"
)

func TestAll(t *testing.T) {
	is, s := is.New(t), NewServer(config.CustomConfig(config.DevPlaywrightConfigPath))

	s.Start()
	defer s.Stop()

	pw := test.Must(playwright.Run())
	defer pw.Stop()

	browser := test.ChromeBrowser(pw, false)
	defer browser.Close()

	page := test.Must(browser.NewPage())
	_, err := page.Goto("localhost:" + s.port)
	is.NoErr(err)

	err = page.Click("#playlist_list > div:nth-child(4) > a")
	is.NoErr(err)

	playerbarVisible, err := page.IsVisible("#playerbar")
	is.NoErr(err)
	is.True(playerbarVisible)

	playButton, err := page.Locator("#playerbar > div > div:nth-child(2) > button")
	is.NoErr(err)

	err = playButton.Click()
	is.NoErr(err)

	// TODO check audio is playing
	// TODO check playerbar persists across page navigations
	// TODO fuzzy check that there is at least 4 buttons in the playlist
	// TODO check on clicking a song dots menu that the modal appears
	// TODO disable go test caching
}
