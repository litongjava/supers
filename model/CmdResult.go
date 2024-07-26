package model

type CommandResult struct {
  Success bool   `json:"success"`
  Output  string `json:"output"`
  Error   error  `json:"error"`
  Time    int64  `json:"time"`
}
