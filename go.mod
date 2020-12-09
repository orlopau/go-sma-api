module github.com/orlopau/go-sma-api

go 1.15

require (
	github.com/gorilla/mux v1.8.0
	github.com/gosuri/uilive v0.0.4
	github.com/olekukonko/tablewriter v0.0.4
	github.com/orlopau/go-energy v0.0.0-20201201231951-d8a9dd845376
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.7.1
	github.com/urfave/cli/v2 v2.3.0
	go.uber.org/goleak v1.1.10
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
)

// replace github.com/orlopau/go-energy => /home/paul/dev/go-energy
