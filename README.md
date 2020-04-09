# Auth

An Authentication service for my personal web app. This service uses Redis as
it's datastore, mainly for it's speed and ability to persist to storage.

This service uses [JSON Web Tokens](https://jwt.io/) to verify a users session.
For security reasons these tokens expire within a short amount of time and can
be blacklisted, to prevent getting logged out every 10-15 minutes this service
also provides refresh tokens which allow for renewing sessions. A more detailed
explanation of this pattern can be found [here](https://hasura.io/blog/best-practices-of-using-jwt-with-graphql/).

Since Redis tries to keep things simple, it unfortunately doesn't come with a
way to expire members of a sorted set. This service uses sorted sets to keep
track of blacklisted JWT's, and as you could imagine this set would start to
eat up memory over a long period of time. To overcome this the authentication
service has a seperate goroutine running in the background that periodically
flushs the blacklist of expired tokens.

## Getting Started

The following instructions will help you spin up a local copy of the service for
tesing purposes.

## Configuration

This service uses a configuration file to set the gRPC server address and the address
to a redis instance, these two fields are mandatory as there really isn't any
default. Within the configuration file you can also specify token generation parameters
such as the length and expiration of a token, these fields aren't required. An
example config file can be found [here](config/config.yml).

**NOTE**: Cipher keys need to be 32 characters long.

### Defaults

These are the default values for the service configuration:

| Field Name         | Value        |
|--------------------|--------------|
| Address            | None         |
| Repo Address       | None         |
| Cipher Keys        | None         |
| Cipher Salt Length | 16 Bytes   	|
| Refresh Length     | 32 Bytes   	|
| Refresh Expiration | 24 Hours   	|
| JWT Expiration     | 15 Minutes 	|

## Building

### Prerequisites

* This service is intended to be run within a docker container in order to keep
the installation simple. Documentation on installing docker can be found [here](https://docs.docker.com/install/).

* Git will also need to be installed in order to fetch the repository from Github.
Installation instructions can be found [here](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

* If building locally you will need to install Go version 1.14 or greater. [Instructions](https://golang.org/doc/install).

#### Fetch from Github

The following commands will clone a local copy of the service and put you in the
root of the project where the next steps will need to take place:

```bash
mkdir $HOME/src
cd $HOME/src
git clone https://github.com/joshturge-io/auth.git
cd auth
```

#### Building with docker

More information about how to develop with docker can be found [here](https://docs.docker.com/develop/).
This command will build an image from the [Dockerfile](build/package/Dockerfile) in the
projects root:

```bash
docker build -f build/package/Dockerfile --tag auth . 
```

#### Building locally

##### Windows

You can just build the project directly without docker by using
the command below:

```bash
mkdir bin/
go build -o bin/auth.exe cmd/auth/main.go
```

##### POSIX

I have provided a [Makefile](Makefile) which can be used on POSIX operating systems.
The file comes with two targets, the first being `build` which will build the project
and the second target `clean` will just remove the bin directory. Example usage:

```bash
make build
```

**NOTE**: The build target will place the resulting binary within the `bin/` directory

## Running

### Before you run

**NOTE:** Since I haven't set up a docker-compose file you won't be able to use
a real redis instance at the moment, but I have provided a test struct that satisfies
the repository interface. In order to run the test repo you need to set an
environment varible which will be listed below.

The following environment variables can be set to change some functionality of
the service:

* `JWT_SECRET` - the secret used to sign JSON Web Tokens, this variable is
mandatory otherwise the service will fail to start.

* `REPO_PSWD` - the password for the redis instance, this variable can be left unset
if there isn't a password.

* `TEST_REPO` - if this variable is set, the test repository will be used
instead of the redis repository. This is extremely useful when testing. **NOTE:**
since this is just a test repository the username can be set to any string you
like, the password for the test user is `123password`. I know, very creative.

#### Running with docker

The following command will run an instance of the docker image we built previously:

```bash
docker run -p 6380:8080 -e JWT_SECRET='secret' -e TEST_REPO=true auth
```

I've set this container to bind to port `6380` however this could be anything. I'm
also using the test repository for the example above, however if you have a real
redis instance running feel free to use that instead.

#### Running locally

By default, the auth service will look for a `config.yml` inside the directory
it was run from. To specify an alternative directory you can use the
`config` argument when running. For example:

```bash
./auth -config=config/
```

## Running Tests

In order to run one of the tests for this project, you will need a redis instance 
and the environment variable `REDIS_ADDR` set to the address of that redis instance.
The easiest way to setup a redis instance is with the [Official Docker Image](https://hub.docker.com/_/redis/). The following command will test all the packages for the project:

```bash
go test -v ./pkg/...
```

In the near future I will write a mock interface for the redis package to remove the
need for a redis instance, but for now this is the only way.

More information about running units tests in Go can be found [here](https://golangdocs.com/unit-testing-in-golang).

## Built With

* [jwt-go](https://github.com/dgrijalva/jwt-go)
* [go-redis](https://github.com/go-redis/redis)
* [protobuf](https://github.com/golang/protobuf)
* [gRPC](https://grpc.io/)
* [viper](https://github.com/spf13/viper)

## License

This project is licensed under the BSD License - see the [LICENSE](LICENSE)
file for details.
