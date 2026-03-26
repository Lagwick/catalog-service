package section

type (
	Processor struct {
		WebServer ProcessorWebServer `split_words:"true"`
	}

	ProcessorWebServer struct {
		ListenPort uint32 `default:"8080" split_words:"true"`
	}
)
