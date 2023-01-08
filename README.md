# Local Setup
1. Go to AWS and create a FIFO SQS queue with the following settings
    - 30s visibility timeout
    - 0s delivery delay
    - 0s wait time
    - encryption disabled (for simplicity)
2. Copy the created SQS queue's name. Make a copy of `.env-template` and rename to `.env`. Add your queue name.'
3. Install AWS CLI with these [instructions](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html).
4. Create a shared local AWS configuration with these [instructions](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html).
3. Run `go run server/server.go` to start server.
4. In separate terminals, start as many clients as you'd like with `go run client/client.go`.
5. Server logs will be written to `./server/output/*`.
