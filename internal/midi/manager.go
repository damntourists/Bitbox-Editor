package midi

import (
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"fmt"
	"sync"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

var globalMidiManager *MidiManager
var Bus = eventbus.NewEventBus()

func init() {
	globalMidiManager = NewMidiManager(Bus)
}

func GetMidiManager() *MidiManager {
	return globalMidiManager
}

// MidiManager handles listening to MIDI ports and publishing events
type MidiManager struct {
	bus  *eventbus.EventBus
	stop func()
	mu   sync.Mutex
}

// NewMidiManager creates a new MIDI manager
func NewMidiManager(bus *eventbus.EventBus) *MidiManager {
	return &MidiManager{
		bus:  bus,
		stop: nil,
	}
}

// Close ensures the MIDI driver is closed (on app shutdown)
func (m *MidiManager) Close() {
	m.StopMonitoring()
	midi.CloseDriver()
}

// ListPorts returns a list of available MIDI input port names
func (m *MidiManager) ListPorts() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	inPorts := midi.GetInPorts()
	portNames := make([]string, len(inPorts))
	for i, port := range inPorts {
		portNames[i] = port.String()
	}
	return portNames
}

// StartMonitoring finds a port by name and starts listening to it
func (m *MidiManager) StartMonitoring(portName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop any existing listener
	if m.stop != nil {
		m.stop()
		m.stop = nil
	}

	inPort, err := midi.FindInPort(portName)
	if err != nil {
		return fmt.Errorf("could not find MIDI port '%s': %w", portName, err)
	}

	// Start the new listener
	stop, err := midi.ListenTo(inPort, func(msg midi.Message, timestampms int32) {
		// Parse the message and publish it
		event := m.parseMessage(msg, timestampms, inPort.String())
		if event != nil {
			m.bus.Publish(event)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to listen to MIDI port: %w", err)
	}

	// Store the stop function
	m.stop = stop
	return nil
}

// StopMonitoring stops the active MIDI listener.
func (m *MidiManager) StopMonitoring() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stop != nil {
		m.stop()
		m.stop = nil
	}
}

// parseMessage converts a raw midi.Message into a MidiEventRecord.
func (m *MidiManager) parseMessage(msg midi.Message, timestampms int32, portID string) events.Event {
	var ch, key, vel, controller, value, program, pressure uint8
	var pitch uint16
	var rel int16

	switch {
	case msg.GetNoteOn(&ch, &key, &vel):
		// A NoteOn with 0 velocity is often a NoteOff
		if vel == 0 {
			return events.MidiEventRecord{
				EventType: events.MidiNoteOffEvent,
				Timestamp: timestampms,
				PortID:    portID,
				Channel:   ch,
				Key:       key,
				Velocity:  0,
			}
		}
		return events.MidiEventRecord{
			EventType: events.MidiNoteOnEvent,
			Timestamp: timestampms,
			PortID:    portID,
			Channel:   ch,
			Key:       key,
			Velocity:  vel,
		}

	case msg.GetNoteOff(&ch, &key, &vel):
		return events.MidiEventRecord{
			EventType: events.MidiNoteOffEvent,
			Timestamp: timestampms,
			PortID:    portID,
			Channel:   ch,
			Key:       key,
			Velocity:  0,
		}

	case msg.GetControlChange(&ch, &controller, &value):
		return events.MidiEventRecord{
			EventType:  events.MidiControlChangeEvent,
			Timestamp:  timestampms,
			PortID:     portID,
			Channel:    ch,
			Controller: controller,
			Value:      value,
		}

	case msg.GetPitchBend(&ch, &rel, &pitch):
		return events.MidiEventRecord{
			EventType: events.MidiPitchBendEvent,
			Timestamp: timestampms,
			PortID:    portID,
			Channel:   ch,
			Value14:   pitch,
		}

	case msg.GetProgramChange(&ch, &program):
		return events.MidiEventRecord{
			EventType: events.MidiProgramChangeEvent,
			Timestamp: timestampms,
			PortID:    portID,
			Channel:   ch,
			Value:     program,
		}

	case msg.GetAfterTouch(&ch, &pressure):
		return events.MidiEventRecord{
			EventType: events.MidiAfterTouchEvent,
			Timestamp: timestampms,
			PortID:    portID,
			Channel:   ch,
			Value:     pressure,
		}

	case msg.GetPolyAfterTouch(&ch, &key, &pressure):
		return events.MidiEventRecord{
			EventType: events.MidiPolyAfterTouchEvent,
			Timestamp: timestampms,
			PortID:    portID,
			Channel:   ch,
			Key:       key,
			Velocity:  pressure,
		}
	}

	return nil
}
