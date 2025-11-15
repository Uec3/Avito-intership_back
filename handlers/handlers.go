package handlers

import (
	"avito_intern_dev/models"
	"errors"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var errTeamExist = errors.New("team_name already exists")

var notFound = errors.New("resource not found")

var PrExist = errors.New("PR id already exists ")

var PrMerged = errors.New(" cannot cannot reassign on merged PR")
var notAssigned = errors.New("reviewer is not assigned to this PR")
var noCandidate = errors.New("no active replacement candidate in team")

func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

func AddTeam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			TeamName string `json:"team_name" binding:"required"`
			Members  []struct {
				UserID   string `json:"user_id" binding:"required"`
				Username string `json:"username" binding:"required"`
				IsActive bool   `json:"is_active`
			} `json:"members" binding:"required,min=1"`
		}
		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var existing models.Team
		if db.Where("name = ?", req.TeamName).First(&existing).Error == nil {
			respondError(c, 400, "TEAM_EXIST", errTeamExist.Error())
			return
		}
		team := models.Team{Name: req.TeamName}
		db.Create(&team)
		for _, m := range req.Members {
			user := models.User{
				ID:       m.UserID,
				Username: m.Username,
				IsActive: m.IsActive,
				TeamName: req.TeamName,
			}
			db.Save(&user)
		}
		var result models.Team

		db.Preload("Members").Where("name = ?", req.TeamName).First(&result)
		c.JSON(201, gin.H{
			"team": gin.H{
				"team_name": result.Name,
				"members":   result.Members,
			},
		})
	}
}

func GetTeam(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamName := c.Query("team_name")
		if teamName == "" {
			c.JSON(400, gin.H{"error": "team_name required"})
			return
		}

		var team models.Team
		err := db.Preload("Members").Where("name = ?", teamName).First(&team).Error
		if err != nil {
			respondError(c, 404, "NOT_FOUND", "team not found")
			return
		}
		c.JSON(200, team.Members)
	}
}
func SetIsActive(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID   string `json:"user_id" binding:"required"`
			IsActive bool   `json:"is_active"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		var user models.User
		if err := db.First(&user, "id = ?", req.UserID).Error; err != nil {
			respondError(c, 404, "NOT_FOUND", "user not found")
			return
		}

		user.IsActive = req.IsActive
		db.Save(&user)

		c.JSON(200, gin.H{"user": user})
	}
}

func CreatePR(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PR_ID    string `json:"pull_request_id" binding:"required"`
			Name     string `json:"pull_request_name" binding:"required"`
			AuthorID string `json:"author_id" binding:"required"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if db.Where("id = ?", req.PR_ID).First(&models.PullRequest{}).Error == nil {
			respondError(c, 409, "PR_EXISTS", PrExist.Error())
			return
		}
		var author models.User
		err := db.Where("id = ?", req.AuthorID).First(&author).Error
		// err := db.Preload("Team").First(&author, "id = ?", req.AuthorID).Error
		if err != nil {
			respondError(c, 404, "NOT_FOUND", "author not found")
			return
		}

		now := time.Now()
		pr := models.PullRequest{
			ID:        req.PR_ID,
			Name:      req.Name,
			AuthorId:  req.AuthorID,
			Status:    "OPEN",
			CreatedAt: &now,
		}

		var candidates []models.User
		db.Where("team_name = ? AND is_active = ? AND id != ?", author.TeamName, true, author.ID).Find(&candidates)

		if len(candidates) < 2 {
			pr.AssignedReviewers = candidates
		} else {
			rand.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
			candidates = candidates[:2]
			pr.AssignedReviewers = candidates

		}
		db.Create(&pr)

		var result models.PullRequest
		db.Preload("AssignedReviewers").First(&result, "id = ?", req.PR_ID)
		c.JSON(201, gin.H{"pr": result})
	}
}
func MergePR(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PR_ID string `json:"pull_request_id" binding:"required"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var pr models.PullRequest
		if err := db.Preload("AssignedReviewers").First(&pr, "id = ?", req.PR_ID).Error; err != nil {
			respondError(c, 404, "NOT_FOUND", "PR not found")
			return
		}

		if pr.Status != "MERGED" {
			now := time.Now()
			pr.Status = "MERGED"
			pr.MergedAt = &now
			db.Save(&pr)
		}

		c.JSON(200, gin.H{"pr": pr})
	}
}

func ReassignReviewer(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PR_ID     string `json:"pull_request_id" binding:"required"`
			OldUserID string `json:"old_user_id" binding:"required"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var pr models.PullRequest
		if err := db.Preload("AssignedReviewers").First(&pr, "id = ?", req.PR_ID).Error; err != nil {
			respondError(c, 404, "NOT_FOUND", "PR not found")
			return
		}

		if pr.Status == "MERGED" {
			respondError(c, 409, "PR_MERGED", PrMerged.Error())
			return
		}

		var oldReviewer models.User
		found := false
		for _, r := range pr.AssignedReviewers {
			if r.ID == req.OldUserID {
				oldReviewer = r
				found = true
				break
			}
		}
		if !found {
			respondError(c, 409, "NOT_ASSIGNED", notAssigned.Error())
			return
		}

		// Кандидаты из команды oldReviewer
		exclude := map[string]bool{pr.AuthorId: true, req.OldUserID: true}
		for _, r := range pr.AssignedReviewers {
			if r.ID != req.OldUserID {
				exclude[r.ID] = true
			}
		}

		var candidates []models.User
		var excludeList []string
		for id := range exclude {
			excludeList = append(excludeList, id)
		}

		db.Where("team_name = ? AND is_active = ? AND id NOT IN ?", oldReviewer.TeamName, true, excludeList).
			Find(&candidates)

		if len(candidates) == 0 {
			respondError(c, 409, "NO_CANDIDATE", notAssigned.Error())
			return
		}

		newReviewer := candidates[rand.Intn(len(candidates))]
		newReviewers := []models.User{newReviewer}
		for _, r := range pr.AssignedReviewers {
			if r.ID != req.OldUserID {
				newReviewers = append(newReviewers, r)
			}
		}
		pr.AssignedReviewers = newReviewers
		db.Save(&pr)

		var updated models.PullRequest
		db.Preload("AssignedReviewers").First(&updated, "id = ?", req.PR_ID)

		c.JSON(200, gin.H{
			"pr":          updated,
			"replaced_by": newReviewer.ID,
		})
	}
}

func GetUserReviews(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(400, gin.H{"error": "user_id required"})
			return
		}

		var prs []models.PullRequest
		db.Preload("AssignedReviewers").Joins("JOIN pr_reviewers ON pr_reviewers.pull_request_id = pull_requests.id").Where("pr_reviewers.user_id = ?", userID).Find(&prs)
		shortPRs := []map[string]interface{}{}
		for _, pr := range prs {
			shortPRs = append(shortPRs, map[string]interface{}{
				"pull_request_id":   pr.ID,
				"pull_request_name": pr.Name,
				"author_id":         pr.AuthorId,
				"status":            pr.Status,
			})
		}
	}
}
