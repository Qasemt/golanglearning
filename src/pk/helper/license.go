package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"
)

type LicenseInfo struct {
	Email   string `json:"email"`
	CpuId   string `json:"cpuid"`
	BinPath string `json:"binpath"`
}
type ActivatedInfo struct {
	Email         string    `json:"email"`
	CpuId         string    `json:"cpuid"`
	End           time.Time `json:"end"`   //time.Now().Add(time.Hour * 24 * 365), // 1 year
	Start         time.Time `json:"start"` //time.Now().Add(time.Hour * 24 * 365), // 1 year
	BinPath       string    `json:"binpath"`
	NumberOfItems int32       `json:"numberofitems"`
	CryptoEnable  bool      `json:"cryptoenable"`
	TehranEnable  bool      `json:"tehranenable"`
	ForexEnable   bool      `json:"forexenable"`
}

func (a ActivatedInfo) RemainingTime() int64 {
	diff := a.End.Sub(time.Now()).Hours() / 24
	return int64(RoundUp(diff, 0))
}

type LicenseGen struct {
}

const phrase = "123==="

var dir_path string = path.Join(GetRootCache(), "license")

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

func (a LicenseGen) MakeLicense(email string) error {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	re_email :=regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !re_email.MatchString(email) {
		return errors.New("Make License : Email Not Valid");
	}
	ma_id, err := machineid.ID()

	if err != nil {
		return err
	}

	doc := LicenseInfo{
		Email:   email,
		CpuId:   ma_id,
		BinPath: pwd,
	}
	docBytes, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	sEnc := encrypt(docBytes, phrase)
	if GetVerbose() {
		fmt.Printf("Encrypted: %x\n", sEnc)
	}
	//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: LICENSE WRITE TO FILE
	if _, err := os.Stat(dir_path); os.IsNotExist(err) {
		os.MkdirAll(dir_path, os.ModePerm)
	}
	var s string = path.Join(dir_path, "license.ini")
	file, err1 := os.Create(s)
	defer file.Close()

	if err1 != nil {
		return errors.New(fmt.Sprintf("make license -> Cannot create file %s", err1))
	}

	file.WriteString(fmt.Sprintf("%x", sEnc))

	file.Sync()
	//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: LICENSE WRITE TO FILE
	fmt.Println("license has been created succefully", s)
	return nil
}
func (a LicenseGen) MakeActivate(license_path string, days int32, items_num int32, is_cryto bool, is_tehran bool, is_forex bool) error {
	var li_path string = license_path
	if !IsExist(li_path) {
		return errors.New(fmt.Sprintf("MakeActivate -> could not find file  %s", li_path))
	}

	basedir := filepath.Dir(license_path)
	//:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: READ LICENSE

	content, err := ioutil.ReadFile(li_path)
	if err != nil {
		return errors.New(fmt.Sprintf("MakeActivate -> Cannot create file %s \n", err))
	}
	data, err := hex.DecodeString(string(content))
	if err != nil {
		return errors.New(fmt.Sprintf("MakeActivate -> decode failed  %s \n", err))
	}
	var license = LicenseInfo{}
	dec := decrypt(data, phrase)

	if err := json.Unmarshal(dec, &license); err != nil {
		return errors.New(fmt.Sprintf("MakeActivate -> json failed %s \n", err))
	}

	//:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: MAKE ACTIVATE
	doc := ActivatedInfo{
		End:           time.Now().Add(time.Hour * 24 * time.Duration(days)), // 1 year
		Start:         time.Now(),
		Email:         license.Email,
		CpuId:         license.CpuId,
		BinPath:       license.BinPath,
		NumberOfItems: items_num,
		CryptoEnable:  is_cryto,
		TehranEnable:  is_tehran,
		ForexEnable:   is_forex,
	}
	docBytes, err := json.Marshal(doc)
	if err != nil {
		return errors.New(fmt.Sprintf("MakeActivate -> json failed %s \n", err))
	}

	sEnc := encrypt(docBytes, phrase)
	//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: ACTIVATE WRITE TO FILE

	if _, err := os.Stat(basedir); os.IsNotExist(err) {
		os.MkdirAll(basedir, os.ModePerm)
	}
	var a_path string = path.Join(basedir, fmt.Sprintf("%v_%v", license.Email, "activated.ini"))
	file, err1 := os.Create(a_path)

	defer file.Close()

	if err1 != nil {
		return errors.New(fmt.Sprintf("make activate -> Cannot create file %s", err1))
	}

	file.WriteString(fmt.Sprintf("%x", sEnc))

	file.Sync()
	//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: LICENSE WRITE TO FILE
	fmt.Println("activate has been created succefully", a_path)
	return nil
	//:::::::::::::::::::::::::::::::::::::::::::::::::::::

}
func (a LicenseGen) ReadFile(file_name string, target_object_json interface{}) error {
	var li_path string = path.Join(dir_path, file_name)
	if !IsExist(li_path) {
		return errors.New(fmt.Sprintf("could not find file ->  %s", li_path))
	}
	content, err := ioutil.ReadFile(li_path)
	if err != nil {
		return errors.New(fmt.Sprintf("ReadFile()-> Cannot create file %s \n", err))
	}
	data, err := hex.DecodeString(string(content))
	if err != nil {
		return errors.New(fmt.Sprintf("ReadFile() -> decode failed  %s \n", err))
	}

	dec := decrypt(data, phrase)

	if err := json.Unmarshal(dec, &target_object_json); err != nil {
		return errors.New(fmt.Sprintf("MakeActivate -> json failed %s \n", err))
	}

	return nil
}
func (a LicenseGen) Validation() error {

	var license = LicenseInfo{}
	var activated = ActivatedInfo{}
	//:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: READ LICENSE
	e1 := a.ReadFile("license.ini", &license)
	if e1 != nil {
		return e1
	}

	e2 := a.ReadFile("activated.ini", &activated)
	if e2 != nil {
		return e2
	}

	if license.CpuId != activated.CpuId {
		return errors.New(fmt.Sprintf("Validation() -> CPU ID Conflict \n"))
	}
	if license.BinPath != activated.BinPath {
		return errors.New(fmt.Sprintf("Validation() -> Bin Path Conflict \n a:[%v] \n l:[%v]\n", activated.BinPath, license.BinPath))
	}

	if license.Email != activated.Email {
		return errors.New(fmt.Sprintf("Validation() -> Bin Path Conflict \n a:[%v] \n l:[%v]\n", activated.Email, license.Email))
	}

	if activated.RemainingTime() <= 0 {
		return errors.New(fmt.Sprintf("Validation() -> license Expire"))
	}
	return nil
}
func (a LicenseGen) Print() error {

	e := a.Validation()
	if e != nil {
		fmt.Println("license not valid")
		if GetVerbose() {
			fmt.Println("%v", e)
		}
	}
	var license = LicenseInfo{}
	var ai = ActivatedInfo{}
	//:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: READ LICENSE
	e1 := a.ReadFile("license.ini", &license)
	if e1 != nil {
		return e1
	}
	e2 := a.ReadFile("activated.ini", &ai)
	if e1 != nil {
		return e2
	}
	fmt.Println(":::::::::::::: LICENSE :::::::::::::::")
	fmt.Println("Email :", ai.Email, "\nStart :", TimeToString(ai.Start, ""), "\nEnd   :", TimeToString(ai.End, ""), "\nRemaining Time :", ai.RemainingTime(), " Days", "\nBin Path :", ai.BinPath, "\nItems Num :", ai.NumberOfItems, "\nTehran :", ai.TehranEnable, "\nCrypto :", ai.CryptoEnable, "\nForex :", ai.ForexEnable)
	return nil
}

func (a LicenseGen) Test() {
	//a.MakeLicense("qasemt@gmail.com")
	a.MakeActivate("D:\\workspace\\goprojects\\golanglearning\\src\\d\\license\\license.ini", 360, 10, true, true, false)

	e := a.Validation()
	if e != nil {
		fmt.Println("license not valid")
	}
	a.Print()
}
