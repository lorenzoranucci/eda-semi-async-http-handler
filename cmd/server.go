package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/lorenzoranucci/eda-semi-async-http-handler/internal/application"
	http2 "github.com/lorenzoranucci/eda-semi-async-http-handler/internal/infrastructure/http"
	"github.com/lorenzoranucci/eda-semi-async-http-handler/internal/infrastructure/json"
	"github.com/lorenzoranucci/eda-semi-async-http-handler/pkg/async/kafka"
	kafka2 "github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverCmd = &cobra.Command{
	Use: "server",
	RunE: func(cmd *cobra.Command, args []string) error {
		brokers := viper.GetStringSlice("kafkaBrokers")

		w := &kafka2.Writer{
			Addr:     kafka2.TCP(brokers...),
			Topic:    "create_user_request_sent",
			Balancer: &kafka2.ReferenceHash{},
		}

		execContext, cf := context.WithCancel(context.Background())
		addObservers := make(chan kafka.Observer)
		removeObservers := make(chan kafka.Observer)
		arh := kafka.NewAsyncRequestHandler(
			w,
			&json.Serializer{},
			addObservers,
			removeObservers,
			execContext,
		)

		errorsChan := make(chan error)

		server := http.Server{Addr: fmt.Sprintf(":%s", viper.GetString("httpPort")), Handler: http2.NewCreateUserHandler(arh)}
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			err := server.ListenAndServe()
			if err != nil {
				errorsChan <- err
			}
			cf()
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			select {
			case <-execContext.Done():
				server.Shutdown(context.Background())
			}

			wg.Done()
		}()

		r := kafka2.NewReader(kafka2.ReaderConfig{
			Brokers: brokers,
			GroupID: "a7187475-1d4b-4ade-bdfa-f754aa381ea7",
			Topic:   "user",
		})

		notifier := kafka.NewObserversNotifier(r, &application.UserEventDeserializer{}, execContext)
		messages := make(chan kafka2.Message)

		wg.Add(1)
		go func() {
			notifier.NotifyObservers(addObservers, removeObservers, messages)
			cf()
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			notifier.ConsumeMessages(messages, errorsChan)
			cf()
			wg.Done()
		}()

		var errs []error
		wg2 := &sync.WaitGroup{}
		wg2.Add(1)
		go func() {
			for {
				select {
				case err, ok := <-errorsChan:
					if !ok {
						wg2.Done()
						return
					}

					if err != nil {
						errs = append(errs, err)
					}
				}
			}
		}()

		wg.Wait()
		close(errorsChan)
		wg2.Wait()
		log.Printf("errorsChan: %v", errs)

		return errors.Join(errs...)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
