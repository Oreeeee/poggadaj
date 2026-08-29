package main

// HtmlClient represents an entry in the website's client list
type HtmlClient struct {
	Name string

	// Name of the translation key for the description
	DescriptionI18nTag string
	ImageUrl           string
	DownloadUrl        string
}
