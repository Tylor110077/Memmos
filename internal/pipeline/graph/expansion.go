package graph

import (
	"errors"
	"fmt"
	"strings"
)

type ExpansionRules struct {
	MaxNewNodes int
}

func (r ExpansionRules) Validate() error {
	if r.MaxNewNodes != 1 {
		return errors.New("expansion must generate exactly one node")
	}
	return nil
}

type ExpansionInput struct {
	Current   Node
	Neighbors []Node
}

type Expansion struct {
	Node Node
	Edge Edge
}

type ExpansionGenerator struct {
	rules ExpansionRules
}

func NewExpansionGenerator(rules ExpansionRules) *ExpansionGenerator {
	return &ExpansionGenerator{rules: rules}
}

func (g *ExpansionGenerator) Generate(input ExpansionInput) (Expansion, error) {
	if err := g.rules.Validate(); err != nil {
		return Expansion{}, err
	}
	base := strings.TrimSpace(input.Current.Name)
	if base == "" {
		return Expansion{}, errors.New("current node name is required")
	}
	name := base + " Related"
	description := "Directly related to " + base
	return Expansion{
		Node: Node{
			ID:          input.Current.ID + "_exp",
			Name:        name,
			Type:        "expanded",
			Description: description,
			Meaning:     description,
			Level:       input.Current.Level + 1,
		},
		Edge: Edge{
			ID:       fmt.Sprintf("%s_to_%s", input.Current.ID, input.Current.ID+"_exp"),
			SourceID: input.Current.ID,
			TargetID: input.Current.ID + "_exp",
			Relation: "expands",
		},
	}, nil
}
