package application

import (
	"devsearch-go/internal/domain"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	CreateUser(user *domain.User) error
	FindUserByUsername(username string) (*domain.User, error)
	FindUserByUsernameOrEmail(username, email string) (*domain.User, error)
	FindUserByID(id uuid.UUID) (*domain.User, error)
	UpdateUser(user *domain.User) error
	DeleteUser(id uuid.UUID) error
}

// ProfileRepository defines the interface for profile data operations.
type ProfileRepository interface {
	CreateProfile(profile *domain.Profile) error
	FindProfileByID(id uuid.UUID) (*domain.Profile, error)
	FindProfileByUserID(userID uuid.UUID) (*domain.Profile, error)
	FindAllProfiles(searchQuery string, page, limit int) ([]domain.Profile, int64, error)
	UpdateProfile(profile *domain.Profile) error
}

// SkillRepository defines the interface for skill data operations.
type SkillRepository interface {
	CreateSkill(skill *domain.Skill) error
	FindSkillByID(id uuid.UUID) (*domain.Skill, error)
	FindUserSkill(skillID, userID uuid.UUID) (*domain.Skill, error)
	UpdateSkill(skill *domain.Skill) error
	DeleteSkill(id uuid.UUID) error
}

// MessageRepository defines the interface for message data operations.
type MessageRepository interface {
	CreateMessage(message *domain.Message) error
	FindMessagesByRecipientID(recipientID uuid.UUID) ([]domain.Message, error)
	FindMessageByIDAndRecipientID(messageID, recipientID uuid.UUID) (*domain.Message, error)
	UpdateMessage(message *domain.Message) error
	GetUnreadMessageCount(recipientID uuid.UUID) (int64, error)
}
