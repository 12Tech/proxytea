import json
import os
from uuid import uuid4

import boto3

sqs = boto3.client('sqs')
queue_url = os.getenv('SQS_URL')

def lambda_handler(event, context):
    """Sample pure Lambda function

    Parameters
    ----------
    event: dict, required
        API Gateway Lambda Proxy Input Format

        Event doc: https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html#api-gateway-simple-proxy-for-lambda-input-format

    context: object, required
        Lambda Context runtime methods and attributes

        Context doc: https://docs.aws.amazon.com/lambda/latest/dg/python-context-object.html

    Returns
    ------
    API Gateway Lambda Proxy Output Format: dict

        Return doc: https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html
    """

    dedup = str(uuid4())

    sqs.send_message(
        QueueUrl=queue_url,
        MessageBody=json.dumps(event),
        MessageGroupId=dedup
    )

    return {
        "statusCode": 200,
        "body": json.dumps(event, indent=4)
    }
