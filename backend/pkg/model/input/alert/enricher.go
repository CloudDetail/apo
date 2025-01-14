// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "github.com/google/uuid"

type AlertEnricherConfig struct {
	SourceID   uuid.UUID `gorm:"primaryKey"`
	EnricherID uint      `gorm:"primaryKey"`
}

type AlertEnricher struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(100);column:name"`
	Type string `gorm:"type:varchar(100);column:type"` // apo,static
}

type AlertEnricherVO struct {
	SourceID uuid.UUID

	EnricherID   uint   `gorm:"column:enricher_id"`
	EnricherName string `gorm:"type:varchar(100);column:enricher_name"`
	EnricherType string `gorm:"type:varchar(100);column:enricher_type"` // apo,static
}
