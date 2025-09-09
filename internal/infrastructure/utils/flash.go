package utils

import (
	"devsearch-go/internal/domain"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ( // Flash message keys
	FlashSuccess = "flash_success"
	FlashError   = "flash_error"
	FlashInfo    = "flash_info"
)

// SetFlashMessage sets a flash message in the session.
func SetFlashMessage(c *gin.Context, key string, message string) {
	session := sessions.Default(c)
	session.AddFlash(message, key)
	session.Save()
}

// GetFlashMessages retrieves and clears flash messages from the session.
func GetFlashMessages(c *gin.Context, key string) []string {
	session := sessions.Default(c)
	flashes := session.Flashes(key)
	session.Save()

	var messages []string
	for _, flash := range flashes {
		if msg, ok := flash.(string); ok {
			messages = append(messages, msg)
		}
	}
	return messages
}

type TemplateData struct {
	FlashSuccess    []string
	FlashError      []string
	FlashInfo       []string
	IsAuthenticated bool
	// Page specific data
	Profile         domain.Profile
	Profiles        []domain.Profile
	Project         domain.Project
	Projects        []domain.Project
	Skill           domain.Skill // Added for skill forms
	Skills          []domain.Skill
	TopSkills       []domain.Skill
	OtherSkills     []domain.Skill
	Message         domain.Message
	MessageRequests []domain.Message
	Recipient       domain.Profile

	SearchQuery string
	Pagination  PaginationData

	UnreadCount int64
	FormTitle   string
	Object      interface{} // For delete operations

	CurrentUserID uuid.UUID
	IsOwner       bool
	HasReviewed   bool
	Page          string // For login/register page differentiation
}

// GetTemplateData initializes common template data, including flash messages and authentication status.
func GetTemplateData(c *gin.Context, isAuthenticated bool) TemplateData {
	return TemplateData{
		FlashSuccess:    GetFlashMessages(c, FlashSuccess),
		FlashError:      GetFlashMessages(c, FlashError),
		FlashInfo:       GetFlashMessages(c, FlashInfo),
		IsAuthenticated: isAuthenticated,
	}
}
