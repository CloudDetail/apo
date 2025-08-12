package dataplane

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	CheckDatasourceAPI = "/datasource/checkProvider"
)

func (repo *dataplaneRepo) CheckDataSource(request *model.CheckDataSourceRequest) (*model.CheckDataSourceResponse, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("request is invalid: %s", err.Error())
	}
	resp, err := repo.client.Post(
		fmt.Sprintf("%s%s", repo.address, CheckDatasourceAPI),
		"application/json",
		bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("query api failed: %s", err.Error())
	}
	defer resp.Body.Close()

	response := &model.CheckDataSourceResponse{}
	if err = json.NewDecoder(resp.Body).Decode(response); err != nil {
		return nil, fmt.Errorf("parse response failed: %s", err.Error())
	}
	return response, nil
}
