package prometheus

import (
	"fmt"
	"strings"
	"time"
)

type MutationPQLCheck struct {
	PQL        string
	UpperLimit string
	LowerLimit string
}

func (c *MutationPQLCheck) GetExecutedPQL() string {
	withComparison := []string{}
	if len(c.UpperLimit) > 0 {
		withComparison = append(withComparison, fmt.Sprintf("((%s) > (%s))", c.PQL, c.UpperLimit))
	}
	if len(c.LowerLimit) > 0 {
		withComparison = append(withComparison, fmt.Sprintf("((%s) < (%s))", c.PQL, c.LowerLimit))
	}

	var pql string
	if len(withComparison) > 0 {
		pql = strings.Join(withComparison, " or ")
	} else {
		pql = c.PQL
	}
	return pql
}

func (r *promRepo) ExecutedMutationCheck(c *MutationPQLCheck, startTime, endTime time.Time, step time.Duration) ([]MetricResult, error) {
	pql := c.GetExecutedPQL()
	return r.QueryRangeData(startTime, endTime, pql, step)
}
