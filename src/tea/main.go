package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

var queueUrl string
var service string
var version string

type ApiGatewayEvent struct {
	HttpMethod string            `json:"httpMethod"`
	Path       string            `json:"path"`
	Headers    map[string]string `json:"headers"`
	Body       *string           `json:"body"`
}

func getSqsClient() sqsiface.SQSAPI {
	SQS := sqs.New(session.New())

	return SQS

}

func initParser() {

	flag.StringVar(&queueUrl, "queueUrl", "", "SQS Queue Url")
	flag.StringVar(&service, "service", "", "Local service (http://localhost:8000)")

	flag.Parse()
}

func parseBody(body string) ApiGatewayEvent {
	parsedBody := ApiGatewayEvent{}
	err := json.Unmarshal([]byte(body), &parsedBody)
	if err != nil {
		panic(err)
	}

	return parsedBody
}

func s3GetObject(svc s3iface.S3API, params *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	result, err := svc.GetObject(params)
	if err != nil {
		panic(err)
		log.Printf("ERROR: unable to get: %s from: %s, %v", *params.Key, *params.Bucket, err)
		return nil, err
	}
	return result, nil
}

func LoadBody(s3Url string) ([]byte, error) {
	// Creates new S3 session.
	s3Client := s3.New(session.New())
	// Create a placeholder buffer.
	buf := bytes.NewBuffer(nil)
	// Create GetObjectInput

	urlObject, err := url.Parse(s3Url)
	if err != nil {
		panic(err)
	}

	bucket_name := urlObject.Host
	object_key := urlObject.Path[1:]

	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket_name),
		Key:    aws.String(object_key),
	}

	req, err := s3GetObject(s3Client, params)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	if _, err := io.Copy(buf, req.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func FromEventToRequest(event ApiGatewayEvent, service string) *http.Request {

	method := strings.ToUpper(event.HttpMethod)
	url := service + event.Path
	headers := event.Headers
	body := event.Body

	var req *http.Request
	var err error

	if body != nil {
		s3Url := *body
		requestBody, err := LoadBody(s3Url)
		if err != nil {
			panic(err)
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		panic(err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	return req

}

func DoRequest(request *http.Request) *http.Response {

	var client = &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	return resp
}

func userWorker() {
	sqsClient := getSqsClient()
	for {
		messageInput := &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueUrl),
			MaxNumberOfMessages: aws.Int64(10),
			VisibilityTimeout:   aws.Int64(30),
			WaitTimeSeconds:     aws.Int64(20),
		}

		receive_resp, err := sqsClient.ReceiveMessage(messageInput)
		if err != nil {
			log.Printf("Unable to receive message from queue  %v.", err)
			continue
		}

		if len(receive_resp.Messages) > 0 {
			// doing stuff
			log.Printf("One Message")
			// Delete message
			for _, message := range receive_resp.Messages {

				body := message.Body
				parsedBody := parseBody(*body)

				req := FromEventToRequest(parsedBody, service)
				resp := DoRequest(req)

				log.Printf("%v", resp)
				log.Printf("%v", parsedBody)
				log.Printf("%s", parsedBody.HttpMethod)

				delete_params := &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(queueUrl),  // Required
					ReceiptHandle: message.ReceiptHandle, // Required
				}
				_, err := sqsClient.DeleteMessage(delete_params) // No response returned when successed.
				if err != nil {
					log.Println(err)
				}
				fmt.Printf("[Delete message] \nMessage ID: %s has beed deleted.\n\n", *message.MessageId)
			}
		}
	}
}

func main() {
	initParser()

	log.Printf("Forwarding requests from %s to %s", queueUrl, service)

	userWorker()
}
