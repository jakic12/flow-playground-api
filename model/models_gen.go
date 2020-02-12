// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"github.com/google/uuid"
)

type Event struct {
	Type   string      `json:"type"`
	Values []*XDRValue `json:"values"`
}

type NewScriptExecution struct {
	ProjectID uuid.UUID `json:"projectId"`
	Script    string    `json:"script"`
}

type NewScriptTemplate struct {
	ProjectID uuid.UUID `json:"projectId"`
	Script    string    `json:"script"`
}

type NewTransactionExecution struct {
	ProjectID uuid.UUID `json:"projectId"`
	Script    string    `json:"script"`
}

type NewTransactionTemplate struct {
	ProjectID uuid.UUID `json:"projectId"`
	Script    string    `json:"script"`
}

type UpdateScriptTemplate struct {
	ID     uuid.UUID `json:"id"`
	Script string    `json:"script"`
}

type UpdateTransactionTemplate struct {
	ID     uuid.UUID `json:"id"`
	Index  *int      `json:"index"`
	Script *string   `json:"script"`
}

type XDRValue struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
