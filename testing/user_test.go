package testing

import (
	"testing"

	"app/config"
	"app/core"
	"github.com/go-rod/rod"
	"github.com/matryer/is"
)

func TestUser(t *testing.T) {
	is, s := is.New(t), core.NewServer(config.CustomConfig(config.PathConfigDevBrowser))

	s.Start()
	defer s.Stop()

	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("http://localhost:" + s.Port)

	wait := page.MustWaitNavigation()
	page.MustElement("#playlist_list > div:nth-child(4) > a").MustClick()
	wait()

	is.True(page.MustHas("#playerbar"))

	playerBarPlayButton := page.MustElement("#playerbar > div > div:nth-child(2) > button")
	playerBarPlayButton.MustClick()

	audioElement := page.MustElement("#audio-player")
	is.True(!audioElement.MustProperty("paused").Bool())

	playerBarTrackName := page.MustElement("#playerbar > div > div:nth-child(1) > div > div:nth-child(1) > a").MustText()

	// navigate back home via title /html/body/main/div[1]/a
	page.MustElement("body > main > div:nth-child(1) > a").MustClick()

	// check if playerbar has playerbarTrackName after navigating back home
	is.True(page.MustElement("#playerbar > div > div:nth-child(1) > div > div:nth-child(1) > a").MustText() == playerBarTrackName)

	// check audio is still playing after navigating back home
	is.True(!audioElement.MustProperty("paused").Bool())
}
