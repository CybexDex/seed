package main

/*
#include <string.h>
*/
import "C"

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"unsafe"
	"time"
	"strings"

	"git.coding.net/bobxuyang/jadepool-seed/utils"
)

var server string

func init() {
	ip, err1 := ioutil.ReadFile("server.data")

	if err1 != nil {
		server = "http://127.0.0.1:8899/"
	} else {
		server = strings.Trim(string(ip), "\r\n")
	}
}

//export GetSeed
func GetSeed(up unsafe.Pointer) int {
	timeout := time.Duration(1 * time.Second)
	client := http.Client {
			Timeout: timeout,
	}

	resp, err := client.Get(server + "seed")

	if err != nil {
		fmt.Println(err.Error())
		return -1
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err.Error())
		return -1
	}

	seedX := string(body[:])
	seed := utils.KeyDecrypt(utils.CommKey, seedX)
	
	cup := (*C.char)(up)
	source := C.CString(seed)
	C.strcpy(cup, source)

	return 0
}

//export GetData
func GetData(up unsafe.Pointer, lengthp *int, t string) int {
	timeout := time.Duration(1 * time.Second)
	client := http.Client {
			Timeout: timeout,
	}

	resp, err := client.Get(server + "data/" + t)

	if err != nil {
		fmt.Println(err.Error())
		return -1
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err.Error())
		return -1
	}
	defer func() {
		if err := recover(); err != nil {
			data := ""
			cup := (*C.char)(up)
			source := C.CString(data)
			C.strcpy(cup, source)
			*lengthp = len(data)
		}
	}()
	dataX := string(body[:])
	data := utils.KeyDecrypt(utils.CommKey, dataX)
	
	cup := (*C.char)(up)
	source := C.CString(data)
	C.strcpy(cup, source)

	*lengthp = len(data)

	return 0
}

func main() {}
