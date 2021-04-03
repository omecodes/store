# Store [unstable]

Store is a backend application that combines:

- A JSON document storage with a search engine to find according to json path value
- A file storage
- A search engine for files and JSON documents
- A rule-based ACL mechanism over files and JSON documents

# Install and setup


### Requirements

Store requires only a MySQL database that runs at:

```
store:store@tcp(localhost:3306)/store
```

### Build store and admin-cli

```shell
git clone https://github.com/omecodes/store.git
cd store
go get -v -t -d ./...

go build -o store ./apps/server.go
go build ./apps/admin-cli.go
```

### Generate the admin credentials

In order to run a Store server you first have to generate admin credentials with admin-cli.
Build the admin-cli tool with the following command:
```shell
go build -o admin-cli apps/admin-cli.go
```

Then you generate the admin credentials:
```shell
./admin-cli auth gen 
```

this command generates a file "admin-auth" that contains a string which is the admin credentials to be used to run an instance of a Store server.

### Start the server

Execute this command to run a server:
```shell
./store run [--dev] --admin="content_of_admin_auth"
```

The `run` command supports additional flags and, you can display them by simply run `./store run`:
```
Usage:
  store run [flags]

Flags:
      --admin string          Admin password info
      --auto-cert             Run TLS server with auto generated certificate/key pair
      --cert string           Certificate filename
      --db-uri string         MySQL database uri (default "store:store@(127.0.0.1:3306)/store?charset=utf8")
      --dev                   Enable development mode. Enables CORS
      --dir string            Data directory (default "./")
      --domains stringArray   Domains name for auto cert
      --fs string             File storage root directory (default "./files")
  -h, --help                  help for run
      --key string            Key filename
      --tls                   Enable TLS secure connexion
      --www string            Web apps directory (apache www equivalent) (default "./www")
```

### Authentication

Store allows authentication for both users and client applications. 

#### 1 - Registering a client application

```shell
./admin-cli auth access set --server=http://localhost:8080/api/auth --in=accesses.json --password=<admin-secret>
```

where `access.json` file that contains sequence of json:

```json
{
  "key": "client-app1",
  "secret": "client-app1-secret",
  "collections": {
    "create": true,
    "read": false
  },
  "sources": {
    "restricted": true,
    "create": true,
    "view": true
  },
  "users": {
    "create": true,
    "view": true
  }
}
```

#### 2 - Registering users

```shell
./admin-cli auth users new --server=http://localhost:8080/api/auth --in=users.json --password=<admin-secret>
```

where `access.json` file that contains sequence of json:

```json
{
  "username": "user1",
  "password": "user1-password"
}
{
  "username": "user2",
  "password": "user2-password"
}
```

#### 3 - Authenticated request

Authentication data are passed through request HTTP headers. Client applications and users authentication are passed through `X-STORE-CLIENT-APP-AUTHENTICATION` and `Authentication`HTTP headers respectively.

# ACL: security rules

Setting a resource access rules consists of creating a set of boolean expressions that are written in common expression language. Rules are constructed using the following items:

```
object user {
    name: string
    group: string
}

object app {
    type: int
    key: string
    collections: struct {
        create: bool
        view: bool
        delete: bool
    }
    sources: struct {
        create: bool
        view: bool
        delete: bool
        restricted: bool
    }
    users: struct {
        create: bool
        view: bool
        delete: bool
    }
    info: json
}

object data {
    id: string
    created_by: string 
    created_at: int64 // in second
    size: int // bytes
}

func now() int64
```

Below are examples of valid rules:

```
    user.name==data.created_by && size<1024 // user is the creator of data and data size are lower than 1024
    user.name!='' && app.key!='' // both client app and user are authenticated 
```

# Document collections and files sources

### JSON document collection

JSON documents are organized in collections. Create a collection with the following command:

```shell
./admin-cli objects collections new --in=collections.json --server="http://localhost:8080/api/objects" --password=<admin-secret>
```

with the `collections.json` file that contains sequence of JSON definition of collections:

```json
{
  "id": "images",
  "label": "Images",
  "description": "Image registry",
  "text_indexes": [
    {
      "path": "$.label"
    },
    {
      "path": "$.description"
    }
  ],
  "default_access_security_rules": {
    "access_rules": {
      "$": {
        "read": [{
          "name": "read",
          "label": "Read permission",
          "description": "Everybody can read objects from this collection",
          "rule": "true"
        }],
        "write": [{
          "name": "write",
          "label": "Write permission",
          "description": "Authenticated users are allowed to edit objects from this collection",
          "rule": "user.name!=''"
        }],
        "delete": [{
          "name": "delete",
          "label": "Delete permission",
          "description": "Only admin is allowed to delete objects from this collection",
          "rule": "user.name=='admin'"
        }]
      }
    }
  }
}
```

the `text_indexes` field is set as info for the search engine. it indicates that documents of the collection can be found in a search request if the search query pattern matches value at one `$.label` and `$.description` json paths. That also forces the type of value at these json paths.

### File sources

A file source is a definition of a file tree located somewhere on the computer the server runs in or on a computer somewhere in the internet. File source definition provide location where file operations are executed
Here is how to create a file source using the admin-cli:

```shell

./admin-cli files sources new --in=source.json --server=http://localhost:8080/api/files --password=<admin-secret>
```

with the `source.json` file that contains sequence of JSON definition of collections:

```json
{
  "id": "file-source-id",
  "label": "Some file source",
  "type": 1,
  "description": "Some file source",
  "created_by": "admin",
  "uri": "files:///path/to/a/folder",
  "permission_overrides": {
    "read": [{
      "name": "read",
      "label": "Read permission",
      "description": "Everybody can read files from this source",
      "rule": "true"
    }],
    "write": [{
      "name": "write",
      "label": "Write permission",
      "description": "Authenticated users are allowed to edit files from this source",
      "rule": "user.name!=''"
    }],
    "delete": [{
      "name": "delete",
      "label": "Delete permission",
      "description": "Only admin is allowed to delete files from this source",
      "rule": "user.name=='admin'"
    }]
  }
}
```

## API with swagger
The [Store API specification](https://github.com/omecodes/store/blob/master/api.swagger.yml) helps you understand how to build HTTP requests to:

- Create JSON documents
- Read JSON documents
- Search for JSON documents
- Delete JSON documents
- Create files and directories
- Upload and download files
- Search for files and directories
- Share files 

If you are not familiar with the raw API specification you can copy its content [here](https://editor.swagger.io/) for a GUI display.
