package cmd

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/lorenzoranucci/eda-semi-async-http-handler/internal/application"
	"github.com/lorenzoranucci/eda-semi-async-http-handler/internal/infrastructure/json"
	"github.com/lorenzoranucci/eda-semi-async-http-handler/internal/infrastructure/kafka"
	"github.com/lorenzoranucci/eda-semi-async-http-handler/internal/infrastructure/mysql"
	kafka2 "github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/go-sql-driver/mysql"
)

var consumerCmd = &cobra.Command{
	Use: "consumer",
	RunE: func(cmd *cobra.Command, args []string) error {
		brokers := viper.GetStringSlice("kafkaBrokers")
		r := kafka2.NewReader(kafka2.ReaderConfig{
			Brokers: brokers,
			GroupID: "777a3202-85b5-4c1e-ad37-d2aa411bc8b6",
			Topic:   "create_user_request_sent",
		})

		db, err := getSqlxDB()
		if err != nil {
			return fmt.Errorf("error while connecting to database: %w", err)
		}

		w := &kafka2.Writer{
			Addr:     kafka2.TCP(brokers...),
			Topic:    "user",
			Balancer: &kafka2.ReferenceHash{},
		}

		h := application.NewCreateUserCommandHandler(
			mysql.NewUserRepository(db),
			kafka.NewUserEventsPublisher(w, &json.Serializer{}),
		)

		err = kafka.NewCreateUserRequestSentEventConsumer(r, h, &json.Deserializer[application.CreateUserRequestSentEvent]{}).ConsumeMessages()
		if err != nil {
			log.Printf("error while consuming messages: %v", err)
		}

		return err
	},
}

func getSqlxDB() (*sqlx.DB, error) {
	return sqlx.Connect("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/app",
		viper.GetString("dbUser"),
		viper.GetString("dbPassword"),
		viper.GetString("dbHost"),
		viper.GetString("dbPort"),
	))
}

func init() {
	rootCmd.AddCommand(consumerCmd)
}
