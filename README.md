goryCadvisor
=============

Cadvisor and Riemann integration. It basically pulls data from cadvisor and pushes it into riemann.

To build the script use 

```
go get github.com/bigdatadev/goryman
go get github.com/golang/glog
go get github.com/google/cadvisor/client
go get github.com/google/cadvisor/info
go build main.go
```

To set your own Cadvisor and Riemann ports, with interval

`./main -riemann_address="localhost:5555" -cadvisor_address="http://localhost:8080" -interval="5s"`

Feel free to modify and add more datapoints to be pushed into Reimann!


## Running with docker 

```
docker run \
    -e RIEMANN_ADDRESS=<ipOfRiemann>:5555 \
    -e CADVISOR_ADDRESS=<ipOfCAdvisor>:8080 \
    -e INTERVAL=5s \
    docktor/gorycadvisor:latest
```
(image based on https://github.com/docktor/goryCadvisor)
