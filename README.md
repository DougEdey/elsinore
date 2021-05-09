# HEAVILY UNDER DEVELOPMENT

I'm finally getting off my ass and working on the new version of Elsinore.

[![golangci-lint](https://github.com/DougEdey/elsinore/actions/workflows/linter.yml/badge.svg)](https://github.com/DougEdey/elsinore/actions/workflows/linter.yml) [![Test](https://github.com/DougEdey/elsinore/actions/workflows/test.yml/badge.svg)](https://github.com/DougEdey/elsinore/actions/workflows/test.yml)

## Rules of development

Rules are

1) Must be written in Go for the backend
2) GraphQL must be the API layer
3) Do whatever you want with the frontend, I'll provide a functional React based front end so you can easily run it wherever
4) Everything goes through a Pull Request

[Architecture](ARCHITECTURE.md)

## Layout

I'll get an architecture document up soon (:TM:) but for now, some common terms

* Devices -> Hardware devices, General Purpose Input Output (GPIO), temperature probes (DS18[B/S]20) to begin with, more to come
* PID -> [Proportional Integral Derivative](https://www.west-cs.com/products/l2/pid-temperature-controller/#:~:text=PID%20temperature%20controllers%20work%20using,possible%20by%20eliminating%20the%20impact)

Elsinore is designed first and foremost as a Brewery controller, in order to do that, it must be able to maintain temperatures in your brewing vessels (this will vary on your setup), so the core part of work is to have a regular poll that will update the current temperature from a probe and determine whether to turn on or off the output associated with that temperature probe.

For this reason, I chose go, for go routines, and I'm ensuring that I keep things well seperated between the UX and the backend.

[GraphQL](https://graphql.org/) is my API layer of choice, I find it much easier to deal with then REST (when you know what you are building)

## Building

This can be built for any platform you want, CI runs against Windows, Mac, and Linux builds, but not on RaspberryPi hardware (it should be the same as Linux, but I make no guarantees)

Currently, the scripts in this repo target Linux as the build host, since that's what I use as my primary development platform (I don't have a Windows machine, I'm open to pull requests)

`bin/build_all` -> This will build Elsinores backend/server side for Linux and RaspberryPi targets, placing the binaries in the `builds/<platform>` directories as appropriate.

## Running

`./elsinore` -> That's the basic startup. By default it will start on port *8080*, and serve a GraphQL API at */graphql* with a GraphiQL UI at */graphiql*

Options are

* `-port` -> Change the port to listen on
* `-graphiql` -> Turn off the GraphiQL interface (this may be turned off by default in the future)
* `-db_name` -> The path/name of the local database, this will default to your starting directory and `elsinore.db`
* `-test_device` -> A boolean flag to add a test Temperature probe, the physical address is `ARealAddress`

Note: Boolean options (true/false) must be set as `-graphiql=true`, this is due to shell restrictions. They can be `1, 0, t, f, T, F, true, false, TRUE, FALSE, True, False`

## Releasing

This is a long way away, I'll come up with something.

## WIP

More to come
