module golanglearning

require (
	github.com/qasemt/helper v0.0.0
	github.com/qasemt/stockwork v0.0.0-00010101000000-000000000000
)

replace github.com/qasemt/helper => ./pk/helper/

replace github.com/qasemt/stockwork => ./pk/stockwork/

replace h12.io/socks => ./pk/h12.io/socks

go 1.13
