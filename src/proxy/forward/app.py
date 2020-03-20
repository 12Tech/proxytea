import json
import os
from uuid import uuid4

import boto3

sqs = boto3.client('sqs')
queue_url = os.getenv('SQS_URL')
bucket_name = os.getenv('BUCKET_NAME')

def lambda_handler(event, context):

    dedup = str(uuid4())

    body = event.get('body')


    sqs.send_message(
        QueueUrl=queue_url,
        MessageBody=json.dumps(event),
        MessageGroupId=dedup
    )

    return {
        "statusCode": 200,
        "body": json.dumps(event, indent=4)
    }
