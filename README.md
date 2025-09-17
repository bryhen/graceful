# Overview

Package graceful manages the sequential startup, monitoring, and shutdown of an application.

An application could configure graceful to take the following steps:
1) Concurrently connect to Postgres, connect to Redis, and allocate resources for a batch process.
2) Warm-up an in-memory cache with data from Redis.
3) Start an HTTP server to begin serving live traffic.
4) Wait for a shutdown signal (ie SIGTERM from Docker).
5) Shutdown the HTTP server to stop receiving live traffic and wait for all handlers to return.
6) Shutdown the batch process, to ensure the last batch is handled.
