package model

type CommandResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Time    int64  `json:"time"`
	Err     error  `json:"error"`
}
