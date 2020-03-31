# Auth

An Authentication service for my personal web app. This service uses Redis as
it's datastore, mainly for it's speed and ability to persist to storage.

This service uses [JSON Web Tokens](https://jwt.io/) to verify a users session.
For security reasons these tokens expire within a short amount of time, to prevent
getting logged out every 10-15 minutes this service also provides refresh tokens
which allows for renewing sessions. A more detailed explanation of this pattern can
be found [here](https://hasura.io/blog/best-practices-of-using-jwt-with-graphql/).

## Getting Started

The following instructions will help you spin up a local copy of the service for
tesing purposes.

### Prerequisites

This service is intended to be run within a docker container in order to keep
the installation simple.

Documentation on installing docker can be found [here](https://docs.docker.com/install/).
Once installed, run the command below to build the docker image for this service:

```bash
docker build --tag auth .
```

More information about how to develop with docker can be found [here](https://docs.docker.com/develop/)

## Running Tests

Unit tests can be located within individual packages, an example command for
running one of the unit tests would be:

```bash
go test -v pkg/auth/service_test.go
```

The go command above will test the core functionality for this service. In order
to keep the test as simple as possible, all the requests to redis are being mocked.

More information about running units tests in Go can be found [here](https://golangdocs.com/unit-testing-in-golang).

## Built With

* [jwt-go](https://github.com/dgrijalva/jwt-go)
* [go-redis](https://github.com/go-redis/redis)
* [protobuf](https://github.com/golang/protobuf)
* [gRPC](https://grpc.io/)

## License

This project is licensed under the BSD License - see the [LICENSE](LICENSE)
file for details.
