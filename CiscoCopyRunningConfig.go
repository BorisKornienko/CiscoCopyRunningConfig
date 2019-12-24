package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"

	// go install "golang.org/x/crypto/ssh"
	// go get "golang.org/x/crypto/ssh"
	"bytes"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

var (
	jsFile = flag.String("f", "1", "File in same dir")
	toEnc  = flag.String("e", "1", "To encrypt in same dir")
)

// Commutators
type Commutators []struct {
	Name     string `json:"Name"`
	Port     int    `json:"Port"`
	User     string `json:"User"`
	Password string `json:"Password"`
}

var passphrase string = "XXXYYYYZZZ"

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, err := aes.NewCipher([]byte(createHash(passphrase)))

	if err != nil {
		log.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatal(err)
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Fatal(err)
	}
	return plaintext
}

func encryptFile(filename string, data []byte, passphrase string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	f.Write(encrypt(data, passphrase))
}

func decryptFile(filename string, passphrase string) ([]byte, error) {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return decrypt(f, passphrase), nil
}

func getCommutators(jsFile string) (Commutators, error) {
	var CommFile Commutators
	byteVal, err := decryptFile(jsFile, passphrase)
	if err != nil {
		log.Fatal("decrypt File read error: ", err)
		return CommFile, err
	}
	byteVal = bytes.TrimPrefix(byteVal, []byte("\xef\xbb\xbf"))
	err = json.Unmarshal(byteVal, &CommFile)
	if err != nil {
		log.Fatal("Unmarshaling error: ", err)
		return CommFile, err
	}

	return CommFile, nil
}

func invokeCmdSSH(host string, port int, user string, password string) (string, error) {
	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		// error 1 Device not allow
		log.Fatal(host+" Device not available: ", err)
		return "", err
	}

	session, err := client.NewSession()
	if err != nil {
		// panic(err)
		log.Fatal(host+" Device not available: ", err)
		return "", err
	}
	defer session.Close()

	outputString := "copy running-config tftp://10.11.9.2/data/" + host + "/" + host + " vrf management"
	commandResult, err := session.Output(outputString)
	if err != nil {
		println(outputString)
		log.Fatal(host+" illegable output: ", err)
		return "", err
	}

	return string(commandResult), nil
}

func main() {
	flag.Parse()
	if *toEnc != "1" {
		encryptedName := "enc_" + *toEnc
		forEnc, err := ioutil.ReadFile(*toEnc)
		if err != nil {
			log.Fatal("read file for encrypt: ", err)
		}
		encryptFile(encryptedName, forEnc, passphrase)
		return
	}
	hosts, err := getCommutators(*jsFile)
	if err != nil {
		log.Fatal("unable get commutators: ", err)
		return
	}
	for _, commutator := range hosts {
		out, err := invokeCmdSSH(commutator.Name, commutator.Port, commutator.User, commutator.Password)
		if err != nil {
			log.Fatal(commutator.Name, err)
		}
		log.Println(commutator.Name, out)
	}
}
