
## Requirements
```
go get -u github.com/golang-jwt/jwt/v4
go get -u github.com/gin-gonic/gin
go get -u github.com/swaggo/swag/cmd/swag

go get -u github.com/golang-jwt/jwt/v4@latest
go get -u github.com/gin-gonic/gin@latest
go get -u github.com/swaggo/swag/cmd/swag@latest

go get -u golang.org/x/crypto/bcrypt
go get -u github.com/oklog/ulid/v2
go get -u github.com/jackc/pgx/v4
go get -u github.com/georgysavva/scany
```

# Swagger
Command to generate a docs for Swagger
```
swag init
```

# Postgresql
```
postgresql://intep:01FKX5PMY5XCG64JNEBHEJST6J@localhost:5432/postgres
```
* [Go driver and toolkit](https://github.com/jackc/pgx)
* [New UUID format](https://github.com/uuid6/uuid6-ietf-draft)
* [Go ulid](https://github.com/oklog/ulid)


# Run
```
http://localhost:9000/swagger/index.html
```