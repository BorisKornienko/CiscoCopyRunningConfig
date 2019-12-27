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
	"time"

	// go install "golang.org/x/crypto/ssh"
	// go get "golang.org/x/crypto/ssh"
	"bytes"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

var (
	jsFile = flag.String("f", "", "Encrypted file in the same dir")
	toEnc  = flag.String("e", "", "File to encrypt in the same dir")
	backup = flag.Bool("b", false, "Move to Backup folder")
	help   = flag.Bool("h", false, "Get help")
)

// Commutators
type Commutators []struct {
	Name     string `json:"Name"`
	Port     int    `json:"Port"`
	Ftp      string `json:"Ftp"`
	User     string `json:"User"`
	Password string `json:"Password"`
}

var passphrase string = "XXXYYYZZZ"

func getBackupFolder(rootFolder string) error {
	entities, err := ioutil.ReadDir(rootFolder)
	if err != nil {
		log.Fatal(err)
	}
	for _, entity := range entities {
		if entity.IsDir() {
			if entity.Name() == "Backup" {
				log.Println(entity.Name() + " folder exist")
				return nil
			}
		}
	}
	err = os.Mkdir("Backup", 0755)
	if err != nil {
		log.Fatal("Backup folder creation error: ", err)
		return err
	}
	log.Println("Backup folder created")
	return nil
}

func moveToBackupAndCreate(rootFolder string) error {
	entities, err := ioutil.ReadDir(rootFolder)
	todayDate := time.Now()
	if err != nil {
		log.Fatal(err)
	}
	for _, entity := range entities {
		if entity.IsDir() {
			if entity.Name() != "Backup" {
				fmt.Println(todayDate.Format("01-02-2006"))
				err := os.Rename(entity.Name(), "./Backup/"+todayDate.Format("01-02-2006")+"_"+entity.Name())
				if err != nil {
					log.Println(entity.Name()+" rename and move err: ", err)
					err := os.Rename(entity.Name(), "./Backup/"+todayDate.Format("01-02-2006")+"_"+entity.Name()+"-1")
					if err != nil {
						log.Fatal(entity.Name()+" rename and move err: ", err)
						return err
					}
				}
				os.Mkdir(entity.Name(), 0755)
				log.Println(entity.Name() + " renamed and moved to Backup")
			}
		}
	}
	return nil
}

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

func invokeCmdSSH(host string, port int, ftp string, user string, password string) (string, error) {
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

	// tftp://10.11.9.2/data/
	outputString := "copy running-config " + ftp + host + "/" + host + " vrf management"
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

	if *help {
		fmt.Println("json:Name,Port,Ftp,User,Password; json:Ftp such as tftp://10.11.9.2/data/ ; file must be uncrypted only; the depth of backup unlim, control it; no run more then twice per day or not use flag -b")
		return
	}
	// File encryption only
	if *toEnc != "" {
		encryptedName := "enc_" + *toEnc
		forEnc, err := ioutil.ReadFile(*toEnc)
		if err != nil {
			log.Fatal("read file for encrypt: ", err)
		}
		encryptFile(encryptedName, forEnc, passphrase)
		return
	}

	//move to Backup folder
	if *backup {
		getBackupFolder("./")
		moveToBackupAndCreate("./")
	}

	// processed command in ssh
	if *jsFile != "" {
		hosts, err := getCommutators(*jsFile)
		if err != nil {
			log.Fatal("unable get commutators: ", err)
			return
		}
		for _, commutator := range hosts {
			if _, err := os.Stat(commutator.Name); os.IsNotExist(err) {
				err := os.Mkdir(commutator.Name, 0755)
				if err != nil {
					log.Println(commutator.Name+" dircreation error: ", err)
				}
			}
			out, err := invokeCmdSSH(commutator.Name, commutator.Port, commutator.Ftp, commutator.User, commutator.Password)
			if err != nil {
				log.Fatal(commutator.Name, err)
			}
			log.Println(commutator.Name, out)
		}
	}
}
