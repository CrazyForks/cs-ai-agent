package dsl

import "encoding/json"

type Definition struct {
	SchemaVersion int    `json:"schemaVersion"`
	EntryNodeID   string `json:"entryNodeId"`
	Nodes         []Node `json:"nodes"`
	Edges         []Edge `json:"edges"`
}

type Node struct {
	ID       string                      `json:"id"`
	Type     string                      `json:"type"`
	Name     string                      `json:"name"`
	Position Position                    `json:"position"`
	Config   json.RawMessage             `json:"config"`
	Inputs   map[string]VariableSelector `json:"inputs,omitempty"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Edge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type ConditionConfig struct {
	Branches []ConditionBranch `json:"branches,omitempty"`
}

type ConditionBranch struct {
	ID           string     `json:"id"`
	Name         string     `json:"name,omitempty"`
	TargetNodeID string     `json:"targetNodeId"`
	Condition    *Condition `json:"condition,omitempty"`
	Default      bool       `json:"default,omitempty"`
}

type Condition struct {
	Expression string            `json:"expression,omitempty"`
	Left       *VariableSelector `json:"left,omitempty"`
	Operator   string            `json:"operator,omitempty"`
	Right      any               `json:"right,omitempty"`
}

type VariableSelector struct {
	NodeID string `json:"nodeId"`
	Field  string `json:"field"`
}
