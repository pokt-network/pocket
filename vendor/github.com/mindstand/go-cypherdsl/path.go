package go_cypherdsl

import (
	"errors"
	"strings"
)

type PathBuilder struct {
	//match steps are a linked list, store the first node for further iteration
	firstStep *matchStep

	//current step so we can append easily
	currentStep *matchStep

	//save all of the errors for the end
	errors []error
}

func Path() *PathBuilder {
	return NewPath()
}

func NewPath() *PathBuilder {
	return &PathBuilder{}
}

type VertexStep struct {
	builder *PathBuilder
}

func (v *VertexStep) Build() *PathBuilder {
	return v.builder
}

type EdgeStep struct {
	builder *PathBuilder
}

func (e *EdgeStep) Done() *PathBuilder {
	return e.builder
}

type PStep struct {
	builder *PathBuilder
}

func (p *PStep) Done() *PathBuilder {
	return p.builder
}

type matchStep struct {
	Vertices []V
	Edge     *E

	P bool

	OtherOperation string

	Next *matchStep
}

func (m *matchStep) ToCypher() (string, error) {
	if m.P {
		return "p=", nil
	}

	if m.OtherOperation != "" {
		return m.OtherOperation, nil
	}

	if m.Edge != nil {
		return m.Edge.ToCypher()
	}

	if m.Vertices != nil && len(m.Vertices) > 0 {
		str := ""

		for _, v := range m.Vertices {
			cyph, err := v.ToCypher()
			if err != nil {
				return "", err
			}

			str += cyph + ","
		}

		return strings.TrimSuffix(str, ","), nil
	}

	return "", errors.New("nothing in the match step was specified")
}

func (m *PathBuilder) ToCypher() (string, error) {
	if m.firstStep == nil {
		return "", errors.New("no steps process")
	}

	if len(m.errors) != 0 {
		errStr := ""
		for _, err := range m.errors {
			errStr += ";" + err.Error()
		}

		return "", errors.New("incurred one or many errors: " + errStr)
	}

	query := ""

	step := m.firstStep

	for {
		if step == nil {
			break
		}

		cypher, err := step.ToCypher()
		if err != nil {
			return "", err
		}

		query += cypher

		step = step.Next
	}

	return query, nil
}

func (v *VertexStep) ToCypher() (string, error) {
	return v.builder.ToCypher()
}

func (m *PathBuilder) P() *PathBuilder {
	newStep := &matchStep{
		P: true,
	}

	//its the first step
	if m.currentStep == nil {
		m.firstStep = newStep
		m.currentStep = newStep
	} else {
		m.currentStep.Next = newStep
		m.currentStep = newStep
	}

	return m
}

func (p *PStep) V(vertices ...V) *PathBuilder {
	if vertices == nil || len(vertices) == 0 {
		if p.builder.errors == nil {
			p.builder.errors = []error{}
		}
		p.builder.errors = append(p.builder.errors, errors.New("vertices can not be nil or empty"))
	}

	newStep := &matchStep{
		Vertices: vertices,
	}

	p.builder.currentStep.Next = newStep
	p.builder.currentStep = newStep

	return p.builder
}

func (m *PathBuilder) V(vertices ...V) *VertexStep {
	if vertices == nil || len(vertices) == 0 {
		if m.errors == nil {
			m.errors = []error{}
		}
		m.errors = append(m.errors, errors.New("vertices can not be nil or empty"))
	}

	newStep := &matchStep{
		Vertices: vertices,
	}

	//its the first step
	if m.currentStep == nil {
		m.firstStep = newStep
		m.currentStep = newStep
	} else {
		m.currentStep.Next = newStep
		m.currentStep = newStep
	}

	return &VertexStep{
		builder: m,
	}
}

func (e *EdgeStep) V(vertices ...V) *VertexStep {
	if vertices == nil || len(vertices) == 0 {
		if e.builder.errors == nil {
			e.builder.errors = []error{}
		}
		e.builder.errors = append(e.builder.errors, errors.New("vertices can not be nil or empty"))
	}

	newStep := &matchStep{
		Vertices: vertices,
	}

	e.builder.currentStep.Next = newStep
	e.builder.currentStep = newStep

	return &VertexStep{
		builder: e.builder,
	}
}

func (v *VertexStep) E(edge E) *EdgeStep {
	newStep := &matchStep{
		Edge: &edge,
	}

	v.builder.currentStep.Next = newStep
	v.builder.currentStep = newStep

	return &EdgeStep{
		builder: v.builder,
	}
}
