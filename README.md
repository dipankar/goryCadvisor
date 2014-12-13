goryCadvisor
=============

Cadvisor and Riemann integration. It basically pulls data from cadvisor and pushes it into riemann.

To build the script use 

`go build main.go`

To set your own Cadvisor and Riemann ports

`./main -riemann_address="localhost:5555" -cadvisor_address="http://localhost:8080"`

Feel free to modify and add more datapoints to be pushed into Reimann!
