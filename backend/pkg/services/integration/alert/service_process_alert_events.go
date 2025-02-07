// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func (s *service) ProcessAlertEvents(source alert.SourceFrom, data []byte) error {
	events, err := s.dispatcher.DispatchEvents(&source, data)
	if err != nil {
		var errSourceNotExist alert.ErrAlertSourceNotExist

		// alertSource is not ready, could be undefined source
		if errors.As(err, &errSourceNotExist) {
			if len(source.SourceID) > 0 {
				// alertSource is not created yet, just dropped event
				return nil
			}

			// undefined alertSource, try to create default alertSource
			enricher, err := s.initDefaultAlertSource(&source)
			log.Printf("init default alertSource failed, err: %v", err)
			enricher.Enrich(events)
		}

		return fmt.Errorf("enrich alertEvent failed, err: %v", err)
	}

	return s.ckRepo.InsertAlertEvent(context.Background(), events, source)
}
