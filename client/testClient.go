package main

import (
	"fmt"
	"log"
	"os"

	"git.coding.net/bobxuyang/jadepool-seed/utils"
	"github.com/levigross/grequests"
)

func main() {
	if os.Args[1] == "seed" {
		resp, err := grequests.Get("http://127.0.0.1:8899/seed", nil)

		if err != nil {
			log.Fatalln("Unable to make request: ", err)
		}

		result := utils.KeyDecrypt(utils.CommKey, resp.String())

		fmt.Printf("Got reuslt: \"%s\"", result)
	} else {
		resp, err := grequests.Get("http://127.0.0.1:8899/data/"+os.Args[1], nil)

		if err != nil {
			log.Fatalln("Unable to make request: ", err)
		}

		result := utils.KeyDecrypt(utils.CommKey, resp.String())

		fmt.Printf("Got reuslt: \"%s\"", result)
	}

}
