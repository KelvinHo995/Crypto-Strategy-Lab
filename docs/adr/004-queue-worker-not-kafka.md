# ADR 004: Why Queue/Worker Pool (and why NOT Kafka)?

## Context
The Experiment domain needs to execute hundreds or thousands of backtest jobs when a search is initiated. We need a way to distribute these jobs to workers efficiently to maximize throughput, while avoiding blocking the main HTTP thread that received the `/search/start` request.

## Decision
We decided to implement an in-process Job Queue and Worker Pool using native Go channels (`chan Job`) and goroutines.
- A buffered channel acts as the central queue.
- A configurable number of worker goroutines pull jobs from the channel and execute them concurrently.
- The HTTP handler simply pushes generated `CandidateStrategy` jobs into the channel and returns immediately.

## Alternatives Considered
- **External Message Brokers (Kafka, RabbitMQ)**: Highly scalable and durable. However, they add significant operational complexity (managing Zookeeper/Kafka clusters), infrastructure overhead, and learning curve for a 2-week lab project. Our system currently runs as a single monolithic instance, so cross-network message brokers are unnecessary.
- **Redis Queue (e.g., asynq)**: Easier than Kafka, but still requires a separate Redis instance. Given our expected workload and constraints, adding Redis is still over-engineering.
- **Synchronous Execution**: The HTTP handler waits for all backtests to finish. This would lead to HTTP timeouts and terrible user experience.

## Consequences
- **Pros**: 
  - Zero external dependencies.
  - Extremely fast (in-memory communication).
  - Easy to implement and test within Go.
  - Scales vertically very well up to the CPU limits of the machine.
- **Cons**: 
  - **No Durability**: If the server crashes, all jobs currently in the queue are lost. For this lab, a user can simply click "Search" again, making this an acceptable tradeoff.
  - **Horizontal Scaling Limit**: We cannot easily distribute the queue across multiple servers if we outgrow a single machine (though we could scale up the machine first).

## Evidence
- (To be filled by Person 3 during Throughput tests): We observed X jobs/sec with 1 worker, scaling to Y jobs/sec with N workers, proving vertical scalability.
