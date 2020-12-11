module meow.tf/crewlink-server

go 1.15

replace github.com/ambelovsky/gosf-socketio => ../gosf-socketio

require (
	github.com/go-chi/chi v1.5.1
	github.com/gobuffalo/packr/v2 v2.8.1
	github.com/tystuyfzand/gosf-socketio v0.0.0-20201211021827-4134a4acb784
)
