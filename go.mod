module github.com/antonybholmes/go-pathway

go 1.22.5

replace github.com/antonybholmes/go-basemath => ../go-basemath

replace github.com/antonybholmes/go-sys => ../go-sys

require (
	github.com/antonybholmes/go-basemath v0.0.0-20240215163921-12bb7e52185c
	github.com/antonybholmes/go-sys v0.0.0-20240505052557-9f8864ac77aa
	github.com/rs/zerolog v1.33.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.22.0 // indirect
)
