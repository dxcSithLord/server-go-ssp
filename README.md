# sqrl-ssp #
SQRL is a identiy managment system that is meant to replace usernames and passwords for online
account authentication. It requires a user to have a SQRL client that securely manages their 
identity. The server interacts with the SQRL client to authenticate a user (similar but more 
secure than a username/password challenge). Once a user's identity is established, a session
should be established if the desired behavior is for a user to remain "logged in". This is typically
a session cookie or authentication token.

This implements the public parts of the SQRL authentication server API as specified here: https://www.grc.com/sqrl/sspapi.htm.
This library is meant to be pluggable into a broader infrastructure to handle whatever type
of session management you desire. It also allows pluggable storage options and scales horizontally.

This project is still very much a work in-progress. All the endpoints log a ton of debugging information.

[![Documentation](https://godoc.org/github.com/smw1218/sqrl-ssp?status.svg)](https://godoc.org/github.com/smw1218/sqrl-ssp)
[![Go Report Card](https://goreportcard.com/badge/github.com/smw1218/sqrl-ssp)](https://goreportcard.com/report/github.com/smw1218/sqrl-ssp)

## Integration ##
The ssp.SqrlSspAPI struct is a configurable server that exposes http.HandlerFuncs that implement the SSP API. The main one
is the ssp.SqrlSspAPI.Cli handler which directly handles communication from the SQRL client. These endpoints can be configured to 
run as a standalone service or as part of a larger API structure. There are several required pieces
of configuration that must be provided to integrate SQRL into broader user management. 

### ssp.Authenticator ###
The basis of the SSP API is to manage SQRL identities. The goal of this library is to manage these identities and allow
for loosly coupling an identity to a "user". This is similar in concept to a user having a username and password which may be
changed for a given user. A SQRL idenity can be associated with a user, and at a later time that identity may be disabled or
removed from a user, or a new identity may be associated with that user. These actions are supported by the ssp.Authenticator
interface.

### Hoard and AuthStore ##
The SSP API has requirements for storage exposed by the Hoard and AuthStore interfaces. Because an extended pun is always fun, a Hoard stores Nuts
Nuts are SQRL's cryptographic nonces. A Hoard also has stores pending auth information associated with the Nut. These are ephemperal and have an
expiration so are best stored in a in-memory store like Redis or memcached. The AuthStore saves the SQRL identity information and should be a durable database like PostgreSQL or MariaDB. Both are interfaces so any storage should be able to be plugged in. The ssp package provides map-backed implementations for both which are *NOT* recommended for production use. 

I've written a Redis-backed Hoard implementation at [github.com/smw1218/sqrl-redishoard](https://github.com/smw1218/sqrl-redishoard)

### Trees ###
Trees produce Nuts. There are several ways to produce a secure nonce. GRC reccommends an in-memory counter-based nonce, but the design
does not easily scale horizontally. Multiple servers could produce the same nonce if they are not externally coordinated (like through
a globally consistent counter like a PostgreSQL sequence.) The ssp package provides ssp.GrcTree as an implementation of this, but I 
reccommend using ssp.RandomTree if you're using multiple servers.
