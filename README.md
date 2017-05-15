# Rpounder -- stress testing for DNS  resolvers

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
and other summary informeation:

	# rpounder -p 1000 -c 5 -r "192.168.0.1" -n "foo.example.com"

	2017/05/15 16:00:39 rpounder resolver benchmark 1.0.0 starting
	2017/05/15 16:00:44 Percentage of the requests served within a certain time:
	2017/05/15 16:00:44 50th  => 12.574632ms
	2017/05/15 16:00:44 66th  => 14.003579ms
	2017/05/15 16:00:44 75th  => 14.529408ms
	2017/05/15 16:00:44 80th  => 15.128422ms
	2017/05/15 16:00:44 90th  => 16.5929ms
	2017/05/15 16:00:44 95th  => 19.918966ms
	2017/05/15 16:00:44 99th  => 32.612096ms
	2017/05/15 16:00:44 Fastest: 9.008518ms  -- Slowest: 2.003962086s
	2017/05/15 16:00:44 Total passes: 1000. Total errors: 4

## Installation

You can grab the latest release under releases, unarchive and just copy the correct binary somewhere on your path. There are no dependencies.

## Building it

You'll need go 1.7 or better. I use [gb](https://getgb.io/) to build, in which case you can simply change to the project directory and run `gb build`. If you prefer to use just go, just change to the project directory and run

    export GOPATH=$GOPATH:`pwd`:`pwd`/vendor
    
Once you've done that, you can use `go build` in the usual way to build the executable. 
