package main

import (
	"./command"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"gopkg.in/urfave/cli.v1"
	"os"
	"strconv"
	"strings"
)

func main() {

	type Values struct {
		Name        string
		Description string
		Command     string
		Timeout     int
		Region      string
		InstanceId  string
		Output      string
		Return      string
		Success     string
	}
	values := &Values{}

	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "name, N, n", Usage: "The name of the job"},
		cli.StringFlag{Name: "description, D, d", Usage: "Description of what this job is for"},
		cli.StringFlag{Name: "region, R, r", Usage: "The AWS region where the event will be pushed", EnvVar: "AWS_REGION"},
		cli.IntFlag{Name: "timeout, T, t", Usage: "Amount of time before the command times out"},
		cli.BoolFlag{Name: "verbose, V", Usage: "Print verbose output"},
	}

	app.Name = "Cloudwatch Cron Wrapper"
	app.Version = "0.0.1"
	app.Usage = "Execute a command and send the result to cloudwatch events"
	app.Authors = []cli.Author{
		cli.Author{
			Name: "Aaron Cossey",
		},
	}
	app.Action = func(c *cli.Context) error {

		// Name
		if !c.IsSet("name") {
			cli.ShowAppHelp(c)
			return cli.NewExitError("Error: No check name specified", -1)
		} else {
			values.Name = c.String("name")
		}

		// Description
		if !c.IsSet("description") {
			values.Description = c.String("name")
		} else {
			values.Description = c.String("description")
		}

		// Region
		if !c.IsSet("region") {
			cli.ShowAppHelp(c)
			return cli.NewExitError("Error: No region specified and unable to determine", -1)
		} else {
			values.Region = c.String("region")
		}

		// Timeout
		if !c.IsSet("timeout") {
			values.Timeout = 0
		} else {
			values.Timeout = c.Int("timeout")
		}

		// InstanceID
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Config:            aws.Config{Region: aws.String(values.Region)},
		}))
		//Try to get InstanceID from ec2metadata
		svc := ec2metadata.New(sess)
		eiddoc, err := svc.GetInstanceIdentityDocument()
		if err != nil {
			// Unable to get ec2 instance id document, try just get a hostname
			instanceid, err := os.Hostname()
			if err != nil {
				return cli.NewExitError("Error: Unable to determine InstanceID or hostname", -1)
			}
			values.InstanceId = instanceid
		} else {
			values.InstanceId = eiddoc.InstanceID
		}

		//Command
		comm := c.Args()
		if len(comm) == 0 {
			cli.ShowAppHelp(c)
			return cli.NewExitError("Error: No command was given to run", -1)
		} else {
			values.Command = strings.Join(comm, " ")
		}

		//Run Command
		status, output := command.RunCommand(c.Args().First(), c.Args().Tail(), values.Timeout)
		//Limit bytes in output (cloudwatch limits event to 256KB)
		var out_lim_bytes int = 200000
		//Removing newlines
		line := strings.Replace(output, "\n", "", -1)
		if len(line) > out_lim_bytes {
			values.Output = line[:out_lim_bytes]
		} else {
			values.Output = line
		}
		values.Return = strconv.Itoa(status)
		if values.Return == "0" {
			values.Success = "true"
		} else {
			values.Success = "false"
		}

		//Prep data for creating cloudwatch event
		detail := fmt.Sprintf(
			"{ \"Name\":\"%s\",\"Command\":\"%s\",\"Output\":\"%s\",\"Return\":\"%s\",\"Success\":\"%s\"}",
			values.Name,
			values.Command,
			values.Output,
			values.Return,
			values.Success,
		)

		//Create the cloudwatch events client
		svc2 := cloudwatchevents.New(sess)
		result, err := svc2.PutEvents(&cloudwatchevents.PutEventsInput{
			Entries: []*cloudwatchevents.PutEventsRequestEntry{
				&cloudwatchevents.PutEventsRequestEntry{
					Detail:     aws.String(detail),
					DetailType: aws.String(values.Description),
					Resources: []*string{
						aws.String(values.InstanceId),
					},
					Source: aws.String("au.com.formbay.cloudwatch_wrapper"),
				},
			},
		})

		if err != nil {
			fmt.Println("Error", err)
			return err
		}

		if c.Bool("verbose") {
			fmt.Println("Success", result)
		}

		return nil
	}
	app.Run(os.Args)

}
