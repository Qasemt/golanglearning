package avardstock
import (
   "github.com/jinzhu/gorm"
   _ "github.com/jinzhu/gorm/dialects/sqlite"
)


type Nemad struct {
   ID	string
   GroupCode   string
	GroupName	string
   NemadCode string
   NameEn string
  NameFa string
  NameFull string
}

func DatabaseInit() error {
   
   db, err := gorm.Open("sqlite3", "./stock.db")
  if err != nil {
    panic("failed to connect database")
  }
  defer db.Close()

  // Migrate the schema
  db.AutoMigrate(&Nemad{})

  
    return nil
}