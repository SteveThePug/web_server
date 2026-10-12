package graph

import (
	"fmt"

	"adam-french.co.uk/backend/models"
)

// setLeadState moves a lead from one state to another and returns it. It
// lives here, not in job_lead.resolvers.go, because gqlgen moves anything in
// a resolvers file that is not a resolver to the bottom on regeneration.
//
// The move is refused unless the lead is in `from`, which is what stops a
// promoted lead being dismissed or restored: its application exists, and
// flipping the lead would only make the inbox disagree with the tracker.
func (r *mutationResolver) setLeadState(id int, from, to string) (*models.JobLead, error) {
	var lead models.JobLead
	if err := r.Store.DB.First(&lead, id).Error; err != nil {
		return nil, err
	}
	if lead.State == to {
		return &lead, nil
	}
	if lead.State != from {
		return nil, fmt.Errorf("a %s lead cannot be made %s", lead.State, to)
	}
	if err := r.Store.DB.Model(&lead).Update("state", to).Error; err != nil {
		return nil, err
	}
	return &lead, nil
}
