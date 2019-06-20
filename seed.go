package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/rjeczalik/notify"

	"git.coding.net/bobxuyang/jadepool-seed/utils"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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

func updateClientIPs() {
	cli := configurJSONObj["client"]
	if strings.Compare(cli, "") != 0 {
		clientIP = utils.KeyDecrypt(password, cli)
		clientIPs = strings.Split(clientIP, ",")
	} else {
		clientIPs = append(clientIPs, "127.0.0.1")
	}
}

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

func notifier(file string) {
	c := make(chan notify.EventInfo, 1)

	if err := notify.Watch(args.Path+file, c, notify.Write); err != nil {
		panic(err)
	}
	//defer notify.Stop(c)

	go func() {
		for {
			// Block until an event is received.
			switch ei := <-c; ei.Event() {
			case notify.Write:
				if notInit {
					continue
				}

				fmt.Printf("[Notify] Updated file %s.\n", ei.Path())

				data, err := ioutil.ReadFile(args.Path + file)
				if err != nil {
					panic(fmt.Sprintf("read %s file error in notifier", file))
				}

				jmap := make(map[string]string)
				err = json.Unmarshal(data, &jmap)
				if err != nil {
					panic(fmt.Sprintf("read %s file error in notifier", file))
				}

				if file == "data.json" {
					dataJSONObj = jmap
				} else if file == "config.json" {
					configurJSONObj = jmap
					updateClientIPs()
				}

				fmt.Printf("[Notify] Updated %s in memory.\n", file)
				// fmt.Printf("[Notify] [")
				// for v := range jmap {
				//     fmt.Printf("%s:\"%s\",", v, utils.KeyDecrypt(password, jmap[v]))
				// }
				// fmt.Printf("]\n")
			}
		}
	}()
}

func lockFile() *os.File {
	if _, err := os.Stat(".lock"); os.IsNotExist(err) {
		lockF, _ := os.OpenFile(".lock", os.O_CREATE, 0600)
		lockF.Close()
	}

	lockf, err := os.OpenFile(".lock", os.O_RDWR, os.ModeExclusive)
	if err != nil {
		fmt.Println(err)
		panic("lock file is owned by another process!")
	}

	c := make(chan error)
	go func() {
		c <- syscall.Flock(int(lockf.Fd()), syscall.LOCK_EX)
	}()

	select {
	case err := <-c:
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("lock file successfully")
		}
	case <-time.After(time.Second * 5):
		fmt.Println("lock file timeout, there is another seed process running. This process will halt right-now.")
		os.Exit(0)
	}

	// graceful terminate program
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGHUP)

	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v\n", sig)
		fmt.Println("Wait for 2 second to finish processing")

		if lockf != nil {
			fmt.Println("unlock the lock file")
			lockf.Close()
		}

		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()

	return lockf
}

