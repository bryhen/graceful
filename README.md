# Overview

Package graceful manages the sequential startup, monitoring, and shutdown of an application.

An application could configure graceful to take the following steps:
1) Concurrently: connect to Postgres, connect to Redis, and allocate resources for a batch processing goroutine related to HTTP requests.
2) Warm-up an in-memory cache by loading data from Redis, which was previously connected.
3) Start an HTTP server to begin serving live traffic.
4) Wait for a shutdown signal (ie SIGTERM from Docker).
5) Shutdown the HTTP server to stop receiving live traffic and wait for all handlers to return.
6) Shutdown the batch process by handling the last batch since no more HTTP requests will be received.
