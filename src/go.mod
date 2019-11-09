module golanglearning

require (
	github.com/qasemt/binance v0.0.0
	github.com/qasemt/helper v0.0.0
	github.com/qasemt/stockwork v0.0.0-00010101000000-000000000000
	github.com/samuel/go-socks v0.0.0-20130725190102-f6c5f6a06ef6 // indirect
//h12.io/socks v1.0.0
)

replace github.com/qasemt/helper => ./pk/helper/

replace github.com/qasemt/stockwork => ./pk/stockwork/

//replace h12.io/socks => ./pk/h12.io/socks

//replace h12.io/socks => ./../pkg/mod/h12.io/socks@1.0.0
replace github.com/qasemt/binance => ./pk/binance/

//replace h12.io/socks => ./pk/h12.io/socks@v1.0.0

go 1.13
