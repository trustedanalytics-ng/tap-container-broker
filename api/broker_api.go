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

package api

import (
	"errors"
	"net/http"

	"github.com/gocraft/web"

	"github.com/trustedanalytics-ng/tap-container-broker/engine"
	"github.com/trustedanalytics-ng/tap-container-broker/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
)

type Context struct{}

var ServiceEngine engine.ServiceEngine
var logger, _ = commonLogger.InitLogger("api")

func (c *Context) CheckEngineService(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	if ServiceEngine == nil {
		commonHttp.Respond500(rw, errors.New("ServiceEngine not set!"))
	}
	next(rw, req)
}

func (c *Context) GetInstanceLogs(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["instanceId"]

	result, err := ServiceEngine.GetInstanceLogs(instanceId)
	if err != nil {
		commonHttp.HandleError(rw, err)
		return
	}
	commonHttp.WriteJson(rw, result, http.StatusOK)
}

func (c *Context) GetCredentials(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["instanceId"]

	result, err := ServiceEngine.GetCredentials(instanceId)
	if err != nil {
		commonHttp.HandleError(rw, err)
		return
	}
	commonHttp.WriteJson(rw, result, http.StatusOK)
}

func (c *Context) GetInstanceHosts(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["instanceId"]

	hosts, err := ServiceEngine.GetIngressHosts(instanceId)
	if err != nil {
		commonHttp.HandleError(rw, err)
		return
	}

	commonHttp.WriteJson(rw, hosts, http.StatusOK)
}

func (c *Context) GetSecret(rw web.ResponseWriter, req *web.Request) {
	secretName := req.PathParams["secretName"]

	secret, err := ServiceEngine.GetSecret(secretName)
	if err != nil {
		commonHttp.HandleError(rw, err)
		return
	}

	commonHttp.WriteJson(rw, secret, http.StatusOK)
}

func (c *Context) GetConfigMap(rw web.ResponseWriter, req *web.Request) {
	configMapName := req.PathParams["configMapName"]

	configMap, err := ServiceEngine.GetConfigMap(configMapName)
	if err != nil {
		commonHttp.HandleError(rw, err)
		return
	}

	commonHttp.WriteJson(rw, configMap, http.StatusOK)
}

func (c *Context) Bind(rw web.ResponseWriter, req *web.Request) {
	if message, status, err := ServiceEngine.ValidateAndLinkInstance(req.PathParams["srcInstanceId"], req.PathParams["dstInstanceId"], true); err != nil {
		logger.Errorf("InstanceID: %s, Bind error: %v", req.PathParams["dstInstanceId"], err)
		commonHttp.WriteJson(rw, message, status)
		return
	}
	commonHttp.WriteJson(rw, models.MessageResponse{Message: models.DeployResponseStatusSuccess}, http.StatusAccepted)

}

func (c *Context) Unbind(rw web.ResponseWriter, req *web.Request) {
	if message, status, err := ServiceEngine.ValidateAndLinkInstance(req.PathParams["srcInstanceId"], req.PathParams["dstInstanceId"], false); err != nil {
		logger.Errorf("InstanceID: %s, Bind error: %v", req.PathParams["dstInstanceId"], err)
		commonHttp.WriteJson(rw, message, status)
		return
	}
	commonHttp.WriteJson(rw, models.MessageResponse{Message: models.DeployResponseStatusSuccess}, http.StatusAccepted)
}

func (c *Context) Expose(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["instanceId"]

	request := models.ExposeRequest{}
	if err := commonHttp.ReadJson(req, &request); err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	hosts, err := ServiceEngine.Expose(instanceId, request)
	if err != nil {
		commonHttp.HandleError(rw, err)
	} else {
		commonHttp.WriteJson(rw, hosts, http.StatusOK)
	}
}

func (c *Context) Hide(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["instanceId"]

	hosts, err := ServiceEngine.Hide(instanceId)
	if err != nil {
		commonHttp.HandleError(rw, err)
	} else {
		commonHttp.WriteJson(rw, hosts, http.StatusNoContent)
	}
}

func (c *Context) GetContainerBrokerHealth(rw web.ResponseWriter, req *web.Request) {
	err := ServiceEngine.GetContainerBrokerHealth()
	if err != nil {
		commonHttp.WriteJson(rw, err, http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (c *Context) GetVersions(rw web.ResponseWriter, req *web.Request) {
	result, err := ServiceEngine.GetVersions()
	if err != nil {
		commonHttp.HandleError(rw, err)
		return
	}
	commonHttp.WriteJson(rw, result, http.StatusOK)
}

func (c *Context) Error(rw web.ResponseWriter, r *web.Request, err interface{}) {
	logger.Error("Respond500: reason: error ", err)
	rw.WriteHeader(http.StatusInternalServerError)
}