func main() {
	args.Path = "./"
	args.Port = "8899"
	//args.Import = ""

	arg.MustParse(&args)
	fmt.Printf("data path: \"%s\", port: \"%s\".\n", args.Path, args.Port)

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

	notifier("config.json")
	notifier("data.json")

	if notInit { // not INIT
		lock := lockFile()
		defer lock.Close()

		fmt.Println("The seed has not been initialized, please type \"yes\" to continue.")
		input := utils.ReadFromStdin()

		if input != "yes" {
			fmt.Printf("Input %s, Bye-Bye!\n", input)
			return
		}

		// input password twice
		password = utils.ConfirmInput2("password", true)

		// password is correct, gen hash of the password
		h := sha256.New()
		h.Write([]byte(password))
		hash := hex.EncodeToString(h.Sum(nil))
		hashX := utils.KeyEncrypt(password, hash)

		// derive the seed with password, current time, and random number
		ran := utils.GenRandom(password)

		// gen seed
		h = sha256.New()
		h.Write([]byte(ran))
		seed = hex.EncodeToString(h.Sum(nil))

		// test use only, need to be deleted in prd code
		// fmt.Println("Seed: ", seed)

		// encrpyted the seed with hash
		seedX := utils.KeyEncrypt(password, seed)

		seedMap := make(map[string]string)
		seedMap["hash"] = hashX
		seedMap["seed"] = seedX

		// init the json file
		writeMap("seed.json", seedMap)

		fmt.Println("Initialization finished. Please start the server again.")
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
		if hashIn != hashOut {
			fmt.Println("Password is incorrect, Bye-Bye!")
			return
		}

		fmt.Println("Password is correct.")

		// verify
		verify("config.json", configurJSONObj)
		verify("data.json", dataJSONObj)
		verify("seed.json", seedJSONObj)

		// decrpyted the seed with hash
		seedX := seedJSONObj["seed"]
		seed = utils.KeyDecrypt(password, seedX)

		// todo: for TEST only
		// fmt.Println("Seed: ", seed)

		if args.Verify {
			fmt.Println("You are verifying the seed now.")
			fmt.Print("Seed:     ", seed[0:8])
			fmt.Print("................")
			fmt.Println(seed[len(seed)-8 : len(seed)])

			h := md5.New()
			h.Write([]byte(seed))
			seedMD5 := hex.EncodeToString(h.Sum(nil))

			fmt.Println("Seed MD5:", seedMD5)
			fmt.Println("Verification finished. Please start the server again.")
		} else if args.Data {
			fmt.Println("You are set the data now.")
			fmt.Println("Please input the name first.")
			name := utils.ReadFromStdin()

			dd := utils.ConfirmInput("the data", false)

			fmt.Printf("%s:%s\n", name, dd)
			for v := range dataJSONObj {
				kv := utils.KeyDecrypt(password, dataJSONObj[v])
				if kv == dd {
					fmt.Printf("Value already used %s: \"%s\",confirm by type 'yes' \n", v, kv)
					input := utils.ReadFromStdin()
					if input != "yes" {
						fmt.Printf("Input %s, Bye-Bye!\n", input)
						return
					}
				}
			}
			dataJSONObj[name] = utils.KeyEncrypt(password, dd)

			writeMap("data.json", dataJSONObj)
			fmt.Println("Setting data finished. Please start the server again.")
		} else if args.List {
			fmt.Println("The data list is:")

			for v := range dataJSONObj {
				fmt.Printf("%s: \"%s\"\n", v, utils.KeyDecrypt(password, dataJSONObj[v]))
			}

			fmt.Println("The config list is:")

			for v := range configurJSONObj {
				fmt.Printf("%s: \"%s\"\n", v, utils.KeyDecrypt(password, configurJSONObj[v]))
			}
		} else if args.Config {
			fmt.Println("You are set the config now.")
			fmt.Println("Please input the name first.")
			name := utils.ReadFromStdin()

			dd := utils.ConfirmInput("the config", false)

			fmt.Printf("%s:%s\n", name, dd)

			configurJSONObj[name] = utils.KeyEncrypt(password, dd)

			writeMap("config.json", configurJSONObj)
			fmt.Println("Setting config finished. Please start the server again.")
		} else if args.Password {
			lock := lockFile()
			defer lock.Close()

			fmt.Println("You are change the password now.")
			oldPassword := password
			newPassword := utils.ConfirmInput2("new password", true)
			password = newPassword

			// 备份seed.json => seed.last.json,data.json=>data.last.json,config.json=>config.last.json
			seedM, _ := json.Marshal(seedJSONObj)
			ioutil.WriteFile(args.Path+"seed.json.last", seedM, 0600)
			fmt.Println("seed.json move to seed.json.last")

			dataM, _ := json.Marshal(dataJSONObj)
			ioutil.WriteFile(args.Path+"data.json.last", dataM, 0600)
			fmt.Println("data.json move to data.json.last")

			configM, _ := json.Marshal(configurJSONObj)
			ioutil.WriteFile(args.Path+"config.json.last", configM, 0600)
			fmt.Println("config.json move to config.json.last")

			f, err := ioutil.ReadFile(args.Path + ".seed.json.md5")
			if err != nil {
				panic(err)
			}
			ioutil.WriteFile(args.Path+"seed.json.md5.last", f, 0600)
			fmt.Println(".seed.json.md5 move to seed.json.md5.last")

			f, err = ioutil.ReadFile(args.Path + ".data.json.md5")
			if err != nil {
				panic(err)
			}
			ioutil.WriteFile(args.Path+"data.json.md5.last", f, 0600)
			fmt.Println(".data.json.md5 move to data.json.md5.last")

			f, err = ioutil.ReadFile(args.Path + ".config.json.md5")
			if err != nil {
				panic(err)
			}
			ioutil.WriteFile(args.Path+"config.json.md5.last", f, 0600)
			fmt.Println(".config.json.md5 move to config.json.md5.last")

			// hash和seed都要改变并保存
			h := sha256.New()
			h.Write([]byte(newPassword))
			hashnew := hex.EncodeToString(h.Sum(nil))
			hashX := utils.KeyEncrypt(newPassword, hashnew)
			seedJSONObj["hash"] = hashX

			seedX := utils.KeyEncrypt(newPassword, seed)
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

			fmt.Println("Change password succeeded, Please start the server again.")
		} else if args.Import != "" {
			// parse file
			iData, err := ioutil.ReadFile(args.Import)
			if err != nil {
				panic(err)
			}

			iString := string(iData)
			s := strings.Split(iString, "\n")
			aDataNum := 0
			inDataNum := 0
			for l := range s {
				// fmt.Printf("line: %s \n",s[l])
				wArr := strings.Split(s[l], ":")
				if s[l] == "" {

				} else if len(wArr) != 2 {
					fmt.Printf("You have error in file line: %d \n", l+1)
					fmt.Printf("line: %s \n", s[l])
					return
				} else {
					if _, ok := dataJSONObj[wArr[0]]; ok {
						//do something here
						inDataNum = inDataNum + 1
					}
					aDataNum = aDataNum + 1
					dataJSONObj[wArr[0]] = utils.KeyEncrypt(password, wArr[1])
				}
			}
			fmt.Printf("You will import %d datas,overwrite %d datas,confirm by type 'yes' \n", aDataNum, inDataNum)
			input := utils.ReadFromStdin()

			if input != "yes" {
				fmt.Printf("Input %s, Bye-Bye!\n", input)
				return
			}

			writeMap("data.json", dataJSONObj)
			fmt.Println("import data finished. Please start the server again.")
		} else { // start the server
			lock := lockFile()
			defer lock.Close()

			fmt.Println("Successfully recover the seed, now starting the seed server :)")

			updateClientIPs()

			fmt.Printf("Authorized client is: %s.\n", clientIPs)

			// Echo instance
			e := echo.New()

			// Middleware
			e.Use(middleware.Logger())
			e.Use(middleware.Recover())

			// Routes
			e.GET("/seed", getSeed)
			e.GET("/data/:type", getData)

			// Start server
			e.Logger.Fatal(e.Start(":" + args.Port))
		}
	}
}

// Handler
func getSeed(c echo.Context) error {
	inIP := strings.Split(c.Request().RemoteAddr, ":")[0]
	sort.Strings(clientIPs)
	i := sort.SearchStrings(clientIPs, inIP)
	if i < len(clientIPs) && clientIPs[i] == inIP {

	} else {
		return c.String(http.StatusUnauthorized, "")
	}

	seedX := utils.KeyEncrypt(utils.CommKey, seed)

	return c.String(http.StatusOK, seedX)
}

// Handler
func getData(c echo.Context) error {
	inIP := strings.Split(c.Request().RemoteAddr, ":")[0]
	sort.Strings(clientIPs)
	i := sort.SearchStrings(clientIPs, inIP)
	if i < len(clientIPs) && clientIPs[i] == inIP {

	} else {
		return c.String(http.StatusUnauthorized, "")
	}
	v := dataJSONObj[c.Param("type")]
	d := ""
	if v != "" {
		d = utils.KeyDecrypt(password, v)
	}
	dd := utils.KeyEncrypt(utils.CommKey, d)

	return c.String(http.StatusOK, dd)
}
