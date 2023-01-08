package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/joho/godotenv"
)

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
	if err != nil {
		return nil, err
	}

	return result, nil
}

func connect() (*sqs.SQS, *sqs.GetQueueUrlOutput) {
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

type Node struct {
	prev *Node
	val  string
	next *Node
}

var cache = make(map[string]*Node)
var head *Node
var tail *Node
var mutex = &sync.RWMutex{}

func addNode(f *os.File, key string, value string) {
	mutex.Lock()
	_, ok := cache[key]
	if ok {
		mutex.Unlock()
		f.WriteString("Node already exists\n")
		return
	}
	if head == nil {
		head = &Node{nil, key, nil}
		tail = head
	} else {
		tail.next = &Node{tail, key, nil}
		tail = tail.next
	}
	cache[key] = tail
	mutex.Unlock()
	f.WriteString("Node added\n")
}

func removeNode(f *os.File, key string) {
	mutex.Lock()
	node, ok := cache[key]

	if !ok {
		mutex.Unlock()
		f.WriteString("Not found\n")
		return
	}

	if node.prev != nil {
		node.prev.next = node.next
	} else {
		head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		tail = node.prev
	}
	delete(cache, key)
	mutex.Unlock()
	f.WriteString("Node deleted\n")
}

func getNode(f *os.File, key string) {
	mutex.RLock()
	node, ok := cache[key]
	if !ok {
		mutex.RUnlock()
		f.WriteString("Not found\n")
	} else {
		f.WriteString(node.val + "\n")
		mutex.RUnlock()
	}
}

func getNodes(f *os.File) {
	resString := "GetItems: "
	mutex.RLock()
	node := head
	for node != nil {
		resString += node.val + ", "
		node = node.next
	}

	resString += "\n"
	mutex.RUnlock()

	f.WriteString(resString)
}

func worker(id int, f *os.File, jobs <-chan sqs.Message, results chan<- int) {
	for m := range jobs {
		// fmt.Println("worker", id, "started  job", m)
		fmt.Println(len(cache))
		method, ok := m.MessageAttributes["Method"]
		if !ok {
			f.WriteString("No method provided\n")
		}
		address, ok := m.MessageAttributes["Address"]

		switch *method.StringValue {
		case "getItem":
			if ok {
				getNode(f, *address.StringValue)
			} else {
				f.WriteString("No address provided\n")
			}
		case "getItems":
			getNodes(f)
		case "addItem":
			comment := m.Body
			if ok {
				addNode(f, *address.StringValue, *comment)
			} else {
				f.WriteString("No address provided\n")
			}
		case "removeItem":
			if ok {
				removeNode(f, *address.StringValue)
			} else {
				f.WriteString("No address provided\n")
			}
		}

		// fmt.Println("worker", id, "finished job", m)
		results <- id
	}
}
func processMessages(f *os.File, messages *sqs.ReceiveMessageOutput, mutex *sync.RWMutex, cache map[string]*Node, head *Node, tail *Node) {
	numJobs := len(messages.Messages)
	jobs := make(chan sqs.Message, numJobs)
	results := make(chan int, numJobs)

	for w := 1; w <= 10; w++ {
		go worker(w, f, jobs, results)
	}

	for j := 0; j < numJobs; j++ {
		jobs <- *messages.Messages[j]
	}
	close(jobs)

	for a := 1; a <= numJobs; a++ {
		<-results
	}
}

func deleteMessages(svc *sqs.SQS, queueURL *string, messages *sqs.ReceiveMessageOutput) {
	var deleteInput []*sqs.DeleteMessageBatchRequestEntry
	for _, message := range messages.Messages {
		deleteEntryInput := &sqs.DeleteMessageBatchRequestEntry{
			Id:            message.MessageId,
			ReceiptHandle: message.ReceiptHandle,
		}
		deleteInput = append(deleteInput, deleteEntryInput)
	}

	params := &sqs.DeleteMessageBatchInput{
		QueueUrl: queueURL,
		Entries:  deleteInput,
	}

	output, err := svc.DeleteMessageBatch(params)
	check(err)
	for _, failure := range output.Failed {
		log.Panic(failure)
	}
}

func retrieveMessages(svc *sqs.SQS, queueURL *sqs.GetQueueUrlOutput) *sqs.ReceiveMessageOutput {
	msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            queueURL.QueueUrl,
		MaxNumberOfMessages: aws.Int64(10),
		VisibilityTimeout:   aws.Int64(20),
		WaitTimeSeconds:     aws.Int64(20),
	})
	check(err)
	return msgResult
}

func main() {
	err := godotenv.Load(".env")
	check(err)

	t := time.Now().Format("2006-01-02-15:04:05")
	f, err := os.Create("./server/output/" + t + ".txt")
	check(err)
	defer f.Close()

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sqs.New(sess)

	queueURL, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(os.Getenv("SQS_QUEUE_NAME")),
	})

	// infinite loop polling for and dispatching messages
	for {
		msgResult := retrieveMessages(svc, queueURL)

		processMessages(f, msgResult, mutex, cache, head, tail)

		deleteMessages(svc, queueURL.QueueUrl, msgResult)
	}
}
