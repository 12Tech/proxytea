import argparse
import json
import traceback
from urllib.parse import urlparse

from requests import Request, Session

import boto3

sqs = boto3.resource('sqs')
s3_client = boto3.client('s3')


def read_messages_from_queue(queue):

    remote_queue = sqs.Queue(queue)
    yield from remote_queue.receive_messages(
            MaxNumberOfMessages=10,
            VisibilityTimeout=300,
            WaitTimeSeconds=20)

def compose_request(message, url):

    event = json.loads(message.body)

    print(json.dumps(event, indent=3))

    return from_event_to_request(event, url)


def load_body(s3_url: str) -> bytes:

    parsed = urlparse(s3_url)
    bucket_name = parsed.netloc
    object_key = parsed.path[1:]

    print(bucket_name, object_key)

    response = s3_client.get_object(
        Bucket=bucket_name,
        Key=object_key
    )

    return response['Body'].read()


def from_event_to_request(event, url):

    method = event['httpMethod']
    path = event.get('path', "")
    url = f"{url}{path}"
    headers = event['headers']
    body = event.get('body')

    if body is not None:
        body = load_body(body)

    r = Request(method, url, headers, data=body)

    return r.prepare()


def do_request(request):

    s = Session()
    response = s.send(request)

    return response

def to_curl(req):

    command = "curl -X {method} -H {headers} -d '{data}' '{uri}'"
    method = req.method
    uri = req.url
    data = req.body
    headers = ['"{0}: {1}"'.format(k, v) for k, v in req.headers.items()]
    headers = " -H ".join(headers)
    return command.format(method=method, headers=headers, data=data, uri=uri)


def main():

    parser = argparse.ArgumentParser()
    parser.add_argument("-q", "--queue", type=str, help="Queue Url to read messages from")
    parser.add_argument("-u", "--url", type=str, help="Base url to forward message to")
    args = parser.parse_args()

    if args.url and args.queue:

        queue = args.queue
        url = args.url

        while True:

            for message in read_messages_from_queue(queue):
                request = compose_request(message, url)
                response = do_request(request)
                try:
                    response.raise_for_status()
                except Exception as e:
                    traceback.print_exc()
                    message.delete()
                    continue
                finally:
                    print(to_curl(response.request))





if __name__ == '__main__':
    main()
