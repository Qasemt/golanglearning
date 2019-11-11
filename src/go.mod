module golanglearning

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/qasemt/avardstock v0.0.0
	github.com/qasemt/binance v0.0.0
	github.com/qasemt/helper v0.0.0
	github.com/qasemt/stockwork v0.0.0-00010101000000-000000000000
	github.com/samuel/go-socks v0.0.0-20130725190102-f6c5f6a06ef6 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.2.5 // indirect
)

replace github.com/qasemt/helper => ./pk/helper/

replace github.com/qasemt/stockwork => ./pk/stockwork/

replace github.com/qasemt/binance => ./pk/binance/

replace github.com/qasemt/avardstock => ./pk/avardstock

go 1.13
