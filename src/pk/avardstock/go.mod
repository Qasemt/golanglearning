module avardstock

require (
	github.com/jinzhu/gorm v1.9.11
	// github.com/jinzhu/gorm v1.9.11
	github.com/qasemt/helper v0.0.0
	golang.org/x/sync v0.0.0-20190227155943-e225da77a7e6
)

replace github.com/qasemt/helper => ./../helper/

go 1.13
