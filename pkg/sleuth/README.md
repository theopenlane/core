# Sleuth

Package sleuth is a set of tools for performing DNS enumeration, http probing, and other reconnaissance tasks. It is designed to be used as a library and contains wrappers around several [Project Discovery](https://github.com/projectdiscovery) for ease of use as a library.

## Structure

The vast majority of projectdisocvery's tooling are CLI's and geared towards being run from the command line and using pipe magic to chain the tools together. Sleuth is attempting to wrap these tools into libraries with functional options parameters, client initialization (for scale-out / performance, etc.), and eventually cleaner report output and statistics.

Each subpackage was written with these general ideas in mind:
- create a wrapper struct for each client, Options struct for the options parameters, and Options type definitions and With functions
- create "new" functions for both the options and tool
- use "models" or mapping structs for data types that go in between what the upstream package is doing and what we want to do
- basic `test` directory with a `main` func that imports the subpackage and demonstrates basic use
- a `README.md` file that describes the tool, how to use it, and any other relevant information

There are additionally some nice to have printer functions in some cases to have pretty console output (rather than the spew of json that will eventually go into a schema)

## Next steps

The scope of the initial PR is to just lay out the package structure and basic functionality. The next steps are to add the following:
- add tests with ideally benchmarks
- evaluate moving this package out of this repo and into a dedicated one
- add riverqueue job launching leveraging the package
- functionality to add the results / reports of the subpackages into our schemas