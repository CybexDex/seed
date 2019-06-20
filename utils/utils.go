package utils

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io/ioutil"
	mrand "math/rand"

	"github.com/bgentry/speakeasy"
)

//CommKey ...
const CommKey string = "0efee2b9b23f5337fd4b39621e55e0c12f6d7430bde05023db6d19d4b6853de0"

//ReadFile ...
func ReadFile(f string) (string, error) {
	content, err := ioutil.ReadFile(f)
	if err != nil {
		// fmt.Println(err.Error())
		return "", err
	}

	return fmt.Sprintf("%s", content), nil
}

//ConfirmInput ...
func ConfirmInput(s string, isPwd bool) string {
	// input password twice
	fmt.Printf("Please input %s.\n", s)

	var one string

	if isPwd {
		one = ReadPwdFromStdin()
	} else {
		one = ReadFromStdin()
	}

	return one
}

//ConfirmInput2 ...
func ConfirmInput2(s string, isPwd bool) string {
	// input password twice
	fmt.Printf("Input %s first time.\n", s)

	var one, two string

	if isPwd {
		one = ReadPwdFromStdin()
	} else {
		one = ReadFromStdin()
	}

	fmt.Printf("Input %s again.\n", s)

	if isPwd {
		two = ReadPwdFromStdin()
	} else {
		two = ReadFromStdin()
	}

	if one != two {
		fmt.Printf("Two %s isn't the same, Bye-Bye!.\n", s)
		os.Exit(1)
	}

	return one
}

//ReadFromStdin ...
func ReadFromStdin() string {
	fmt.Print("-> ")
	var input string
	fmt.Scanln(&input)
	return input
}

//ReadPwdFromStdin ...
func ReadPwdFromStdin() string {
	fmt.Print("-> ")

	password, err := speakeasy.Ask("")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return password
}

//GenRandom ...
func GenRandom(p string) string {
	t := time.Now().UnixNano()
	mrand.Seed(t)
	r := mrand.Intn(2<<62-1) + 1
	return strconv.Itoa(int(t)) + p + strconv.Itoa(r)
}

//Check ...
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

//KeyEncrypt ...
func KeyEncrypt(keyStr string, cryptoText string) string {
	keyBytes := sha256.Sum256([]byte(keyStr))
	return encrypt(keyBytes[:], cryptoText)
}

// encrypt string to base64 crypto using AES
func encrypt(key []byte, text string) string {
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext)
}

//KeyDecrypt ...
func KeyDecrypt(keyStr string, cryptoText string) string {
	keyBytes := sha256.Sum256([]byte(keyStr))
	return decrypt(keyBytes[:], cryptoText)
}

// decrypt from base64 to decrypted string
func decrypt(key []byte, cryptoText string) string {
	ciphertext, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	return fmt.Sprintf("%s", ciphertext)
}
