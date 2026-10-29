package dto

type TaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Assignee    string `json:"assignee"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	DueDate     string `json:"dueDate"`
}
type SaveRequest struct {
	TaskInput
	Version int64 `json:"version"`
}
type MoveRequest struct {
	Version  int64  `json:"version"`
	Status   string `json:"status"`
	BeforeID int64  `json:"beforeId"`
}
type DeleteRequest struct {
	Version int64 `json:"version"`
}
type Task struct {
	ID int64 `json:"id"`
	TaskInput
	UpdatedAt string `json:"updatedAt"`
}
type Board struct {
	Items   []Task `json:"items"`
	Version int64  `json:"version"`
}
