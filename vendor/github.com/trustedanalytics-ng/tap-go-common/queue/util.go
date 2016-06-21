/**
 * Copyright (c) 2016 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package queue

import (
	"fmt"

	"github.com/streadway/amqp"

	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
	"github.com/trustedanalytics-ng/tap-go-common/util"
)

var logger, _ = commonLogger.InitLogger("queue")

/*
Remember to call defer on outputs:
- defer conn
- defer ch
*/
func GetConnectionChannel() (*amqp.Channel, *amqp.Connection) {
	conn, err := amqp.Dial(getQueueConnectionString())
	if err != nil {
		logger.Panic("Failed to connect to Queue on address"+getQueueConnectionString(), err)
	}

	ch, err := conn.Channel()
	if err != nil {
		logger.Panic("Failed to open a channel", err)
	}
	return ch, conn
}

func getQueueConnectionString() string {
	queueNameInEnvs := "QUEUE"
	address, _ := util.GetConnectionAddressFromEnvs(queueNameInEnvs)
	user, pass, _ := util.GetConnectionCredentialsFromEnvs(queueNameInEnvs)

	return fmt.Sprintf("amqp://%v:%v@%v/", user, pass, address)
}

type ConsumeHandlerFunc func(msg amqp.Delivery)

func ConsumeMessages(ch *amqp.Channel, handler ConsumeHandlerFunc, queueName string, messagesTaken int) {
	terminationChannel := util.GetTerminationObserverChannel()

	err := ch.Qos(
		messagesTaken, // max number of messages taken
		0,             // max size of message
		false,         // global
	)
	if err != nil {
		logger.Fatalf("Failed to set qos on queue: %s, %v", queueName, err)
	}

	messageChannel, err := ch.Consume(
		queueName, //queue
		"",        // consumer - empty means generate unique id
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		logger.Fatalf("Failed to register consumer on queue: %s, %v", queueName, err)
	}

	for {
		select {
		case _ = <-terminationChannel:
			logger.Info(fmt.Sprintf("System is going down, proccessing messeges on queue: %s stopped!", queueName))
			return
		case message, ok := <-messageChannel:
			if ok {
				go handler(message)
			} else {
				logger.Fatalf(fmt.Sprintf("Consumer channel close unexpectedly! Queue: %s", queueName))
			}
		}
	}
}

func SendMessageToQueue(ch *amqp.Channel, message []byte, queueName, routingKey string) error {
	return ch.Publish(
		queueName,  // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		})
}

func CreateExchangeWithQueueByRoutingKeys(ch *amqp.Channel, name string, routingKeys []string) {
	err := ch.ExchangeDeclare(
		name,     // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		logger.Panic("Failed to declare an exchange", err)
	}

	q, err := ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		logger.Panic("Failed to declare a queue", err)
	}

	for _, key := range routingKeys {
		err = ch.QueueBind(
			q.Name, // queue name
			key,    // routing key
			name,   // exchange
			false,  // no-wait
			nil,    // arguments
		)
		if err != nil {
			logger.Panic("Failed to bind a queue", err)
		}
	}
}
