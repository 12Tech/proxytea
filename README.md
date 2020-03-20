# ProxyTea

## What is ProxyTea?

ProxyTea is a neat and cost-effective solution to forward HTTP Requests to an internal network.

## How does it work?

ProxyTea is based on two connected components:

- Proxy: A simple AWS Stack, including an API Gateway and an SQS Queue.
- Tea: a command line tool to run locally, pulling requests from the SQS Queue and forwarding them to your local service.

![](resources/PRoxyTea.png)