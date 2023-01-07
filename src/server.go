package main

import (
	"fmt"
	"net/http"
	"time"
	"log"
	"os"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/joho/godotenv"
)

func VerifyRouteAndMethod(w http.ResponseWriter, req *http.Request, route string, method string) bool {
	if req.URL.Path != route || req.Method != method {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return false
	}
	return true
}

func GetItem(w http.ResponseWriter, req *http.Request) {
	if !VerifyRouteAndMethod(w, req, "/getItem", "GET") {
		return
	}

	fmt.Fprintf(w, "Hello, %s!", req.URL.Path[1:])
}

func GetItems(w http.ResponseWriter, req *http.Request) {
	if !VerifyRouteAndMethod(w, req, "/getItems", "GET") {
		return
	}

	fmt.Fprintf(w, "Hello, %s!", req.URL.Path[1:])
}

func AddItem(w http.ResponseWriter, req *http.Request) {
	if !VerifyRouteAndMethod(w, req, "/addItem", "POST") {
		return
	}

	fmt.Fprintf(w, "Hello, %s!", req.URL.Path[1:])
}

func RemoveItem(w http.ResponseWriter, req *http.Request) {
	if !VerifyRouteAndMethod(w, req, "/removeItem", "DELETE") {
		return
	}

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		fmt.Println("worker", id, "started  job", j)
		time.Sleep(time.Second)
		fmt.Println("worker", id, "finished job", j)
		results <- j * 2
	}
}
func processMessages(messages *sqs.ReceiveMessageOutput) {
	for _, message := range messages.Messages {
		fmt.Println(*message.Body)
	}
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


	queueURL, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(os.Getenv("SQS_QUEUE_NAME")),
	})

	// infinite loop polling for and dispatching messages
	for {
		// get message from queue
		msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			AttributeNames: []*string{
				aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
			},
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			QueueUrl:            queueURL.QueueUrl,
			MaxNumberOfMessages: aws.Int64(20),
			VisibilityTimeout:   aws.Int64(20),
			WaitTimeSeconds:     aws.Int64(20),
		})
		
		// dispatch message
		// delete message from queue
	}
	// In order to use our pool of workers we need to send
	// them work and collect their results. We make 2
	// channels for this.
	const numJobs = 5
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	// This starts up 3 workers, initially blocked
	// because there are no jobs yet.
	for w := 1; w <= 3; w++ {
		go worker(w, jobs, results)
	}

	// Here we send 5 `jobs` and then `close` that
	// channel to indicate that's all the work we have.
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	// Finally we collect all the results of the work.
	// This also ensures that the worker goroutines have
	// finished. An alternative way to wait for multiple
	// goroutines is to use a [WaitGroup](waitgroups).
	for a := 1; a <= numJobs; a++ {
		<-results
	}
}
