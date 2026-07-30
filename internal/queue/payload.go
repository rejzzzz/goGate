package queue

// AsyncPayload represents the envelope for asynchronous requests sent to the queue.
type AsyncPayload struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}
