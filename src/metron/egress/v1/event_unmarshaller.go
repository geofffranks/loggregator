package v1

import (
	"errors"
	"fmt"
	"log"
	"unicode"
	"unicode/utf8"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

var (
	invalidEnvelope = errors.New("Invalid Envelope")
	metricNames     map[events.Envelope_EventType]string
)

func init() {
	metricNames = make(map[events.Envelope_EventType]string)
	for eventType, eventName := range events.Envelope_EventType_name {
		r, n := utf8.DecodeRuneInString(eventName)
		modifiedName := string(unicode.ToLower(r)) + eventName[n:]
		metricName := "dropsondeUnmarshaller." + modifiedName + "Received"
		metricNames[events.Envelope_EventType(eventType)] = metricName
	}
}

// An EventUnmarshaller is an self-instrumenting tool for converting Protocol
// Buffer-encoded dropsonde messages to Envelope instances.
type EventUnmarshaller struct {
	outputWriter EnvelopeWriter
	batcher      EventBatcher
}

func NewUnMarshaller(outputWriter EnvelopeWriter, batcher EventBatcher) *EventUnmarshaller {
	return &EventUnmarshaller{
		outputWriter: outputWriter,
		batcher:      batcher,
	}
}

func (u *EventUnmarshaller) Write(message []byte) {
	envelope, err := u.UnmarshallMessage(message)
	if err != nil {
		log.Printf("Error unmarshalling: %s", err)
		return
	}
	u.outputWriter.Write(envelope)
}

func (u *EventUnmarshaller) UnmarshallMessage(message []byte) (*events.Envelope, error) {
	envelope := &events.Envelope{}
	err := proto.Unmarshal(message, envelope)
	if err != nil {
		log.Printf("eventUnmarshaller: unmarshal error %v", err)
		u.batcher.BatchIncrementCounter("dropsondeUnmarshaller.unmarshalErrors")
		return nil, err
	}

	if !valid(envelope) {
		log.Printf("eventUnmarshaller: validation failed for message %v", envelope.GetEventType())
		u.batcher.BatchIncrementCounter("dropsondeUnmarshaller.unmarshalErrors")
		return nil, invalidEnvelope
	}

	if err := u.incrementReceiveCount(envelope.GetEventType()); err != nil {
		log.Printf("Error incrementing receive count: %s", err)
		return nil, err
	}

	return envelope, nil
}

func (u *EventUnmarshaller) incrementReceiveCount(eventType events.Envelope_EventType) error {
	var err error
	switch eventType {
	case events.Envelope_LogMessage:
		// LogMessage is a special case. `logMessageReceived` used to be broken out by app ID, and
		// `logMessageTotal` was the sum of all of those.
		u.batcher.BatchIncrementCounter("dropsondeUnmarshaller.logMessageTotal")
	default:
		metricName := metricNames[eventType]
		if metricName == "" {
			metricName = "dropsondeUnmarshaller.unknownEventTypeReceived"
			err = fmt.Errorf("eventUnmarshaller: received unknown event type %#v", eventType)
		}
		u.batcher.BatchIncrementCounter(metricName)
	}

	u.batcher.BatchCounter("dropsondeUnmarshaller.receivedEnvelopes").
		SetTag("protocol", "udp").
		SetTag("event_type", eventType.String()).
		Increment()

	return err
}

func valid(env *events.Envelope) bool {
	switch env.GetEventType() {
	case events.Envelope_HttpStartStop:
		return env.GetHttpStartStop() != nil
	case events.Envelope_LogMessage:
		return env.GetLogMessage() != nil
	case events.Envelope_ValueMetric:
		return env.GetValueMetric() != nil
	case events.Envelope_CounterEvent:
		return env.GetCounterEvent() != nil
	case events.Envelope_Error:
		return env.GetError() != nil
	case events.Envelope_ContainerMetric:
		return env.GetContainerMetric() != nil
	}
	return true
}
