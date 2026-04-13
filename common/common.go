package common

type Task struct {
	ID        int      `json:"id"`
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	Output    string   `json:"output"`
	Status    string   `json:"status"`
	Error     string   `json:"error"`
	CreatedAt string   `json:"created_at"`
}
