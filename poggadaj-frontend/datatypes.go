package main

// HtmlClient represents an entry in the website's client list
type HtmlClient struct {
	Name                 string
	Description          map[string]string
	ImageUrl             string
	InstallerDownloadUrl *string
	ExtractedDownloadUrl *string
}
