## kusk mock

Spin up a local mocking server serving your API

### Synopsis

Spin up a local mocking server that generates responses from your content schema or returns your defined examples.
Schema example:

content:
 application/json:
  schema:
   type: object
   properties:
    title:
     type: string
     description: Description of what to do
    completed:
     type: boolean
    order:
     type: integer
     format: int32
    url:
     type: string
     format: uri
   required:
    - title
    - completed
    - order
    - url

The mock server will return a response like the following that matches the schema above:
{
 "completed": false,
 "order": 1957493166,
 "title": "Inventore ut.",
 "url": "http://langosh.name/andreanne.parker"
}

Example with example responses:

application/xml:
 example:
  title: "Mocked XML title"
  completed: true
  order: 13
  url: "http://mockedURL.com"

The mock server will return this exact response as its specified in an example:
<doc>
 <completed>true</completed>
 <order>13</order>
 <title>Mocked XML title</title>
 <url>http://mockedURL.com</url>
</doc>


```
kusk mock [flags]
```

### Examples

```

To mock an api on the local file system
$ kusk mock -i path-to-openapi-file.yaml

To mock an api from a url
$ kusk mock -i https://url.to.api.com

```

### Options

```
  -h, --help          help for mock
  -i, --in string     path to openapi spec you wish to mock
  -p, --port uint32   port to expose mock server on. If none specified, will search for next available port starting from 8080
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.kusk.yaml)
```

### SEE ALSO

* [kusk](kusk.md)	 - 

