# chirpy
Twitter simulator in Go. Practice project for working with SQL integration in Go and creating a working server with different endpoints.

# requirements
## postgresql
postgresql 14+

## goose
go install github.com/pressly/goose/v3/cmd/goose@latest
`cd sql/schema`
`goose postgres <connection_string> up`

## .env
Needs four variables:
- DB_URL: connection string with ?sslmode=disable query parameter. E.g. if running locally with default settings: postgres://postgres:postgres@localhost:5432/chirpy
- PLATFORM: when set to "dev" will enable admin endpoints. Otherwise does nothign a.t.m.
- SECRET: required string for authentication.
- POLKA_KEY: for webhook shenanigans.

# usage
## endpoints
- POST /api/users: takes `email` and `password` strings in JSON to create a new user in database. Email must be unique.
- PUT /api/users: takes `email` and `password` strings in JSON and updates the user in database based on access token.
- POST /api/login: takes `email` and `password` strings in JSON and provides client with an access and a refresh token. Access token lasts 1 hour, refresh token lasts 60 days.
- POST /api/refresh: provides user with a new access token if the current refresh token is still valid.
- POST /api/revoke: revokes the client's current refresh token. This effectively logs them out of the service.
- POST /api/chirps: takes `body` string in JSON and adds a Chirp to the database based on the client's access token. 
- GET /api/chirps: returns all Chirps in the database. Can be specified to /api/chirps/{chirpID} to only return a single Chirp based on Chirp ID. Two query parameters: `authorid` takes a UUID in string format to only return Chirps that were POSTed by the user with that UUID; `sort` sorts in either `asc`ending or `desc`ending order based on creation timestamp.
- DELETE /api/chirps/{chirpID}: deletes specific Chirp based on UUID in {chirpID}. If the entire database is to be wiped, please use /admin/reset instead (requires PLATFORM variable to be set to "dev" in .env.)