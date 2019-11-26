module avardstock

require (
	github.com/jinzhu/gorm v1.9.11
	// github.com/jinzhu/gorm v1.9.11
	github.com/qasemt/helper v0.0.0
)

replace github.com/qasemt/helper => ./../helper/

go 1.13
