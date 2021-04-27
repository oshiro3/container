## Usage 

```
$ go build
```

## Call From Docker

```
$less /etc/docker/daemon.json
{
    "runtimes": {
        "runner": {
            "path": "$GOPATH/src/github.com/oshiro3/container/runtime/spec/spec"
        }
    }
}

$systemctl daemon-reload
#systemctl restart docker

$sudo docker run --runtime runner
```
