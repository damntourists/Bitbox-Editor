package dragdrop

import (
	"sync"

	"github.com/AllenDang/cimgui-go/imgui"
)

var dragDropStore = make(map[imgui.ID]interface{})
var dragDropMutex sync.Mutex

type DataPayload struct {
	Type string
	Data interface{}
}

type TooltipFunc func()

// SetData stores the data for a drag operation (source drag)
func SetData(id imgui.ID, data interface{}) {
	dragDropMutex.Lock()
	defer dragDropMutex.Unlock()
	dragDropStore[id] = data
}

// GetData retrieves and removes the data for a drag operation (target drop)
func GetData(id imgui.ID) (interface{}, bool) {
	dragDropMutex.Lock()
	defer dragDropMutex.Unlock()
	data, ok := dragDropStore[id]
	if ok {
		delete(dragDropStore, id)
	}
	return data, ok
}

// ClearData removes data if the drag is cancelled
func ClearData(id imgui.ID) {
	dragDropMutex.Lock()
	defer dragDropMutex.Unlock()
	delete(dragDropStore, id)
}
