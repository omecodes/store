# Store [unstable]

Store is a Firebase-like backend server application that provides JSON objects and files storages. On top of that resources access are controlled via rule write in CEL (common expression language).

## Install and run

``` sh
go get github.com/omecodes/store
```

### requirements

Store requires only a MySQL database server.

### Run


#### The admin credentials

In order to run a Store server you first have to generate admin credentials with admin-cli.
Build the admin-cli tool with the following command:
```shell
go build -o admin-cli apps/admin-cli.go
```

Then you generate the admin credentials this way:
```shell
./admin-cli auth gen 
```

this command generates a file "admin-auth" that contains a string which is the admin credentials to be used to run an instance of a Store server.


#### Starting a Store server

``` sh
./store run [--dev] --admin="content_of_admin_auth" --db-uri=store:bome@tcp(localhost:3306)/store
```

You can get the usage of the run command with `./store run`

```
Usage:
  store run [flags]

Flags:
      --admin string          Admin password info
      --auto-cert             Run TLS server with auto generated certificate/key pair
      --cert string           Certificate filename
      --db-uri string         MySQL database uri (default "store:store@(127.0.0.1:3306)/store?charset=utf8")
      --dev                   Enable development mode
      --dir string            Data directory (default "./")
      --domains stringArray   Domains name for auto cert
      --fs string             File storage root directory (default "./files")
  -h, --help                  help for run
      --key string            Key filename
      --tls                   Enable TLS secure connexion
      --www string            Web apps directory (apache www equivalent) (default "./www")
```

## Access security rules

Data and files accesses are controlled by rules written in common expression language. Access are checked by executing boolean expression constructed using the following objects:
```C
object user {
    name: string
    group: string
}

object app {
    type: int
    key: string
    info: json
}

object data {
    id: string
    created_by: string 
    created_at: int64
    size: int
}

func now() int64
```

and look like:

```
    user.name==data.created_by && size<1024
```


## JSON objects collection

Like Firebase, Store organize JSON objects in collection in collections. To create a collection use the admin-cli.

```shell
./admin-cli objects collections new --in=collections.json --server="http://localhost:8080/api/objects" --password=security
```

with collections.json file that contains sequence of JSON definition of collections:

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
          "description": "Authenticated users are allowed to delete objects from this collection",
          "rule": "user.name=='admin'"
        }]
      }
    }
  }
}
```

The json above specify a collection of objects that are readable by everyone and editable by authenticated user and can only be deleted by admin.


## API with swagger

Go to [Swagger online editor](https://editor.swagger.io/) and paste the content of the [store API specification](https://github.com/omecodes/store/blob/master/api.swagger.yml) to generate API clients and learn how to communicate with a Store server.