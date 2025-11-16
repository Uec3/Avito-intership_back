package models

type Team struct {
	Name    string `gorm:"primaryKey" json:"team_name"`
	Members []User `gorm:"foreignKey:TeamName" json:"-"`
}

type User struct {
	ID       string `gorm:"primaryKey" json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `gorm:"default:true" json:"is_active"`
	TeamName string `gorm:"foreignKey:TeamName" json:"-"`
}

type PullRequest struct {
	ID                string `gorm:"primaryKey" json:"pull_request_id"`
	Name              string `json:"pull_request_name"`
	AuthorId          string `json:"author_id"`
	Status            string `json:"status"` //[OPEN, MERGED]
	AssignedReviewers []User `gorm:"many2many:pr_reviewers;" json:"assigned_reviewers"`
	Author            User   `gorm:"foreginKey:AuthorID" json:"-"`
}
