// Суть задачи:
// 1.	Подключиться используя доменную учетную запись по ssh  к коммутатору (12 коммутаторов)
// 2.	Выполнить команду на коммутаторе copy running-config tftp://10.11.9.2/N+1/имя_коммутатора или copy running-config tftp://10.11.9.2/data/имя_коммутатора

// Пожелания:
// Должна быть возможность запуска скрипта в планировщике.
// Список IP коммутаторов и логин пароль могут меняться(один из вариантов скрипт будет брать эту информацию из текстового файла или должна быть возможность менять скрипт)

// encrypt file and decrypt it after With flags -enc

// 10.41.1.3, 10.49.0.11 kbn Qwerty2019
// tftp://10.11.9.2/data/10.49.0.11/kie-dc1-swdc-03-running-config vrf management

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	// go install "golang.org/x/crypto/ssh"
	// go get "golang.org/x/crypto/ssh"
	"bytes"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

var (
	// user     = flag.String("u", "1", "User name")
	// password = flag.String("p", "1", "Password")
	// host     = flag.String("h", "1", "Host")
	// port     = flag.Int("pt", 22, "Port")
	jsFile = flag.String("f", "1", "File in same dir")
	// toEnc    = flag.String("e", "1", "To encrypt in same dir")
)

// Commutators
type Commutators []struct {
	Name     string `json:"Name"`
	Port     int    `json:"Port"`
	User     string `json:"User"`
	Password string `json:"Password"`
}

func getCommutators(jsFile string) (Commutators, error) {
	var CommFile Commutators
	f, err := os.Open(jsFile)
	if err != nil {
		log.Fatal("File open ", err)
		return CommFile, err
	}
	byteVal, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal("File read ", err)
		return CommFile, err
	}
	byteVal = bytes.TrimPrefix(byteVal, []byte("\xef\xbb\xbf"))
	err = json.Unmarshal(byteVal, &CommFile)
	if err != nil {
		log.Fatal("Unmarshaling error: ", err)
		return CommFile, err
	}

	defer f.Close()

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
		log.Fatal("Device not available: ", err)
		return "", err
	}

	session, err := client.NewSession()
	if err != nil {
		// panic(err)
		log.Fatal("Device not available: ", err)
		return "", err
	}
	defer session.Close()

	outputString := "copy running-config tftp://10.11.9.2/data/" + host + "/" + host + "-running-config  vrf management"
	commandResult, err := session.Output(outputString)
	if err != nil {
		println(outputString)
		log.Fatal("illegable output: ", err)
		return "", err
	}

	return string(commandResult), nil
}

func main() {
	flag.Parse()

	hosts, err := getCommutators(*jsFile)
	if err != nil {
		log.Fatal("unable get commutators: ", err)
		return
	}
	for _, commutator := range(hosts){
		out, err := invokeCmdSSH(commutator.Name, commutator.Port, commutator.User, commutator.Password)
		if err != nil {
			log.Fatal(commutator.Name, err)
		}
		log.Println(out)
	}
}
