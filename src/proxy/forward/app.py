import io
import json
import os
from uuid import uuid4

import boto3

sqs = boto3.client('sqs')
s3_client = boto3.client('s3')

queue_url = os.getenv('SQS_URL')
bucket_name = os.getenv('BUCKET_NAME')
prefix = os.getenv('BODY_PREFIX', 'proxytea')

def dump_body(body: bytes) -> str:

    body_uuid = str(uuid4())
    object_prefix = body_uuid[:2]

    object_key = f"{prefix}/{object_prefix}/{body_uuid}"

    s3_client.put_object(
        Body=body.encode('utf-8'),
        Bucket=bucket_name,
        Key=object_key
    )

    return f"s3://{bucket_name}/{object_key}"

def lambda_handler(event, context):

    dedup = str(uuid4())

    body = event.get('body')

    if body is not None:
        event['body'] = dump_body(body)

    sqs.send_message(
        QueueUrl=queue_url,
        MessageBody=json.dumps(event),
        MessageGroupId=dedup
    )

    return {
        "statusCode": 200,
        "body": json.dumps(event, indent=4)
    }
