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

package adapter

import (
	"sync"

	"github.com/streadway/amqp"

	"github.com/trustedanalytics-ng/tap-container-broker/engine"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
	"github.com/trustedanalytics-ng/tap-go-common/queue"
)

var ServiceEngine engine.ServiceEngine
var logger, _ = commonLogger.InitLogger("adapter")

const (
	maxConsumerRoutines = 20
)

func StartConsumer(waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)

	ch, conn := queue.GetConnectionChannel()
	defer conn.Close()
	defer ch.Close()

	queue.CreateExchangeWithQueueByRoutingKeys(ch, models.CONTAINER_BROKER_QUEUE_NAME,
		[]string{
			models.CONTAINER_BROKER_CREATE_ROUTING_KEY,
			models.CONTAINER_BROKER_DELETE_ROUTING_KEY,
			models.CONTAINER_BROKER_SCALE_ROUTING_KEY,
		})
	queue.ConsumeMessages(ch, handleMessage, models.CONTAINER_BROKER_QUEUE_NAME, maxConsumerRoutines)
	waitGroup.Done()
}

func handleMessage(msg amqp.Delivery) {
	defer msg.Ack(false)

	switch msg.RoutingKey {
	case models.CONTAINER_BROKER_CREATE_ROUTING_KEY:
		createInstanceRequest := models.CreateInstanceRequest{}
		if err := commonHttp.ReadJsonFromByte(msg.Body, &createInstanceRequest); err != nil {
			logger.Error("Cannot unmarshal CreateInstanceRequest!", err)
			return
		}

		logger.Info("Received message! Create instance: " + string(createInstanceRequest.InstanceId))

		if err := ServiceEngine.CreateKubernetesInstance(createInstanceRequest); err != nil {
			logger.Errorf("Cannot create K8S instance from queue task! InstanceId: %q Error: %q", createInstanceRequest.InstanceId, err)
			return
		}
	case models.CONTAINER_BROKER_DELETE_ROUTING_KEY:
		deleteRequest := models.DeleteRequest{}
		if err := commonHttp.ReadJsonFromByte(msg.Body, &deleteRequest); err != nil {
			logger.Error("Cannot unmarshal DeleteRequest!", err)
			return
		}

		logger.Info("Received message! Delete instance: " + deleteRequest.Id)

		if err := ServiceEngine.DeleteKubernetesInstance(deleteRequest.Id); err != nil {
			logger.Error("Cannot delete K8S instance from queue task! InstanceId:", deleteRequest.Id, err)
			return
		}
	case models.CONTAINER_BROKER_SCALE_ROUTING_KEY:
		scaleRequest := models.ScaleInstanceRequest{}
		if err := commonHttp.ReadJsonFromByte(msg.Body, &scaleRequest); err != nil {
			logger.Error("Cannot unmarshal ScaleInstanceRequest:", err)
			return
		}

		logger.Info("Received message! Scale instance: " + scaleRequest.Id)

		if err := ServiceEngine.ScaleInstance(scaleRequest.Id); err != nil {
			logger.Errorf("Cannot scale instance from queue task! InstanceId: %s, error: %v", scaleRequest.Id, err)
			return
		}
	default:
		logger.Error("Routing key not supported: ", msg.RoutingKey)
	}
}
