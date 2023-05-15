package test

import "github.com/playwright-community/playwright-go"

func ChromeBrowser(pw *playwright.Playwright, headless bool) playwright.Browser {
	var slowmo float64
	if !headless {
		slowmo = 500
	}

	browser := Must(pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		SlowMo:   playwright.Float(slowmo),
	}))

	return browser
}

func SafariBrowser(pw *playwright.Playwright, headless bool) playwright.Browser {
	var slowmo float64
	if !headless {
		slowmo = 500
	}

	browser := Must(pw.WebKit.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		SlowMo:   playwright.Float(slowmo),
	}))

	return browser
}
