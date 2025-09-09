package infrastructure

import (
	"devsearch-go/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormUserRepository implements the application.UserRepository interface using GORM.
type GormUserRepository struct {
	DB *gorm.DB
}

// CreateUser creates a new user.
func (r *GormUserRepository) CreateUser(user *domain.User) error {
	return r.DB.Create(user).Error
}

// FindUserByUsername retrieves a user by their username.
func (r *GormUserRepository) FindUserByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByUsernameOrEmail retrieves a user by their username or email.
func (r *GormUserRepository) FindUserByUsernameOrEmail(username, email string) (*domain.User, error) {
	var user domain.User
	if err := r.DB.Where("username = ?", username).Or("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByID retrieves a user by their ID.
func (r *GormUserRepository) FindUserByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.DB.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user.
func (r *GormUserRepository) UpdateUser(user *domain.User) error {
	return r.DB.Save(user).Error
}

// DeleteUser deletes a user by their ID.
func (r *GormUserRepository) DeleteUser(id uuid.UUID) error {
	return r.DB.Delete(&domain.User{}, "id = ?", id).Error
}

// GormProfileRepository implements the application.ProfileRepository interface using GORM.
type GormProfileRepository struct {
	DB *gorm.DB
}

// CreateProfile creates a new profile.
func (r *GormProfileRepository) CreateProfile(profile *domain.Profile) error {
	return r.DB.Create(profile).Error
}

// FindProfileByID retrieves a profile by its ID.
func (r *GormProfileRepository) FindProfileByID(id uuid.UUID) (*domain.Profile, error) {
	var profile domain.Profile
	if err := r.DB.Preload("Skills").Preload("Projects.Tags").First(&profile, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

// FindProfileByUserID retrieves a profile by its user ID.
func (r *GormProfileRepository) FindProfileByUserID(userID uuid.UUID) (*domain.Profile, error) {
	var profile domain.Profile
	if err := r.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

// FindAllProfiles retrieves all profiles with optional search and pagination.
func (r *GormProfileRepository) FindAllProfiles(searchQuery string, page, limit int) ([]domain.Profile, int64, error) {
	var profiles []domain.Profile
	query := r.DB.Preload("Skills")

	if searchQuery != "" {
		query = query.Where("name ILIKE ? OR short_intro ILIKE ? OR bio ILIKE ?", "%"+searchQuery+"%", "%"+searchQuery+"%", "%"+searchQuery+"%")
	}

	var totalProfiles int64
	query.Model(&domain.Profile{}).Count(&totalProfiles)

	offset := (page - 1) * limit
	err := query.Order("created ASC").Limit(limit).Offset(offset).Find(&profiles).Error
	if err != nil {
		return nil, 0, err
	}
	return profiles, totalProfiles, nil
}

// UpdateProfile updates an existing profile.
func (r *GormProfileRepository) UpdateProfile(profile *domain.Profile) error {
	return r.DB.Save(profile).Error
}

// GormSkillRepository implements the application.SkillRepository interface using GORM.
type GormSkillRepository struct {
	DB *gorm.DB
}

// CreateSkill creates a new skill.
func (r *GormSkillRepository) CreateSkill(skill *domain.Skill) error {
	return r.DB.Create(skill).Error
}

// FindSkillByID retrieves a skill by its ID.
func (r *GormSkillRepository) FindSkillByID(id uuid.UUID) (*domain.Skill, error) {
	var skill domain.Skill
	if err := r.DB.First(&skill, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &skill, nil
}

// FindUserSkill retrieves a skill for a specific user.
func (r *GormSkillRepository) FindUserSkill(skillID, userID uuid.UUID) (*domain.Skill, error) {
	var skill domain.Skill
	if err := r.DB.Where("owner_id IN (SELECT id FROM profiles WHERE user_id = ?)", userID).First(&skill, "id = ?", skillID).Error; err != nil {
		return nil, err
	}
	return &skill, nil
}

// UpdateSkill updates an existing skill.
func (r *GormSkillRepository) UpdateSkill(skill *domain.Skill) error {
	return r.DB.Save(skill).Error
}

// DeleteSkill deletes a skill by its ID.
func (r *GormSkillRepository) DeleteSkill(id uuid.UUID) error {
	return r.DB.Delete(&domain.Skill{}, "id = ?", id).Error
}

// GormMessageRepository implements the application.MessageRepository interface using GORM.
type GormMessageRepository struct {
	DB *gorm.DB
}

// CreateMessage creates a new message.
func (r *GormMessageRepository) CreateMessage(message *domain.Message) error {
	return r.DB.Create(message).Error
}

// FindMessagesByRecipientID retrieves all messages for a recipient.
func (r *GormMessageRepository) FindMessagesByRecipientID(recipientID uuid.UUID) ([]domain.Message, error) {
	var messages []domain.Message
	if err := r.DB.Preload("Sender").Where("recipient_id = ?", recipientID).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// FindMessageByIDAndRecipientID retrieves a single message by ID and recipient ID.
func (r *GormMessageRepository) FindMessageByIDAndRecipientID(messageID, recipientID uuid.UUID) (*domain.Message, error) {
	var message domain.Message
	if err := r.DB.Preload("Sender").Preload("Recipient").Where("recipient_id = ?", recipientID).First(&message, "id = ?", messageID).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

// UpdateMessage updates an existing message.
func (r *GormMessageRepository) UpdateMessage(message *domain.Message) error {
	return r.DB.Save(message).Error
}

// GetUnreadMessageCount retrieves the count of unread messages for a recipient.
func (r *GormMessageRepository) GetUnreadMessageCount(recipientID uuid.UUID) (int64, error) {
	var count int64
	if err := r.DB.Model(&domain.Message{}).Where("recipient_id = ? AND is_read = ?", recipientID, false).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
