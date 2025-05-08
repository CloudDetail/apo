// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package receiver

import (
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/prometheus/alertmanager/template"
)

func (r *InnerReceivers) updateReceiversInMemory(receivers []amconfig.Receiver) error {
	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return err
	}
	tmpl.ExternalURL = r.externalURL

	newReceiver, err := buildInnerReceivers(receivers, tmpl, r.logger)
	if err != nil {
		return err
	}
	r.receivers = newReceiver.receivers
	return nil
}

func (r *InnerReceivers) GetAMConfigReceiver(filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]amconfig.Receiver, int) {
	receivers, count, err := r.database.GetAMConfigReceiver(filter, pageParam)
	if err != nil {
		r.logger.Error("failed to list amconfigReceiver", "err", err)
		return []amconfig.Receiver{}, 0
	}
	return receivers, count
}

func (r *InnerReceivers) AddAMConfigReceiver(receiver amconfig.Receiver) error {
	err := r.database.AddAMConfigReceiver(receiver)
	if err != nil {
		return err
	}

	receivers, _, err := r.database.GetAMConfigReceiver(nil, nil)
	if err != nil {
		return err
	}
	return r.updateReceiversInMemory(receivers)
}

func (r *InnerReceivers) UpdateAMConfigReceiver(receiver amconfig.Receiver, oldName string) error {
	err := r.database.UpdateAMConfigReceiver(receiver, oldName)
	if err != nil {
		return err
	}
	receivers, _, err := r.database.GetAMConfigReceiver(nil, nil)
	if err != nil {
		return err
	}
	return r.updateReceiversInMemory(receivers)
}

func (r *InnerReceivers) DeleteAMConfigReceiver(name string) error {
	err := r.database.DeleteAMConfigReceiver(name)
	if err != nil {
		return err
	}
	receivers, _, err := r.database.GetAMConfigReceiver(nil, nil)
	if err != nil {
		return err
	}
	return r.updateReceiversInMemory(receivers)
}
