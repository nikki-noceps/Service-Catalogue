# serviceCatalogue
A service catalogue with basic support for filtering sorting and pagination

#### **Quick Links**

- [Usage](#running-the-application)
- [User Stories](https://docs.google.com/document/d/1mSUxHGgqKK7xKgIBs9irV_5po_tuqZrHvSRJZ6LhzrU/edit)
- [Figma](https://www.figma.com/design/zeaWiePnc3OCe34I4oZbzN/Service-Card-List?node-id=0-1&t=dnxSHj6txsZ4x4Is-0)
- [Technical Specification](#technical-specification)
- [Postman Collection](https://api.postman.com/collections/37722834-94e0d2f9-e65d-4e35-b5a8-310dcf21fc02?access_key=PMAT-01J5NN4SE7HJZXKA6N3M7EXGRX)


## Features Included
- [x] :feelsgood: Listing services with sort, filter and versions available. Also fetch all versions
- [x] :feelsgood: Ability to fetch single service and specific version
- [x] :feelsgood: Fuzzy search on name and description
- [x] :feelsgood: Pagination. Caveat [explained](#2-fuzzy-search-sorting-filtering-version-fetching--pagination)
- [x] :feelsgood: Request Validation
- [x] :feelsgood: Graceful Shutdown and health checks
- [x] :feelsgood: Panic Recovery Middleware & RequestId propagation
- [x] :feelsgood: Build checks
- [x] :feelsgood: CRUD endpoints
- [x] :feelsgood: Soft deletes
- [x] :feelsgood: Migration File
- [x] :feelsgood: Cross Origin Resource Sharing (CORS) middleware
- [x] :feelsgood: Custom Context
- [x] :see_no_evil: Poor Man Basic Authentication. All non GET api's have a basic authentication check
- [ ] :hear_no_evil: Seeding File
- [ ] :hear_no_evil: Unit & Integration Tests
- [ ] :hear_no_evil: DockerFile and application containerization

## Technical Specification
The service catalogue has three major components:

#### 1. Service Catalogue Persistence
The catalogue will be organization agnostic. To make the catalogue specific to organization an extra field will suffice. Elasticsearch will be used to store the catalogue data as it has inherent support for fuzzy search, filtering and sorting. Another case in point is that the service data is not relational, so a non relational database will cater to all our needs. While elasticsearch has inherent versioning it does not keep historical versions of the documents out of the box and only maintains the latest versions.

The document will look as such in `servicecatalogue` index
```json
{
    "serviceId": "25ef909a-b7c6-4d9c-9d38-ab9e10fdc686",   // uuid auto-generated 
    "name": "Test",     // user input
    "description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit", // user input
    "createdAt": "2019-11-14T00:55:31.820Z", // ISO-8601 format in UTC to avoid timezone issues
    "updatedAt": "2019-11-15T12:76:21.820Z", // Same as createdAt
    "version": 1,  // increments on every update
    "createdBy": "", // for user information not part of current scope
    "updatedBy": "" // for user information not part of current scope
}
```
The `serviceId` field is a unique identifier for services in the catalogue. 
The `version` field shows latest version of the service. Inherently shows total versions available


Another index called `servicecatalogueversions` will be created to keep track of versions of a service catalogue. This index will be similar to an archive storage and will only store historical versions of a service. The current live version will not reside in this index.
The document structure will look as such
```json
{
    "versionId": "25ef909a-b7c6-4d9c-9d38-ab9e10fdc686", // random uuid identifier for this document
    "parentId": "25ef909a-b7c6-4d9c-9d38-ab9e10fdc686",   // matches the document in service catalogue above
    "name": "Test",     // user input
    "description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit", // user input
    "createdAt": "2019-11-14T00:55:31.820Z", // ISO-8601 format in UTC to avoid timezone issues
    "decomissionedAt": "2020-11-15T13:76:21.820Z", // whenever this version was replaced by a new one
    "version": 1,  // stores the version
    "createdBy": "", // for user information not part of current scope
    "decomissionedBy": "" // store the user who replaced this version
}
```

The `parentId` field links all the historical versions to the live document in servicecatalogue index
The `versionId` field is a unique identifier for each document.
The `decomissioned` fields store information related to when this version was put into archival and by whom.

This version document shall be created on every update or when the live one is deleted from servicecatalogue index. 

> **_NOTE:_** The versions are historical versions and does not store the current version in this index

#### 2. Fuzzy search, sorting, filtering, version fetching & pagination
Fuzzy Search is an inbuilt feature in elasticsearch and also available on [multiple fields](https://www.elastic.co/guide/en/elasticsearch/guide/current/fuzzy-match-query.html). It will allow us to search on both the name and description of the services. 
Also elasticsearch has inbuilt queries for sorting and filtering for documents based on timestamp ranges and pagination support.
As for fetching the latest versions of documents in the list all queries, we can aggregate on version and fetch top hits by 
descending sort on version.
As for pagination based on our requirements, increasing the `from` parameter should suffice. Application has a hard limit of from set at 1000.

> **_NOTE:_**  from suffices to our pagination use case since we are showing only a handful of results on the catalogue page. For a full fledged pagination refer [search after](https://www.elastic.co/guide/en/elasticsearch/reference/current/paginate-search-results.html#search-after) which should be used for pagination at scale roughly if the document count exceeds 5000.

#### 3. Authentication 

Authentication should ideally be done by an authentication service configured on the api gateway which generates a policy document which allows the request to go through the api gateway and contact the internal services hosted in your VPC. The policy document should be generated only after verifying the identity of the jwt token or base64 token in request with the backend system responsible for authentication. Various cache mechanisms should be put in place to ensure minimal latency on this service.

However for the scope of this project we shall put in place a middleware which does basic authentication using hard coded credentials which should NOT be used in production. There should be a way to invalidate the credentials and generate new credentials on the whim which should be stored in a secure database and hashed supporting decryption with a custom salt addition stored in secure place, ideally we want the hashing key also to be securely stored. This is a symmetric encryption relying on decryption of the token to validate the identity, there are also asymmetric authentication mechanisms using private and public keys. Since the scope of this service is not authentication we shall not delve more on this topic.


### Running the Application
Ensure to install docker and keep docker daemon running

Then run the below command in terminal
```
make all
```

This will bring up an elasticsearch container on your machine and create required indexes with required mappings. It will then proceed to run the server. Refer [Postman Collection](https://api.postman.com/collections/37722834-94e0d2f9-e65d-4e35-b5a8-310dcf21fc02?access_key=PMAT-01J5NN4SE7HJZXKA6N3M7EXGRX) for api contracts and usage.


For future runs
```
go run main.go
```

will suffice

>**_NOTE:_** Future runs assume you have an elasticsearch container running with required indexes created

## Production Readiness
For more information, refer [Guide](docs/production_readiness.md)