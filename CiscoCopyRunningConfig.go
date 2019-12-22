// Суть задачи:
// 1.	Подключиться используя доменную учетную запись по ssh  к коммутатору (12 коммутаторов)
// 2.	Выполнить команду на коммутаторе copy running-config tftp://10.11.9.2/N+1/имя_коммутатора или copy running-config tftp://10.11.9.2/data/имя_коммутатора

// Пожелания:
// Должна быть возможность запуска скрипта в планировщике.
// Список IP коммутаторов и логин пароль могут меняться(один из вариантов скрипт будет брать эту информацию из текстового файла или должна быть возможность менять скрипт) 

// encrypt file and decrypt it after With flags -enc


package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

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

	// type PrtgErr struct {
	// 	XMLName   xml.Name `xml:"prtg"`
	// 	PrtgError int      `xml:"error"`
	// 	ErrorText string   `xml:"text"`
	// }

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
		v := &PrtgErr{PrtgError: 1, ErrorText: "Device not available"}
		outputErr, err := xml.MarshalIndent(v, " ", " ")
		os.Stdout.Write(outputErr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		return
	}

	session, err := client.NewSession()
	if err != nil {
		// panic(err)
		v := &PrtgErr{PrtgError: 1, ErrorText: "Device not available"}
		outputErr, err := xml.MarshalIndent(v, " ", " ")
		os.Stdout.Write(outputErr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		return
	}
	defer session.Close()

	b, err := session.Output("show mac address-table count")
	if err != nil {
		v := &PrtgErr{PrtgError: 2, ErrorText: "Unknown output"}
		outputErr, err := xml.MarshalIndent(v, " ", " ")
		os.Stdout.Write(outputErr)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	// d := strings.Split(string(b), "\n") // "d" It's an array of output lines
	// // check for right output possibility
	// // if no ":" in output, so we can't parse with our template
	// if len(strings.Split(string(b), ":")) < 2 {
	// 	v := &PrtgErr{PrtgError: 2, ErrorText: "Unknown output"}
	// 	outputErr, err := xml.MarshalIndent(v, " ", " ")
	// 	os.Stdout.Write(outputErr)
	// 	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	// 	return
	// }

	// type Result struct {
	// 	XMLName xml.Name `xml:"result"`
	// 	Channel string   `xml:"channel"`
	// 	Value   int      `xml:"value"`
	// }

	// type Prtg struct {
	// 	XMLName xml.Name `xml:"prtg"`
	// 	Results []Result `xml:"result"`
	// }

	// create a high-level xml teg
	// v := &Prtg{}
	// // for each lines we check splitting by ":"
	// for _, line := range d {
	// 	// fmt.Println(i)
	// 	lineParse := strings.Split(line, ":")
	// 	if len(lineParse) == 2 { // if it contains name of parameter and it's value
	// 		cha := string(lineParse[0])
	// 		// fmt.Println(cha)
	// 		val := strings.TrimSpace(lineParse[1])
	// 		// fmt.Println(val)
	// 		if val != "" {
	// 			val, err := strconv.Atoi(val)
	// 			// I'm sorry for this
	// 			if err != nil {
	// 				// =======================TEMP======================
	// 				fmt.Fprintf(os.Stderr, "error: %v\n", err)
	// 				v := &PrtgErr{PrtgError: 22, ErrorText: "Unknown output"}
	// 				outputErr, err := xml.MarshalIndent(v, " ", " ")
	// 				os.Stdout.Write(outputErr)
	// 				if err != nil {
	// 					fmt.Fprintf(os.Stderr, "error: %v\n", err)
	// 				}
	// 				return
	// 			}
	// 			v.Results = append(v.Results, Result{Channel: cha, Value: val})
	// 		}
	// 	}
	// }
	// output, err := xml.MarshalIndent(v, " ", " ")
	// if err != nil {
	// 	fmt.Printf("error: %v\n", err)
	// }
	// os.Stdout.Write(output)
}