// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// AddUser implements DifyRepo.
func (d *difyRepo) AddUser(username string, password string, role string) (*DifyResponse, error) {
	url := d.url + DIFY_ADD_USER

	req := &DifyUser{
		Password: password,
		Role:     role,
		Username: username,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := d.cli.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res DifyResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// RemoveUser implements DifyRepo.
func (d *difyRepo) RemoveUser(username string) (*DifyResponse, error) {
	url := d.url + DIFY_REMOVE_USER + username

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res DifyResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// UpdatePassword implements DifyRepo.
func (d *difyRepo) UpdatePassword(username string, oldPassword string, newPassword string) (*DifyResponse, error) {
	url := d.url + DIFY_PASSWORD_UPDATE

	req := &DifyUser{
		Password:    oldPassword,
		NewPassword: newPassword,
		Username:    username,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := d.cli.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res DifyResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (d *difyRepo) ResetPassword(username string, newPassword string) (*DifyResponse, error) {
	url := d.url + DIFY_RESET_PASSWORD

	req := &DifyUser{
		NewPassword: newPassword,
		Username:    username,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := d.cli.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res DifyResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
