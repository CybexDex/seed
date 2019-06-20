package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alexflint/go-arg"

	"git.coding.net/bobxuyang/jadepool-seed/utils"
)

var seed string
var password string
var dataJSONObj map[string]string
var configurJSONObj map[string]string
var clientIP string
var notInit = false

var args struct {
	Path     string
	Port     string
	Verify   bool   `arg:"-v" help:"verify seed data"`
	Data     bool   `arg:"-d" help:"add data into seed"`
	Config   bool   `arg:"-c" help:"add config into seed"`
	List     bool   `arg:"-l" help:"list add data & config"`
	Password bool   `arg:"-p" help:"change password"`
	Import   string `help:"import the data from file, format should be: (a:b\\n)*"`
}

var clientIPs []string

func writeMap(path string, data map[string]string) {
	tmpMap := make(map[string]string)

	for k, v := range data {
		rawv := utils.KeyDecrypt(password, v)
		tmpMap[k] = rawv
	}

	rawDataM, _ := json.Marshal(tmpMap)
	dataM, _ := json.Marshal(data)
	writeBytes(path, rawDataM, dataM)
}

func writeBytes(path string, rawData []byte, encryptedData []byte) {
	ioutil.WriteFile(args.Path+path, encryptedData, 0600)

	m := md5.New()
	m.Write(rawData)
	s := hex.EncodeToString(m.Sum(nil)[:16])
	ioutil.WriteFile(args.Path+"."+path+".md5", []byte(s), 0600)
}

func verify(path string, jsonMap map[string]string) {
	tmpMap := make(map[string]string)

	for k, v := range jsonMap {
		rawv := utils.KeyDecrypt(password, v)
		tmpMap[k] = rawv
	}

	data, _ := json.Marshal(tmpMap)
	m := md5.New()
	m.Write(data)
	s := hex.EncodeToString(m.Sum(nil)[:16])

	filemd5, _ := ioutil.ReadFile(args.Path + "." + path + ".md5")
	if s != string(filemd5) {
		panic("verify error!!! " + path + " have been changed by somebody else!")
	}

	fmt.Printf("%s md5[%s] pass check\n", path, s)
}

func main() {
	args.Path = "./"

	arg.MustParse(&args)
	fmt.Printf("data path: \"%s\".\n", args.Path)

	if _, err := os.Stat(args.Path + "seed.json"); os.IsNotExist(err) {
		notInit = true
	}

	// read data json
	data, err := ioutil.ReadFile(args.Path + "data.json")
	if err != nil {
		data = []byte(`{}`)
		writeMap("data.json", map[string]string{})
	}

	if err := json.Unmarshal(data, &dataJSONObj); err != nil {
		panic(err)
	}

	// read config json
	configur, err := ioutil.ReadFile(args.Path + "config.json")
	if err != nil {
		configur = []byte(`{}`)
		writeMap("config.json", map[string]string{})
	}

	if err := json.Unmarshal(configur, &configurJSONObj); err != nil {
		panic(err)
	}

	if notInit { // not INIT
		fmt.Println("The seed has not been initialized, program will halt right-now.")
		os.Exit(0)
	} else {
		var seedJSONObj map[string]string

		js, err := ioutil.ReadFile(args.Path + "seed.json")
		if err != nil {
			panic(err)
		}

		if err := json.Unmarshal(js, &seedJSONObj); err != nil {
			panic(err)
		}

		password = utils.ConfirmInput("passowrd", true)

		// compare the hash of password
		hashO := seedJSONObj["hash"]
		hashOut := utils.KeyDecrypt(password, hashO)
		h := sha256.New()
		h.Write([]byte(password))
		hashIn := hex.EncodeToString(h.Sum(nil))
		if hashIn != hashOut && hashIn != hashO {
			fmt.Println("Password is incorrect, Bye-Bye!")
			return
		}

		fmt.Println("Password is correct.")

		// decrpyted the seed with hash
		seedX := seedJSONObj["seed"]
		seed = utils.KeyDecrypt(password, seedX)

		// todo: for TEST only
		// fmt.Println("Seed: ", seed)

		fmt.Println("You are upgrade the SEED database now.")
		oldPassword := password
		newPassword := password

		// 备份seed.json => seed.last.json,data.json=>data.last.json,config.json=>config.last.json
		seedM, _ := json.Marshal(seedJSONObj)
		ioutil.WriteFile(args.Path+"seed.last.json", seedM, 0600)
		fmt.Println("Old seed.json move to seed.last.json")

		dataM, _ := json.Marshal(dataJSONObj)
		ioutil.WriteFile(args.Path+"data.last.json", dataM, 0600)
		fmt.Println("Old data.json move to data.last.json")

		configM, _ := json.Marshal(configurJSONObj)
		ioutil.WriteFile(args.Path+"config.last.json", configM, 0600)
		fmt.Println("Old config.json move to config.last.json")

		// hash和seed都要改变并保存
		h = sha256.New()
		h.Write([]byte(newPassword))
		hashnew := hex.EncodeToString(h.Sum(nil))
		hashX := utils.KeyEncrypt(newPassword, hashnew)
		seedJSONObj["hash"] = hashX

		seedX = utils.KeyEncrypt(newPassword, seed)
		seedJSONObj["seed"] = seedX

		// seedN, _ := json.Marshal(seedJSONObj)
		// ioutil.WriteFile(args.Path + "seed.json", seedN, 0600)

		writeMap("seed.json", seedJSONObj)

		//data.json和config.json中的数据全部要重新加密
		for v := range dataJSONObj {
			rawdata := utils.KeyDecrypt(oldPassword, dataJSONObj[v])
			newDatax := utils.KeyEncrypt(newPassword, rawdata)
			dataJSONObj[v] = newDatax
		}

		writeMap("data.json", dataJSONObj)

		for v := range configurJSONObj {
			rawdata := utils.KeyDecrypt(oldPassword, configurJSONObj[v])
			newDatax := utils.KeyEncrypt(newPassword, rawdata)
			configurJSONObj[v] = newDatax
		}

		writeMap("config.json", configurJSONObj)

		// verify
		verify("config.json", configurJSONObj)
		verify("data.json", dataJSONObj)
		verify("seed.json", seedJSONObj)

		fmt.Println("Seed database upgrade finished. Please start the server again.")
	}
}
