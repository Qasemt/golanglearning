package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"
)

type LicenseGen struct {

}
const phrase = "abc===="
func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}
func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}
func (a LicenseGen) Make(email string) (*LicenseInfo,error) {
doc:=	LicenseInfo{
	End:time.Now().Add(time.Hour * 24 * 365), // 1 year
	Email:"qasemt@gmail.com",
	CpuId:"123123123123" ,
	BinPath:"c:/bin/" ,
}
	docBytes, err := json.Marshal(doc)
	if err != nil {
		log.Fatal(err)
	}
	sEnc := encrypt(docBytes,phrase)

	dec := decrypt(sEnc,phrase)

	fmt.Printf("Encrypted: %x\n", sEnc)
	fmt.Printf("Encrypted: %s\n", dec)

	var dat	=LicenseInfo {}

	if err := json.Unmarshal(dec, &dat); err != nil {
		panic(err)
	}
	fmt.Println(dat)
return nil,nil
}