// @@@SNIPSTART temporal-ecommerce-worker
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cenkalti/backoff/v4"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"temporal-ecommerce/app"
)

const (
	EnvHostKey = "TEMPORAL_HOST"
	EnvPortKey = "TEMPORAL_PORT"
)

const (
	DefaultHost = "127.0.0.1"
	DefaultPort = "7233"
)

var (
	stripeKey     = os.Getenv("STRIPE_PRIVATE_KEY")
	mailgunDomain = os.Getenv("MAILGUN_DOMAIN")
	mailgunKey    = os.Getenv("MAILGUN_PRIVATE_KEY")
)

func main() {
	var err error

	host := os.Getenv(EnvHostKey)
	if host == "" {
		host = DefaultHost
	}

	port := os.Getenv(EnvPortKey)
	if port == "" {
		port = DefaultPort
	}

	hostPort := fmt.Sprintf("%s:%s", host, port)
	log.Printf("connecting to temporal: %s\n", hostPort)

	var temporalClient client.Client
	err = backoff.Retry(
		func() error {
			// Create the client object just once per process
			temporalClient, err = client.NewClient(
				client.Options{
					HostPort: hostPort,
				},
			)
			return err
		},
		backoff.NewExponentialBackOff(),
	)
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer temporalClient.Close()

	// This worker hosts both Worker and Activity functions
	w := worker.New(temporalClient, "CART_TASK_QUEUE", worker.Options{})

	a := &app.Activities{
		StripeKey:     stripeKey,
		MailgunDomain: mailgunDomain,
		MailgunKey:    mailgunKey,
	}

	w.RegisterActivity(a.CreateStripeCharge)
	w.RegisterActivity(a.SendAbandonedCartEmail)

	w.RegisterWorkflow(app.CartWorkflow)
	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}

// @@@SNIPEND
