# format-graphql

[![GitSpo Mentions](https://gitspo.com/badges/mentions/gajus/format-graphql?style=flat-square)](https://gitspo.com/mentions/gajus/format-graphql)
[![Travis build status](http://img.shields.io/travis/gajus/format-graphql/master.svg?style=flat-square)](https://travis-ci.org/gajus/format-graphql)
[![Coveralls](https://img.shields.io/coveralls/gajus/format-graphql.svg?style=flat-square)](https://coveralls.io/github/gajus/format-graphql)
[![NPM version](http://img.shields.io/npm/v/format-graphql.svg?style=flat-square)](https://www.npmjs.org/package/format-graphql)
[![Canonical Code Style](https://img.shields.io/badge/code%20style-canonical-blue.svg?style=flat-square)](https://github.com/gajus/canonical)
[![Twitter Follow](https://img.shields.io/twitter/follow/kuizinas.svg?style=social&label=Follow)](https://twitter.com/kuizinas)

Formats GraphQL schema definition language (SDL) document.

---

* [Motivation](#motivation)
* [Behaviour](#behaviour)
  * [Example](#example)
* [Usage](#usage)
  * [Command Line](#command-line)
  * [Node API](#node-api)
  * [Hooks](#hooks)

## Motivation

As schema grows in size, it becomes desirable to automate schema organisation. The primary function of `format-graphql` is to sort definitions and fields in an alphabetical order, therefore enabling predictable discovery and grouping of related schema entities.

## Behaviour

Alphabetically sorts definitions, fields and arguments.

### Example

Input:

```graphql
type Query {
  bananas: [Banana!]!
  apples: [Apple!]!
}

type Apple {
  name: String!
  id: ID!
}

type Banana {
  name: String!
  id: ID!
}

```

Output:

```graphql
type Apple {
  id: ID!
  name: String!
}

type Banana {
  id: ID!
  name: String!
}

type Query {
  apples: [Apple!]!
  bananas: [Banana!]!
}

```

## Usage

### Command Line

```bash
$ format-graphql --help
Sort GraphQL schema definition language (SDL) document.

Positionals:
  sdl-path  Path to the GraphQL schema definition (SDL) document.       [string]

Options:
  --version           Show version number                              [boolean]
  --help              Show help                                        [boolean]
  --sort-arguments    Sort on arguments                [boolean] [default: true]
  --sort-definitions  Sort on definitions              [boolean] [default: true]
  --sort-enums        Sort on enums                    [boolean] [default: true]
  --sort-fields       Sort on fields                   [boolean] [default: true]
  --write             Overrides contents of the SDL document.
                                                      [boolean] [default: false]

$ # Prints formatted schema.
$ format-graphql ./schema.graphql
$
$ # Overrides target schema.
$ format-graphql --write=true ./schema.graphql

```

### Node API

`formatSdl(schema, options)`

Returns a formatted GraphQL SDL String.

#### Parameters

- `schema`: string
- `options` (optional): object:

  ```
  {
    sortDefinitions?: boolean,
    sortEnums?: boolean,
    sortFields?: boolean,
    sortArguments?: boolean,
  }
  ```

#### Example

```js
import {formatSdl} from 'format-graphql';

formatGraphql('type Foo { bar: String }');
```

### Hooks

I recommend using [husky](https://www.npmjs.com/package/husky) to setup a pre-commit hook that would format the schema, e.g.

```json
"husky": {
  "hooks": {
    "pre-commit": "format-graphql --write true src/schema.graphql"
  }
},

```
