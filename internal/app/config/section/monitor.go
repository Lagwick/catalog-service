package section

type Monitor struct {
	LogLevel    string `default:"debug" split_words:"true"`
	Environment string `default:"development"`
	Sentry      MonitorSentry
}

type MonitorSentry struct {
	Enabled bool   `default:"false"`
	DSN     string `default:""`
}
