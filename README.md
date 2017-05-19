# Rpounder: stress testing for DNS  resolvers

Rpounder is essentially apache bench for DNS resolvers. To get help, just run `rpounder --help`

    # rpounder --help
	Usage of bin/rpounder:
	  -c int
			Concurrent processes (default 10)
	  -n string
			Hostname(s) to look up. Separate multiple hostnames with spaces or commas (default "localhost")
	  -p int
			Number of passes (default 100)
	  -r string
			resolver(s) ip/port. Separate multiple resolvers with spaces or commas (default "localhost:53")

Rpounder will run repeated host look-ups against the specified host(s) and return precentile rankings for request times
and other summary information. Here we're running 5 concurrent processes sending a total of 1000 requests to the resolver at 192.168.0.1 (default port 53) to resolve "foo.example.com":

	# rpounder -p 1000 -c 5 -r "192.168.0.1" -n "foo.example.com"

	2017/05/15 16:00:39 rpounder resolver benchmark 1.0.0 starting
	2017/05/15 16:00:44 Percentage of the requests served within a certain time:
	2017/05/15 16:00:44 50%  => 12.574632ms
	2017/05/15 16:00:44 66%  => 14.003579ms
	2017/05/15 16:00:44 75%  => 14.529408ms
	2017/05/15 16:00:44 80%  => 15.128422ms
	2017/05/15 16:00:44 90%  => 16.5929ms
	2017/05/15 16:00:44 95%  => 19.918966ms
	2017/05/15 16:00:44 99%  => 32.612096ms
	2017/05/15 16:00:44 Fastest: 9.008518ms  -- Slowest: 2.003962086s
	2017/05/15 16:00:44 Total passes: 1000. Total errors: 4
	
Multiple resolvers or hosts can be separated by spaces and/or commas. Specify an alternate port by appending a colon and a port number to the resolver spec.

## Installation

You can grab the latest release from the [release page](https://github.com/mowings/rpounder/releases), unarchive,  and merely copy the correct binary somewhere on your path. There are no dependencies.

Alternatively, if you already have go installed you can run:

```shell
go get github.com/mowings/rpounder/src/rpounder
```
per usual, which will build and drop the binaries in your go bin directory.


## Building it

You'll need go 1.7 or better. I use [gb](https://getgb.io/) to build, in which case you can simply change to the project directory and run `gb build`. 

If you prefer to use just go, change to the project directory and run

    export GOPATH=$GOPATH:`pwd`:`pwd`/vendor
    
Once you've done that, you can use `go build` in the usual way to build the executable. 

For both build methods, set your desired os and architecture via environment variables GOOS and GOARCH to build for your desired platform. 
The release provides binaries for OSX, Linux and Windows, but you can build for any supported Go platform/architecture. 
