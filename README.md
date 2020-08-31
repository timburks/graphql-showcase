# GraphQL Showcase

## A GraphQL wrapper for (parts of) the [GAPIC Showcase API](https://github.com/googleapis/gapic-showcase).

This project contains a simple and rough wrapper for a few methods
of the GAPIC Showcase API, and is intended to provide an example of
manual integration of a gRPC API into a GraphQL service.

It uses the [graphql-go](https://github.com/graphql-go/graphql) package.

**This repository is temporary.**
This code is published to support the workshop
[Implementing OpenAPI and GraphQL Services with gRPC](https://asc2019.sched.com/event/T6u9/workshop-implementing-openapi-and-graphql-services-with-grpc-tim-burks-google)
at the 2019 API Specifications Conference.
It is intended for submission later into an `experimental` 
subdirectory of [GAPIC Showcase](https://github.com/googleapis/gapic-showcase).

## Credits
Contents of the `static` directory are manually vendored from
[github.com/graphql/graphiql](https://github.com/graphql/graphiql).

## Installation
`go get github.com/timburks/graphql-showcase`

## Invocation
Just run the `graphql-showcase` program. It currently takes no options
and expects a GAPIC Showcase server to be running locally on port 7469
(the default).

## Usage

After you've started the `graphql-showcase` server, visit http://localhost:8080
to open the GraphiQL browser. Then use standard GraphQL to explore the schema
and make queries. For example, to create a new user, enter:
```
mutation {
  createUser(display_name:"me", email:"me@example.com") {
    id
    display_name
    email
  }
}
```
To see a list of users, enter:
```
{
  users {
    id
    display_name
    email
  }
}
```
To delete a user, enter:
```
mutation {
  deleteUser(id:"users/10") {
    id
  }
}
```

## Go Version Supported
This code was developed with Go 1.12.

## Disclaimer
This is not an official Google product.
