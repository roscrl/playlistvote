//nolint:ireturn,nolintlint
package test

import "github.com/playwright-community/playwright-go"

func ChromeBrowser(play *playwright.Playwright, headless bool) playwright.Browser {
	var slowmo float64
	if !headless {
		slowmo = 500
	}

	browser := Must(play.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		SlowMo:   playwright.Float(slowmo),
	}))

	return browser
}

func SafariBrowser(play *playwright.Playwright, headless bool) playwright.Browser {
	var slowmo float64
	if !headless {
		slowmo = 500
	}

	browser := Must(play.WebKit.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		SlowMo:   playwright.Float(slowmo),
	}))

	return browser
}
