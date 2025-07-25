package clickhouse

import (
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

const (
	SQL_LOAD_ALERT_EVENTS_FROM_INCIDENTS = `SELECT
		i2a.incident_id,
		ae.*
	FROM incident2alert i2a
	INNER JOIN alert_event ae ON i2a.alert_event_id = ae.id
	%s`
)

type AlertEventWithIncidentID struct {
	IncidentID string `ch:"incident_id"`
	alert.AlertEvent
}

func (ch *chRepo) LoadAlertEventsFromIncidents(ctx core.Context, incidents []alert.Incident) ([]alert.Incident, error) {
	incidentsIDs := make([]string, len(incidents))

	var from, to int64
	for _, incident := range incidents {
		incidentsIDs = append(incidentsIDs, incident.ID)
		if from == 0 || incident.CreateTime < from {
			from = incident.CreateTime
		}

		if to == 0 || incident.UpdateTime > to {
			to = incident.UpdateTime
		}
	}

	qb := NewQueryBuilder().
		In("i2a.incident_id", incidentsIDs).
		Between("ae.update_time", from/1e6, to/1e6) // unixMicro -> unix

	sql := fmt.Sprintf(SQL_LOAD_ALERT_EVENTS_FROM_INCIDENTS, qb.String())

	var res []AlertEventWithIncidentID
	err := ch.GetConn(ctx).Select(ctx.GetContext(), &res, sql, qb.values...)
	if err != nil {
		return nil, err
	}

	for i, incident := range incidents {
		for _, event := range res {
			if event.IncidentID == incident.ID {
				incidents[i].AlertEvents = append(incidents[i].AlertEvents, event.AlertEvent)
			}
		}
	}
	return incidents, nil
}
