package api

import "time"

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Workspace struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	TeamSize string           `json:"team_size"`
	Projects []Project        `json:"projects"`
	Member   *WorkspaceMember `json:"workspace_member"`
}

type WorkspaceMember struct {
	Role int `json:"role"`
}

type Project struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Spec struct {
	ID        string    `json:"id"`
	RequestID string    `json:"request_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Locked    bool      `json:"locked"`
	Request   *Request  `json:"request"`
	CreatedAt time.Time `json:"created_at"`
}

type Request struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Analysis      string         `json:"analysis"`
	Type          string         `json:"type"`
	Status        string         `json:"status"`
	Reach         float64        `json:"reach"`
	Impact        float64        `json:"impact"`
	GoalAlignment float64        `json:"goal_alignment"`
	Effort        float64        `json:"effort"`
	Score         float64        `json:"score"`
	Captures      []Capture      `json:"captures"`
	SolutionPaths []SolutionPath `json:"solution_paths"`
	Spec          *Spec          `json:"spec"`
	CreatedAt     time.Time      `json:"created_at"`
}

type SolutionPath struct {
	ID         string              `json:"id"`
	Question   string              `json:"question"`
	Context    string              `json:"context"`
	Impact     string              `json:"impact"`
	Multiple   bool                `json:"multiple"`
	Options    []SolutionOption    `json:"options"`
	Selections []SolutionSelection `json:"selections"`
}

type SolutionOption struct {
	Label  string `json:"label"`
	Impact string `json:"impact"`
}

type SolutionSelection struct {
	Label  string `json:"label"`
	Custom bool   `json:"custom"`
}

type Capture struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type Source struct {
	ID            string    `json:"id"`
	Provider      string    `json:"provider"`
	Type          string    `json:"type"`
	Subject       string    `json:"subject"`
	DisplayName   string    `json:"display_name"`
	LookupStatus  string    `json:"lookup_status"`
	LastIndexedAt time.Time `json:"last_indexed_at"`
}
