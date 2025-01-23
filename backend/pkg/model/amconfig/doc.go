// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package amconfig

// Configuration file for parsing and deleting alertmanager
// Body from github.com/prometheus/alertmanager@v0.27.0/config

/**
Modified:
-Change SecretURL to URL to avoid losing information in Marshal/Unmarshal
-Modify the JSON tag of all configurations to form a small hump
-Removes the logic of transferring different verification information to the header in the Validate, only checks
-Modify the JSON tag of other alarm methods other than EmailConfigs/WebhookConfigs to hide, and gradually open it later.
**/
