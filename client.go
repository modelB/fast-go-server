package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"
	"github.com/joho/godotenv"
)

func main() {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sqs.New(sess)

	urlResult, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: queue,
	})

	f, err := os.Open("names.csv")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		log.Println("Cannot read CSV file:", err)
	}

	for i, row := range rows {
		time.Sleep(4000000000)
		
		for j, domain := range row {
			if j > 0 && len(domain) > 3 {
				// fmt.Println(domain)
				ethName := domain + ".eth"
				resp, err := http.Get("https://etherscan.io/enslookup-search?search=" + ethName)
				if err != nil {
					log.Fatalln(err)
				}
				// //We Read the response body on the line below.
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatalln(err)
				}
				//Convert the body to type string
				sb := string(body)

				if i%50 == 0 {
					e := os.Remove("output.txt")
					if e != nil {
						log.Fatal(e)
					}
					f2, err3 := os.Create("output.txt")

					if err3 != nil {
						log.Fatal(err)
					}

					defer f2.Close()

					_, err2 := f2.WriteString(sb)

					if err2 != nil {
						log.Fatal(err2)
					}
				}

				if strings.Contains(sb, "is either not registered on ENS") {
					fmt.Println(domain)
					fmt.Println(ethName)
				} else if strings.Contains(sb, "security") {
					fmt.Println("SECURITY FAIL")
					// os.Exit(1)
				}
			}
		}

	}

	// if err := scanner.Err(); err != nil {
	// 	log.Fatal(err)
	// }
}
  