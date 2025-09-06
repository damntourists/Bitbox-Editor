package events

type LogRecord struct {
	index      int
	Timestamp  string `json:"ts"`
	Name       string `json:"name"`
	Level      string `json:"level"`
	Message    string `json:"msg"`
	Caller     string `json:"caller"`
	Func       string `json:"func"`
	Stacktrace string `json:"stacktrace"`
}
