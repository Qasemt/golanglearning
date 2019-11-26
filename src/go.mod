module golanglearning

require (
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/qasemt/avardstock v0.0.0
	github.com/qasemt/binance v0.0.0
	github.com/qasemt/helper v0.0.0
	github.com/qasemt/stockwork v0.0.0-00010101000000-000000000000
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.2.5 // indirect
)

replace github.com/qasemt/helper => ./pk/helper/

replace github.com/qasemt/stockwork => ./pk/stockwork/

replace github.com/qasemt/binance => ./pk/binance/

replace github.com/qasemt/avardstock => ./pk/avardstock

go 1.13
