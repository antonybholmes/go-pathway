module github.com/antonybholmes/go-pathway

go 1.23

replace github.com/antonybholmes/go-basemath => ../go-basemath

replace github.com/antonybholmes/go-sys => ../go-sys

replace github.com/antonybholmes/go-auth => ../go-auth

require (
	github.com/antonybholmes/go-basemath v0.0.0-20240825181410-a6174a39116c
	github.com/antonybholmes/go-sys v0.0.0-20240901191116-4f230479c4a8
	github.com/rs/zerolog v1.33.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/labstack/echo/v4 v4.12.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/matoous/go-nanoid/v2 v2.1.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/xyproto/randomstring v1.0.5 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/text v0.18.0 // indirect
)

require (
	github.com/antonybholmes/go-auth v0.0.0-00010101000000-000000000000
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.25.0 // indirect
)
