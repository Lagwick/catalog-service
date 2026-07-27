package section

type (
	Processor struct {
		WebServer ProcessorWebServer `split_words:"true"`
		Grpc      ProcessorGrpc      `split_words:"true"`
		Gateway   ProcessorGateway   `split_words:"true"`
	}

	ProcessorWebServer struct {
		ListenPort uint32 `default:"8080" split_words:"true"`
	}

	ProcessorGrpc struct {
		ListenPort uint32 `split_words:"true" default:"50051"`
	}

	ProcessorGateway struct {
		ListenPort uint32 `split_words:"true" default:"8081"`
	}
)
