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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	brokerHttp "github.com/trustedanalytics-ng/tap-go-common/http"
	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
	"github.com/trustedanalytics-ng/tap-template-repository/model"
)

var logger, _ = commonLogger.InitLogger("catalog")

// 130 000 is almost 128k of characters
const MaxQueryParamsLength = 130000

type TemplateRepository interface {
	GenerateParsedTemplate(templateId, uuid, planName string, replacements map[string]string) (model.Template, int, error)
	CreateTemplate(template model.RawTemplate) (int, error)
	GetRawTemplate(templateId string) (model.RawTemplate, int, error)
	DeleteTemplate(templateId string) (int, error)
	GetTemplateRepositoryHealth() error
}

type TemplateRepositoryConnector struct {
	Address  string
	Username string
	Password string
	Client   *http.Client
}

func NewTemplateRepositoryBasicAuth(address, username, password string) (*TemplateRepositoryConnector, error) {
	client, _, err := brokerHttp.GetHttpClient()
	if err != nil {
		return nil, err
	}
	return &TemplateRepositoryConnector{Address: address, Username: username, Password: password, Client: client}, nil
}

func (t *TemplateRepositoryConnector) GenerateParsedTemplate(templateId, uuid, planName string,
	replacements map[string]string) (model.Template, int, error) {

	template := model.Template{}

	address := fmt.Sprintf("%s/api/v1/parsed_template/%s?instanceId=%s&planName=%s", t.Address, templateId, uuid, planName)
	if len(replacements) > 0 {
		params := url.Values{}
		for key, value := range replacements {
			if len(params.Encode()+url.QueryEscape(key)+url.QueryEscape(value)) > MaxQueryParamsLength {
				logger.Warning("Replacements are too big - some of them will be omitted!")
				break
			}
			params.Add(key, value)
		}
		address = fmt.Sprintf("%s&%s", address, params.Encode())
	}

	auth := brokerHttp.BasicAuth{User: t.Username, Password: t.Password}
	status, body, err := brokerHttp.RestGET(address, brokerHttp.GetBasicAuthHeader(&auth), t.Client)
	if err != nil {
		return template, status, err
	}
	err = json.Unmarshal(body, &template)
	if err != nil {
		return template, http.StatusInternalServerError, err
	}
	if status != http.StatusOK {
		return template, status, errors.New("Bad response status: " + strconv.Itoa(status) + ". Body: " + string(body))
	}
	return template, http.StatusOK, nil
}

func (t *TemplateRepositoryConnector) CreateTemplate(template model.RawTemplate) (int, error) {
	url := fmt.Sprintf("%s/api/v1/templates", t.Address)

	b, err := json.Marshal(&template)
	if err != nil {
		return 400, err
	}

	auth := brokerHttp.BasicAuth{User: t.Username, Password: t.Password}
	status, _, err := brokerHttp.RestPOST(url, string(b), brokerHttp.GetBasicAuthHeader(&auth), t.Client)
	if err != nil {
		return status, err
	}
	if status != http.StatusCreated {
		return status, errors.New("Bad response status: " + strconv.Itoa(status))
	}
	return status, nil
}

func (t *TemplateRepositoryConnector) GetRawTemplate(templateId string) (model.RawTemplate, int, error) {
	rawTemplate := model.RawTemplate{}

	url := fmt.Sprintf("%s/api/v1/templates/%s", t.Address, templateId)
	auth := brokerHttp.BasicAuth{User: t.Username, Password: t.Password}
	status, body, err := brokerHttp.RestGET(url, brokerHttp.GetBasicAuthHeader(&auth), t.Client)
	if err != nil {
		return rawTemplate, status, err
	} else if status != http.StatusOK {
		return rawTemplate, status, errors.New("Bad response status: " + strconv.Itoa(status))
	}

	if err = json.Unmarshal(body, &rawTemplate); err != nil {
		return rawTemplate, http.StatusInternalServerError, err
	}

	return rawTemplate, http.StatusOK, nil
}

func (t *TemplateRepositoryConnector) DeleteTemplate(templateId string) (int, error) {
	url := fmt.Sprintf("%s/api/v1/templates/%s", t.Address, templateId)

	auth := brokerHttp.BasicAuth{User: t.Username, Password: t.Password}
	status, _, err := brokerHttp.RestDELETE(url, "", brokerHttp.GetBasicAuthHeader(&auth), t.Client)
	if err != nil {
		return status, err
	}
	if status != http.StatusNoContent {
		return status, errors.New("Bad response status: " + strconv.Itoa(status))
	}
	return status, nil
}

func (t *TemplateRepositoryConnector) GetTemplateRepositoryHealth() error {
	url := fmt.Sprintf("%s/healthz", t.Address)

	auth := brokerHttp.BasicAuth{User: t.Username, Password: t.Password}
	status, _, err := brokerHttp.RestGET(url, brokerHttp.GetBasicAuthHeader(&auth), t.Client)
	if status != http.StatusOK {
		err = errors.New("Invalid health status: " + strconv.Itoa(status))
	}
	return err
}
