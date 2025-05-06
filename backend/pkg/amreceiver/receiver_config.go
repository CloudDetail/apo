package amreceiver

import (
	"net/url"

	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/prometheus/alertmanager/template"
)

func (r *InnerReceivers) updateReceiversInMemory(receivers []amconfig.Receiver) error {
	tmpl, err := template.FromGlobs([]string{})
	if err != nil {
		return err
	}
	tmpl.ExternalURL, err = url.Parse(r.externalURL)
	if err != nil {
		return err
	}

	newReceiver, err := buildInnerReceivers(receivers, tmpl, r.logger)
	if err != nil {
		return err
	}
	r.receivers = newReceiver.receivers
	return nil
}

func (r *InnerReceivers) GetAMConfigReceiver(filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]amconfig.Receiver, int) {
	return r.database.GetAMConfigReceiver(filter, pageParam)
}

func (r *InnerReceivers) AddAMConfigReceiver(receiver amconfig.Receiver) error {
	err := r.database.AddAMConfigReceiver(receiver)
	if err != nil {
		return err
	}

	receivers, _ := r.database.GetAMConfigReceiver(nil, nil)
	return r.updateReceiversInMemory(receivers)
}

func (r *InnerReceivers) UpdateAMConfigReceiver(receiver amconfig.Receiver, oldName string) error {
	err := r.database.UpdateAMConfigReceiver(receiver, oldName)
	if err != nil {
		return err
	}
	receivers, _ := r.database.GetAMConfigReceiver(nil, nil)
	return r.updateReceiversInMemory(receivers)
}

func (r *InnerReceivers) DeleteAMConfigReceiver(name string) error {
	err := r.database.DeleteAMConfigReceiver(name)
	if err != nil {
		return err
	}
	receivers, _ := r.database.GetAMConfigReceiver(nil, nil)
	return r.updateReceiversInMemory(receivers)
}
