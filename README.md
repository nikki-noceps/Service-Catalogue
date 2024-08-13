# serviceCatalogue
A service catalogue with basic support for filtering sorting and pagination

#### **Quick Links**

- [Usage](#running-the-application)
- [User Stories](https://docs.google.com/document/d/1mSUxHGgqKK7xKgIBs9irV_5po_tuqZrHvSRJZ6LhzrU/edit)
- [Figma](https://www.figma.com/design/zeaWiePnc3OCe34I4oZbzN/Service-Card-List?node-id=0-1&t=dnxSHj6txsZ4x4Is-0)
- [Technical Specification](#technical-specification)
- [Postman Collection]()


## Technical Specification
The service catalogue has three major components:

#### 1. Service Catalogue Persistence
The catalogue will be organization agnostic. To make the catalogue specific to organization an extra field will suffice. Elasticsearch will be used to store the catalogue data as it has inherent support for fuzzy search, filtering and sorting. Another case in point is that the service data is not relational, so a non relational database will cater to all our needs. While elasticsearch has inherent versioning it does not keep historical versions of the documents out of the box and only maintains the latest versions.

The document will look as such in `servicecatalogue` index
```json
{
    "documentId": "25ef909a-b7c6-4d9c-9d38-ab9e10fdc686",   // uuid auto-generated 
    "name": "Test",     // user input
    "description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit", // user input
    "createdAt": "2019-11-14T00:55:31.820Z", // ISO-8601 format in UTC to avoid timezone issues
    "updatedAt": "2019-11-15T12:76:21.820Z", // Same as createdAt
    "version": 1,  // increments on every update
    "createdBy": "", // for user information not part of current scope
    "updatedBy": "" // for user information not part of current scope
}
```

Another index called `servicecatalogueversions` will be created to keep track of versions of a service catalogue
The document structure will look as such
```json
{
    "versionId": "25ef909a-b7c6-4d9c-9d38-ab9e10fdc686", // random uuid identifier for this document
    "parentId": "25ef909a-b7c6-4d9c-9d38-ab9e10fdc686",   // matches the document in service catalogue above
    "name": "Test",     // user input
    "description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit", // user input
    "createdAt": "2019-11-14T00:55:31.820Z", // ISO-8601 format in UTC to avoid timezone issues
    "updatedAt": "2019-11-15T12:76:21.820Z", // Same as createdAt
    "version": 1,  // stores the version
    "createdBy": "", // for user information not part of current scope
    "updatedBy": "" // for user information not part of current scope
}
```
The version document is linked to the service document via the `parentId` field which points to the `documentId` in the `servicecatalogue` index. 
This version document shall be created on every update or new service catalogue creation. It will simply be copied during creation of new service or copied from current services in case of updates in services.

#### 2. Fuzzy search, sorting, filtering, version fetching & pagination
Fuzzy Search is an inbuilt feature in elasticsearch and also available on [multiple fields](https://www.elastic.co/guide/en/elasticsearch/guide/current/fuzzy-match-query.html). It will allow us to search on both the name and description of the services. 
Also elasticsearch has inbuilt queries for sorting and filtering for documents based on timestamp ranges and pagination support.
As for fetching the latest versions of documents in the list all queries, we can aggregate on version and fetch top hits by 
descending sort on version.
As for pagination based on our requirements, increasing the `from` parameter should suffice.

> **_NOTE:_**  from suffices to our pagination use case since we are showing only a handful of results on the catalogue page. For a full fledged pagination refer [search after](https://www.elastic.co/guide/en/elasticsearch/reference/current/paginate-search-results.html#search-after) which should be used for pagination at scale.

#### 3. Authentication 

Authentication should ideally be done by an authentication service configured on the api gateway which generates a policy document which allows the request to go through the api gateway and contact the internal services hosted in your VPC. The policy document should be generated only after verifying the identity of the jwt token or base64 token in request with the backend system responsible for authentication. Various cache mechanisms should be put in place to ensure minimal latency on this service.

However for the scope of this project we shall put in place a middleware which does basic authentication using hard coded credentials which should NOT be used in production. There should be a way to invalidate the credentials and generate new credentials on the whim which should be stored in a secure database and hashed supporting decryption with a custom salt addition stored in secure place, ideally we want the hashing key also to be securely stored. This is a symmetric encryption relying on decryption of the token to validate the identity, there are also asymmetric authentication mechanisms using private and public keys. Since the scope of this service is not authentication we shall not delve more on this topic.


### Running the Application

### Testing

## Production Code Practices
In order to make this code production ready, the following things are added to improve certain aspects of the running service:

#### 1. Graceful Shutdown
Graceful shutdown is implemented to improve availability of the service. It handles routing traffic to new pods during deployments while ensuring existing traffic is served without any abrupt closing on connections. It drains existing connection to old containers and ensures the load balancer is able to route new traffic to new containers. To implement this the service should be able to catch a SIGTERM signal usually sent by container orchestrator to stop existing containers, the service then toggles its health checks to ensure the load balancer shifts new traffic to new containers and waits for existing connections to close. Finally the existing containers receive a SIGKILL and gracefully shutdown.

#### 2. RequestId propagation
RequestId propagation is an essential part of tracing requests within a ecosystem of microservices to ensure all logs, spans and responses can be traced via one single requestId. This involved passing a requestId header in inter-service calls, having a middleware which injects the requestId in a request logger and request span.

#### 3. Spans & Metics
Spans and metrics are the ideal way to monitor the performance of your service running in production. While metrics can give a status of latencies, load on the service and performance indicator of various key components like a database of the service at a generic request agnostic level, the spans are more request specific and provide insights into issues or bottlenecks of the application. Spans provide a more indepth analysis of request going through our services.
It is impractical to keep all request spans since this data can be huge across GB so sampling is key over here. Hence metrics provide the rest of the complete picture since metrics do not need sampling.

#### 4. Unit & Integration Testing
Unit testing ensures that the singular functions that an individual contributor writes is correct and function returns as per expectation, the integration testing ensure that a component of the service is functioning as expected despite some changes in the inbuilt implementation. 
Unit & Integration testing should be integrated as checks before PR merge. This not only improves developer confidence but also ensure that changes pushed into the main / master branch are stable and developer tested improving code reliability.
The practise of Test Driven Development ensures that all code is tested and code coverage always remain high for any code repository

#### 5. Seeding and Migrations

Migrations are generally a database concept where in we want the ability to create tables or indexes or collections for relational databases or non relational databases respectively. These are essential to keep track of database structure changes and ensure all local runs have the same database configuration and document or record types.

Seeding the database is essential to ensure the integration tests generate same results across different machines and also help developers do a run of the applications on their local systems. These are also required for Build Verification Tests (BVT test suites) for predeployment automated testing.

#### 6. DockerFile and docker compose
Dockerfile or their alternative are required containerize your applications and create images. Docker images provide isolation by running applications in containers. Because each container has its own filesystem, processes and network stack, dependencies and programs are kept separate from both the host system and each other. This isolation improves security and prevents conflicts between applications
Docker file creates reusable consistent images of your application while docker compose on the other hand simplifies orchestrating and coordinating multiple applications or databases. It helps to replicate the application environment on any machine.

#### 7. Release tags
Following the practice of release tags and adding comments to highlight changes done ensure that reverts are smooth and quick in case of any issues. Seldom hotfixes or reverting code changes require revert PR's on the main branch which is simplified by keeping release tags on each release of the service. 