## Production Code Additions
In order to make this code production ready, the following things are added to improve certain aspects of the running service:

#### 1. Graceful Shutdown
Graceful shutdown is implemented to improve availability of the service. It handles routing traffic to new pods during deployments while ensuring existing traffic is served without any abrupt closing on connections. It drains existing connection to old containers and ensures the load balancer is able to route new traffic to new containers. To implement this the service should be able to catch a SIGTERM signal usually sent by container orchestrator to stop existing containers, the service then toggles its health checks to ensure the load balancer shifts new traffic to new containers and waits for existing connections to close. Finally the existing containers receive a SIGKILL and gracefully shutdown.

#### 2. Important Middlewares
RequestId propagation is an essential part of tracing requests within a ecosystem of microservices to ensure all logs, spans and responses can be traced via one single requestId. This involved passing a requestId header in inter-service calls, having a middleware which injects the requestId in a request logger and request span.

Panic handler is another important middleware which can help save panics in the application and prevent the application from crashing. It is also equally important to ensure the panic handler does not silently panic but throws stack trace logs and relevant metrics on which alerts can be configured to notify of any prevalent crashes in the system.

> **_NOTE:_**  Recover (the panic handler for go) only works when called in the goroutine which panics which essentially means any additional goroutines spawned during request handling will required their own panic handler to be implemented.

Cross Origin Request Sharing middleware is also required for preflight requests for browser specific requests. It shares the request methods supported by your application and allowed domains which can communicate with the application.

#### 3. Seeding and Migrations
Migrations are generally a database concept where in we want the ability to create tables or indexes or collections for relational databases or non relational databases respectively. These are essential to keep track of database structure changes and ensure all local runs have the same database configuration and document or record types.

Seeding the database is essential to ensure the integration tests generate same results across different machines and also help developers do a run of the applications on their local systems. These are also required for Build Verification Tests (BVT test suites) for predeployment automated testing.

#### 4. Connection Pooling & Bulk Queries
To run applications in especially with a lot of database operations its is recommended to set up a pool of reusable connections to the database or any external resources. This saves up time and resources spent establishing tcp connections. 
In case of any small increments or multiple transitions of minor data which is essentially stateless it is recommended to use bulk operations in atomic transactions generally supported by most database systems.

> **_NOTE:_**  Since elasticsearch is a api based database, it does not require explicit connection pool management from the user's end, as it internally handles connection pooling using the http.Transport provided by Go's standard library. However we can specify custom transport in the client config if we wish for control on these configurations

#### 5. Spans & Metrics
Spans and metrics are the ideal way to monitor the performance of your service running in production. While metrics can give a status of latencies, load on the service and performance indicator of various key components like a database of the service at a generic request agnostic level, the spans are more request specific and provide insights into issues or bottlenecks of the application. Spans provide a more indepth analysis of request going through our services.
It is impractical to keep all request spans since this data can be huge across GB so sampling is key over here. Hence metrics provide the rest of the complete picture since metrics do not need sampling.

#### 6. Unit & Integration Testing
Unit testing ensures that the singular functions that an individual contributor writes is correct and function returns as per expectation, the integration testing ensure that a component of the service is functioning as expected despite some changes in the inbuilt implementation.  

> **_NOTE:_** It is recommended to use interfaces for your service level functions. Go mocks is a standard mocking library which can be easily integrated into unit tests to specify behavior of functions in tests.

Unit & Integration testing should be integrated as checks before PR merge. This not only improves developer confidence but also ensure that changes pushed into the main / master branch are stable and developer tested improving code reliability.
The practise of Test Driven Development ensures that all code is tested and code coverage always remain high for any code repository

#### 7. DockerFile and docker compose
Dockerfile or their alternative are required containerize your applications and create images. Docker images provide isolation by running applications in containers. Because each container has its own filesystem, processes and network stack, dependencies and programs are kept separate from both the host system and each other. This isolation improves security and prevents conflicts between applications
Docker file creates reusable consistent images of your application while docker compose on the other hand simplifies orchestrating and coordinating multiple applications or databases. It helps to replicate the application environment on any machine.

#### 8. Release tags
Following the practice of release tags and adding comments to highlight changes done ensure that reverts are smooth and quick in case of any issues. Seldom hotfixes or reverting code changes require revert PR's on the main branch which is simplified by keeping release tags on each release of the service. 