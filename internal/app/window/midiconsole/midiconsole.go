package midiconsole

/*
┍━━━━━━━━━━━━━━━━━━━━╳┑
│ MIDI Console Window │
└─────────────────────┘
*/
import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/window"
	"bitbox-editor/internal/logging"
	"bitbox-editor/internal/midi"
	"sync"
	"time"
)

var log = logging.NewLogger("midiconsole")

type MidiPortsPayload struct {
	Ports []string
}

type MidiConsoleWindow struct {
	*window.Window[*MidiConsoleWindow]
	component.CommandRouter
	eventbus.EventRouter

	midiManager *midi.MidiManager

	portMonitor     bool
	stopPortMonitor chan struct{}
	portMonitorWG   sync.WaitGroup
}

func NewMidiConsoleWindow(midiMgr *midi.MidiManager) *MidiConsoleWindow {
	w := &MidiConsoleWindow{
		midiManager: midiMgr,
	}

	w.Window = window.NewWindow[*MidiConsoleWindow]("Midi", "KeyboardMusic")

	w.Window.SetLayoutBuilder(w)

	uuid := w.UUID()
	w.EventRouter.Init(uuid)

	return w
}

func (w *MidiConsoleWindow) startPortMonitor() {
	log.Debug("Starting midi port detection ...")
	w.portMonitorWG.Add(1)
	currentStopChan := w.stopPortMonitor

	go func() {
		defer w.portMonitorWG.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		forceCheck := true
		for {
			select {
			case <-ticker.C:
				forceCheck = true
			case <-currentStopChan:
				log.Debug("Stopping midi port goroutine...")
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
			if !forceCheck {
				continue
			}
			forceCheck = false

			detectedPorts := w.midiManager.ListPorts()
			payload := MidiPortsPayload{
				Ports: detectedPorts,
			}
			cmd := component.UpdateCmd{Type: cmdMidiPorts, Data: payload}
			w.SendUpdate(cmd)
		}
	}()
}

func (w *MidiConsoleWindow) Menu() {}

func (w *MidiConsoleWindow) Layout() {
	w.EventRouter.ProcessEvents()
	w.Window.ProcessUpdates()
}

func (w *MidiConsoleWindow) Destroy() {
	if w.stopPortMonitor != nil {
		select {
		case <-w.stopPortMonitor:
		default:
			close(w.stopPortMonitor)
		}
	}

	log.Debug("Waiting for midi port monitor to stop ...")
	w.portMonitorWG.Wait()

	log.Debug("Midi port monitor stopped.")

	w.EventRouter.Destroy()

}
