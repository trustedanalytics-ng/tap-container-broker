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
package client

import (
	"fmt"
	"net/http"

	"github.com/trustedanalytics-ng/tap-catalog/models"
	brokerHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

func (c *TapCatalogApiConnector) AddImage(image models.Image) (models.Image, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s", c.Address, images))
	result := &models.Image{}
	status, err := brokerHttp.PostModel(connector, image, http.StatusCreated, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) GetImage(imageId string) (models.Image, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, images, imageId))
	result := &models.Image{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) GetImageRefs(imageId string) (models.ImageRefsResponse, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s/%s", c.Address, images, imageId, checkRefs))
	result := &models.ImageRefsResponse{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) ListImages() ([]models.Image, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s", c.Address, images))
	result := &[]models.Image{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) UpdateImage(imageId string, patches []models.Patch) (models.Image, int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, images, imageId))
	result := &models.Image{}
	status, err := brokerHttp.PatchModel(connector, patches, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) DeleteImage(imageId string) (int, error) {
	connector := c.getApiConnector(fmt.Sprintf("%s/%s/%s", c.Address, images, imageId))
	return brokerHttp.DeleteModel(connector, http.StatusNoContent)
}

func (c *TapCatalogApiConnector) WatchImages(afterIndex uint64) (models.StateChange, int, error) {
	connector := c.getWatchApiConnector(fmt.Sprintf("%s/%s/%s?afterIndex=%d", c.Address, images, nextState, afterIndex))
	result := &models.StateChange{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}

func (c *TapCatalogApiConnector) WatchImage(imageId string, afterIndex uint64) (models.StateChange, int, error) {
	connector := c.getWatchApiConnector(fmt.Sprintf("%s/%s/%s/%s?afterIndex=%d", c.Address, images, imageId, nextState, afterIndex))
	result := &models.StateChange{}
	status, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, status, err
}
