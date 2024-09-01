module github.com/antonybholmes/go-pathway

go 1.23

replace github.com/antonybholmes/go-basemath => ../go-basemath

replace github.com/antonybholmes/go-sys => ../go-sys

require (
	github.com/antonybholmes/go-basemath v0.0.0-20240825181410-a6174a39116c
	github.com/antonybholmes/go-sys v0.0.0-20240901041129-6c570bd0bacc
	github.com/rs/zerolog v1.33.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.24.0 // indirect
)
