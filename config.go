package main

// Config is options for application
type Config struct {
	Name          string `default:"PatentFetcher"`
	Debug         bool   `default:"false"`
	Address       string `default:":6789"`
	PagesFilePath string `default:"files/pages.bs"`
	ResultDirPath string `default:"files/results/"`
}
