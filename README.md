# Router Builder

Router builder will works as a bridge between the workflow manager and the corresponding adapter, it will receive a bunch of routers to create/delete on a 'routers.create' and it will emit messages 'router.create' to process them. 

Once all of them are fully processed it will emit back a message 'routers.create.done' as confirmation or 'routers.create.error' in case something is broken

## Build status

* master:  [![CircleCI Master](https://circleci.com/gh/r3labs/router-builder/tree/master.svg?style=svg&circle-token=627e89c447fe342aff9815ca146b081a37c075ad)](https://circleci.com/gh/r3labs/router-builder/tree/master)
* develop: [![CircleCI Develop](https://circleci.com/gh/r3labs/router-builder/tree/develop.svg?style=svg&circle-token=627e89c447fe342aff9815ca146b081a37c075ad)](https://circleci.com/gh/r3labs/router-builder/tree/develop)

## Installation

```
make deps
make install
```


## Running Tests

```
make test
```

## Contributing

Please read through our
[contributing guidelines](CONTRIBUTING.md).
Included are directions for opening issues, coding standards, and notes on
development.

Moreover, if your pull request contains patches or features, you must include
relevant unit tests.

## Versioning

For transparency into our release cycle and in striving to maintain backward
compatibility, this project is maintained under [the Semantic Versioning guidelines](http://semver.org/). 

## Copyright and License

Code and documentation copyright since 2015 r3labs.io authors.

Code released under
[the Mozilla Public License Version 2.0](LICENSE).
