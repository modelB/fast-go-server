package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/joho/godotenv"
)

type Entry struct {
	Address string `json:"address"`
	Comment string `json:"comment"`
	Date    string `json:"date"`
}

func GetQueueURL(svc *sqs.SQS) (*sqs.GetQueueUrlOutput, error) {
	
	result, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(os.Getenv("SQS_QUEUE_NAME")),
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func SendMsg(svc *sqs.SQS, queueURL *string, method string, body string) error {

	_, err := svc.SendMessage(&sqs.SendMessageInput{
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"Method": {
				DataType:    aws.String("String"),
				StringValue: aws.String(method),
			},
		},
		MessageBody: aws.String(body),
		QueueUrl:    queueURL,
	})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create an SQS service client
	svc := sqs.New(sess)


	queueURL, err := GetQueueURL(svc)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("data/ethereum_addresses_darklist.json")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)

	var entries []Entry

	json.Unmarshal(byteValue, &entries)

	for i := 0; i < len(entries); i++ {
		randOption := rand.Intn(8)

		log.Println(randOption)

		if randOption == 0 {
			// valid addItem
			stringifiedEntry, err := json.Marshal(entries[i])
			err = SendMsg(svc, queueURL.QueueUrl, "addItem", string(stringifiedEntry))
			if err != nil {
				log.Fatal(err)
			}
			log.Println(("addItem message sent to queue with date: " + entries[i].Date))
		} else if randOption == 1 {
			// invalid addItem
			err = SendMsg(svc, queueURL.QueueUrl, "addItem", "invalid")
			if err != nil {
				log.Fatal(err)
			}
			// didn't actually add so don't increment i
			i--
		} else if randOption == 2 {
			// valid removeItem (or invalid if already removed)
			randIndex := rand.Intn(i)
			randomAddress := entries[randIndex].Address
			err = SendMsg(svc, queueURL.QueueUrl, "removeItem", randomAddress)
			if err != nil {
				log.Fatal(err)
			}
			i--
		} else if randOption == 3 {
			err = SendMsg(svc, queueURL.QueueUrl, "removeItem", "invalid")
			i--
		} else if randOption == 4 {
			// valid getItem (or invalid if already removed)
			randIndex := rand.Intn(i)
			randomAddress := entries[randIndex].Address
			err = SendMsg(svc, queueURL.QueueUrl, "getItem", randomAddress)
			if err != nil {
				log.Fatal(err)
			}
			i--
		} else if randOption == 5 {
			// invalid getItem
			err = SendMsg(svc, queueURL.QueueUrl, "getItem", "invalid")
			if err != nil {
				log.Fatal(err)
			}
			i--
		} else if randOption == 6 {
			// valid getItems
			err = SendMsg(svc, queueURL.QueueUrl, "getItems", " ")
			if err != nil {
				log.Fatal(err)
			}
			i--
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}
