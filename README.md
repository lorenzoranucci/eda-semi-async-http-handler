# Event-Driven Architecture: semi-async HTTP handler in Go

This project aims to demonstrate how it is possible to handle each HTTP request asynchronously (e.g., using events on
Kafka) on the server side while maintaining a synchronous HTTP interface.

This approach guarantees at-least-once event delivery and does not require breaking changes to the HTTP API (i.e. REST,
GraphQL, etc)

## Context

In event-driven architectures (EDA) it is essential to ensure that events are published at least once, eventually.

The dual-write problem can undermine this guarantee. Example:

```go
repository.Add(user) // state mutation

// error here due to ungraceful shutdown

eventBroker.Publish(UserCreatedEvent{User: user}) // or error here due to network issues or service unavailable
```

In this case, the application's state was changed, but the outside world was not notified of the event.

The application of **retry logic** is essential in order to ensure that the event is **eventually published at least
once**.

However, it is not always straightforward to apply retry logic especially if we are in the context of a HTTP request and
do not want to block the user indefinitely or cannot force the user to re-post the request.

There are several patterns that help us implement retry. One is the outbox pattern, which we have already seen in
another project. Another one is making every HTTP request an async event.

## Make everything async

In this project we have a demonstration of how we can apply the retry pattern by making sure that we handle any
state mutation within an event consuming context instead of the HTTP request context.

This approach ease the application of the retry pattern because we can simply retry the event consumption easily.
The only caveat is that the consuming of the event is idempotent.

So any HTTP request for state mutation, is immediately transformed into an event. This event is processed and the result
is obtained asynchronously. Usually this architecture also involves the front-end client by implementing patterns
such as [Asynchronous Request-Reply](https://learn.microsoft.com/en-us/azure/architecture/patterns/async-request-reply)
or Websocket for asynchronous reading of the outcome.

In some cases, however, it may not be possible or desirable to change the behavior of front-end clients to be
asynchronous. This project aims to demonstrate that this problem can be worked around by injecting asynchronous outcome
handling into the HTTP handler itself, instead of the frontend client.

### How it works

The HTTP handler turns the request into an event (e.g., `HTTPRequestReceived`) and publishes it to Kafka, attaching a
unique ID for the request.

A Kafka consumer, which handles the event, performs the requested action and publishes an outcome event (
e.g. `UserCreated` or `UserDuplicated`), again attaching the unique ID for the request.

Another Kafka consumer, executed in the same process as the initial HTTP handler but in a separate goroutine, handles
the outcome event and responds to the initial HTTP handler with the result. It can associate the outcome event with the
initial HTTP request by using the unique ID.

At this point the HTTP handler can return the response to the HTTP client.

### Running the demo

Make sure ports `8080`, `8082`, `3306` and `9093` are available. 
Otherwise, you can change the ports in the `docker-compose.override.yaml` file.

Start the services:
```shell
docker compose up -d --build --force-recreate
```

Make sure that every service is up and running.

Send a request to the HTTP handler:
```shell
curl -X POST --location "http://localhost:8080" \
    -H "Content-Type: application/json" \
    -d "{
          \"name\": \"Lorenzo Ranucci\"
        }"
```



Connect to http://localhost:8082 and watch the Kafka topics.
