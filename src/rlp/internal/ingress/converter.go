package ingress

import (
	"plumbing/conversion"
	v2 "plumbing/v2"

	"github.com/cloudfoundry/sonde-go/events"
)

func NewConverter() *EnvelopeConverter {
	return &EnvelopeConverter{}
}

type EnvelopeConverter struct{}

func (e *EnvelopeConverter) Convert(payload []byte) (*v2.Envelope, error) {
	v1e := &events.Envelope{}
	err := v1e.Unmarshal(payload)
	if err != nil {
		return nil, err
	}

	return conversion.ToV2(v1e), nil
}
