// Суть задачи:
// 1.	Подключиться используя доменную учетную запись по ssh  к коммутатору (12 коммутаторов)
// 2.	Выполнить команду на коммутаторе copy running-config tftp://10.11.9.2/N+1/имя_коммутатора или copy running-config tftp://10.11.9.2/data/имя_коммутатора

// Пожелания:
// Должна быть возможность запуска скрипта в планировщике.
// Список IP коммутаторов и логин пароль могут меняться(один из вариантов скрипт будет брать эту информацию из текстового файла или должна быть возможность менять скрипт)

// encrypt file and decrypt it after With flags -enc

// 10.41.1.3, 10.49.0.11 kbn Qwerty2019 

package main

import (
	"flag"
	"fmt"
	"log"

	// go install "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh"
)

var (
	user     = flag.String("u", "1", "User name")
	password = flag.String("p", "1", "Password")
	host     = flag.String("h", "1", "Host")
	port     = flag.Int("pt", 22, "Port")
)

func main() {
	flag.Parse()

	config := &ssh.ClientConfig{
		User:            *user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(*password),
		},
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		// error 1 Device not allow
		log.Fatal("Device not available: ", err)
		return
	}

	session, err := client.NewSession()
	if err != nil {
		// panic(err)
		log.Fatal("Device not available: ", err)
		return
	}
	defer session.Close()

	outputString := "copy running-config tftp://10.11.9.2/data/" + *host
	commandResult, err := session.Output(outputString)
	if err != nil {
		println(outputString)
		log.Fatal("illegable output: ", err)
		return
	}
	println(commandResult)
}
