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
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type Entry struct {
	Address string `json:"address"`
	Comment string `json:"comment"`
	Date    string `json:"date"`
}

// common code
func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func getQueueURL(svc *sqs.SQS) (*sqs.GetQueueUrlOutput, error) {

	result, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(os.Getenv("SQS_QUEUE_NAME")),
	})
	check(err)

	return result, nil
}

func Connect() (*sqs.SQS, *sqs.GetQueueUrlOutput) {
	err := godotenv.Load(".env")
	check(err)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create an SQS service client
	svc := sqs.New(sess)

	queueURL, err := getQueueURL(svc)
	check(err)

	return svc, queueURL
}
// end common code

func SendMsg(svc *sqs.SQS, queueURL *string, method string, entry *Entry) error {

	messageAttributes := map[string]*sqs.MessageAttributeValue{
		"Method": {
			DataType:    aws.String("String"),
			StringValue: aws.String(method),
		},
	}

	if method == "addItem" || method == "getItem" || method == "removeItem" {
		messageAttributes["Address"] = &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(entry.Address),
		}
	}

	body := " "

	if method == "addItem" && entry.Comment != "" {
		body = entry.Comment
	}

	_, err := svc.SendMessage(&sqs.SendMessageInput{
		MessageAttributes:      messageAttributes,
		MessageBody:            aws.String(body),
		MessageGroupId:         aws.String("groupA"),
		MessageDeduplicationId: aws.String(uuid.New().String()),
		QueueUrl:               queueURL,
	})
	check(err)

	return nil
}

func main() {
	svc, queueURL := Connect()

	f, err := os.Open("client/data/ethereum_addresses_darklist.json")
	check(err)

	defer f.Close()

	byteValue, _ := ioutil.ReadAll(f)

	var entries []Entry

	json.Unmarshal(byteValue, &entries)

	for i := 0; i < len(entries); i++ {
		randOption := rand.Intn(4)
		entry := entries[i]

		switch randOption {
		case 0:
			// addItem
			err = SendMsg(svc, queueURL.QueueUrl, "addItem", &entry)
			check(err)
		case 1:
			// removeItem (or invalid if already removed)
			if i > 0 {
				randIndex := rand.Intn(i)
				err = SendMsg(svc, queueURL.QueueUrl, "removeItem", &entries[randIndex])
				check(err)
			}
			i--
		case 2:
			// getItem (or invalid if already removed)
			if i > 0 {
				randIndex := rand.Intn(i)
				err = SendMsg(svc, queueURL.QueueUrl, "getItem", &entries[randIndex])
				check(err)
				i--
			}
		case 3:
			// getItems
			err = SendMsg(svc, queueURL.QueueUrl, "getItems", &entry)
			check(err)
			i--
		}
	}
}
