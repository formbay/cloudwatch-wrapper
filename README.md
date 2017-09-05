# Cloudwatch Wrapper

## Description

This go binary wraps around shell commands and send the results to AWS Cloudwatch as an Event

It is inpired by [@jaxxstorm](https://github.com/jaxxstorm)'s [Sensu Wrapper](https://github.com/jaxxstorm/sensu-wrapper).

## Usage

```
NAME:
   Cloudwatch Wrapper - Execute a command and send the result to cloudwatch events

USAGE:
   cloudwatch-wapper [global options] command [command options] [arguments...]

VERSION:
   0.0.1

AUTHOR:
   Aaron Cossey

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --name value, -N value, -n value         The name of the cron
   --description value, -D value, -d value  Description of what this cron is for (default: name)
   --region value, -R value, -r value       The AWS region where the event will be pushed [$AWS_REGION]
   --timeout value, -T value, -t value      Amount of time before the command times out (default: 0)
   --verbose, -V                            Print verbose output
   --help, -h                               show help
   --version, -v                            print the version
```

**Basic Example**

The minimum required are the name, region and command. The command isn't a flag, it will just take any arguments from the invocation.

This will run the command and send info about the result to Cloudwatch Events.

```
$ cloudwatch-wapper -n "testing" -d "testing the wrapper" -r "ap-southeast-2" /bin/echo hello
```

**Timeout**

By default, commands will continue to run until they finish. If you wish to adjust that, specify the timeout flag:

```
$ cloudwatch-wapper -n "testing" -d "testing the wrapper" -r "ap-southeast-2" -t 25 ping 8.8.8.8
```

If a command is killed due to timeout, it's assumed to have failed and will return exit code -1

## Building

Make sure your $GOPATH is set: https://github.com/golang/go/wiki/GOPATH Grab the external dependencies:
```
go get gopkg.in/urfave/cli.v1
go get -u github.com/aws/aws-sdk-go/...
```
Build it!
```
go build cloudwatch-wapper.go
```
That's it!

## IAM Authentication

If running on an ec2 instance with an IAM role, the role needs to have at least:
```
"events:PutEvents"
```
You can also use shared credentials (~/.aws/credentials) for local development and testing.

More info about credentials can be found here: https://github.com/aws/aws-sdk-go#configuring-credentials

You will need to specify the region, either with:
```
$ cloudwatch-wapper -r <region>
```
or in the environment with
```
$ AWS_REGION=<region> cloudwatch-wapper ...
```


