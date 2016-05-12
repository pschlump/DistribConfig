Distributed Configuration: A library and tool to push configuration out to lots of clients
==========================================================================================

 [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/pschlump/Go-FTL/master/LICENSE)

Usage Scenario:   You have lots of people that can update sets of configuration.  These
people are at lots of locations.  What you want is for them to check in changes to
git (Could be github.com or a private git system) and then have the post-check-in 
git hook run something to update a lot of clients (think hundreds) on many different
machines.

This library and tool provides a consistent way of implementing changes to
configuration.

1. The CLI tool can be run on the configuration files to validate that they are syntactically correct.
2. The user can then use a "git commit" and a "git push origin master" to transfer the changed files to the server.
3. A git hook can run the CLI tool with different options to load the configuration files that have changed into Redis.
4. A Redis push message can take place or clients can pole a single 256-bit hash to find if anything has changed.
5. Clients can then receive changes.




