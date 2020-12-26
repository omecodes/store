[![Build Status](https://travis-ci.com/omecodes/omestore.svg?token=QUyy7EoZqdBaaAXPQDKS&branch=master)](https://travis-ci.com/omecodes/omestore.svg?token=QUyy7EoZqdBaaAXPQDKS&branch=master)
# Omestore 

Omestore is a real time database backend service designed for fast mobile and web apps development. 

On top of an API to C.R.U.D json documents from a MySQL database omestore uses a CEL based security rules layer to manage access to documents.

## Install and run

``` sh
go get github.com/omecodes/omestore
```

### requirements

Omestore only a need a MySQL database installed to run.

### Run

Executing just `./omestore` command will start a server that listens to the port 80 and assumes you have a MySQL database running on `127.0.0.1:3306` accessible with `omestore:omestore` credentials.

To run it using non default port and database, excute it with `--p` and `--dsn` arguments as follow:

``` sh
./omestore --p=8080 --dsn=ome:code@(loclahost:3306)/omestore?charset=utf8
```


## Setting up security rules

Writing security rule consists of using Common Expression Language to create boolean expressions that are executed at runtime against requests context.

### - Request context

The request context is a set of objects that hold information about:

- the API caller

```c
type auth {
    uid string //user id
    email string
    scope string
    group string
}
```

- the targeted data

```c
type data {
    id string //user id
    col string // collection
    creator string
}
``` 

- some global information like the request time

```C
// permissions

type perm {
    read bool
    write bool
    delete bool
    rules bool
    graft bool
}

// acl laods permissions of user 
// identified by 'uid' on data with 'did' as id 
func acl (uid string, did string) perm
```


### - Write and apply rules

Allowing everybody to read document and allowing only creator to edit documents will give:

``` protobuf
    read: true
    write: auth.uid == data.creator
``` 

and to apply these rules on a server that runs at localhost send an http post request as follow:

``` bash
curl --location --request POST 'http://127.0.0.1/.settings/security/rules/access/data' \
--header 'Authorization: Basic YWRtaW46YWRtaW5wYXNzd29yZA==' \
--header 'Content-Type: application/json' \
--data-raw '{
    "read": "true",
    "write": "auth.uid == data.id"
}'
```

## API with swagger

Go to [Swagger online editor](https://editor.swagger.io/) and paste the content of the [omestore API specification](https://github.com/omecodes/omestore/blob/master/api.swagger.yml) to learn more about the Omestore or generate a client code of the language you are working with.