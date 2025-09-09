package application

import (
	"devsearch-go/internal/domain"

	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserUseCase defines the business logic for users and profiles.
type UserUseCase struct {
	UserRepo    UserRepository
	ProfileRepo ProfileRepository
	SkillRepo   SkillRepository
	MessageRepo MessageRepository
}

// NewUserUseCase creates a new UserUseCase.
func NewUserUseCase(userRepo UserRepository, profileRepo ProfileRepository, skillRepo SkillRepository, messageRepo MessageRepository) *UserUseCase {
	return &UserUseCase{
		UserRepo:    userRepo,
		ProfileRepo: profileRepo,
		SkillRepo:   skillRepo,
		MessageRepo: messageRepo,
	}
}

// RegisterUser registers a new user and creates their profile.
func (uc *UserUseCase) RegisterUser(username, email, password string) (*domain.User, *domain.Profile, error) {
	// Check if username or email already exists
	if existingUser, err := uc.UserRepo.FindUserByUsernameOrEmail(username, email); err == nil && existingUser != nil {
		return nil, nil, fmt.Errorf("username or email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := domain.User{
		Username: username,
		Password: string(hashedPassword),
	}

	if err := uc.UserRepo.CreateUser(&user); err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	profile := domain.Profile{
		UserID:   user.ID,
		Name:     username,
		Email:    email,
		Username: username,
	}

	if err := uc.ProfileRepo.CreateProfile(&profile); err != nil {
		// If profile creation fails, try to delete the user to prevent orphaned data
		uc.UserRepo.DeleteUser(user.ID)
		return nil, nil, fmt.Errorf("failed to create user profile: %w", err)
	}

	return &user, &profile, nil
}

// LoginUser authenticates a user.
func (uc *UserUseCase) LoginUser(username, password string) (*domain.User, error) {
	user, err := uc.UserRepo.FindUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("username or password is incorrect")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("username or password is incorrect")
	}

	return user, nil
}

// GetUserAccount retrieves the authenticated user's account details.
func (uc *UserUseCase) GetUserAccount(userID uuid.UUID) (*domain.User, error) {
	user, err := uc.UserRepo.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// UpdateUserAccount updates the authenticated user's account details.
func (uc *UserUseCase) UpdateUserAccount(userID uuid.UUID, profileData map[string]string, profileImage string) (*domain.Profile, error) {
	user, err := uc.UserRepo.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	profile, err := uc.ProfileRepo.FindProfileByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found")
	}

	// Update profile fields
	profile.Name = profileData["name"]
	profile.Email = profileData["email"]
	profile.Username = profileData["username"]
	profile.ShortIntro = profileData["short_intro"]
	profile.Bio = profileData["bio"]
	profile.SocialGithub = profileData["social_github"]
	profile.SocialWebsite = profileData["social_website"]
	if profileImage != "" {
		profile.ProfileImage = profileImage
	}

	if err := uc.ProfileRepo.UpdateProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Also update the associated User's username if it changed
	if user.Username != profile.Username {
		user.Username = profile.Username
		if err := uc.UserRepo.UpdateUser(user); err != nil {
			return nil, fmt.Errorf("failed to update username in User table: %w", err)
		}
	}

	return profile, nil
}

// CreateSkill creates a new skill for a user.
func (uc *UserUseCase) CreateSkill(userID uuid.UUID, name, description string) (*domain.Skill, error) {
	profile, err := uc.ProfileRepo.FindProfileByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found for user: %w", err)
	}

	skill := domain.Skill{
		OwnerID:     profile.ID,
		Name:        name,
		Description: description,
	}

	if err := uc.SkillRepo.CreateSkill(&skill); err != nil {
		return nil, fmt.Errorf("failed to create skill: %w", err)
	}
	return &skill, nil
}

// UpdateSkill updates an existing skill.
func (uc *UserUseCase) UpdateSkill(skillID, userID uuid.UUID, name, description string) (*domain.Skill, error) {
	skill, err := uc.SkillRepo.FindUserSkill(skillID, userID)
	if err != nil {
		return nil, fmt.Errorf("skill not found or unauthorized: %w", err)
	}

	skill.Name = name
	skill.Description = description

	if err := uc.SkillRepo.UpdateSkill(skill); err != nil {
		return nil, fmt.Errorf("failed to update skill: %w", err)
	}
	return skill, nil
}

// DeleteSkill deletes a skill.
func (uc *UserUseCase) DeleteSkill(skillID, userID uuid.UUID) error {
	skill, err := uc.SkillRepo.FindUserSkill(skillID, userID)
	if err != nil {
		return fmt.Errorf("skill not found or unauthorized: %w", err)
	}

	return uc.SkillRepo.DeleteSkill(skill.ID)
}

// GetSkillByID retrieves a single skill by ID.
func (uc *UserUseCase) GetSkillByID(id uuid.UUID) (*domain.Skill, error) {
	return uc.SkillRepo.FindSkillByID(id)
}

// GetInbox retrieves all messages for the authenticated user.
func (uc *UserUseCase) GetInbox(userID uuid.UUID) ([]domain.Message, int64, error) {
	profile, err := uc.ProfileRepo.FindProfileByUserID(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("profile not found for authenticated user: %w", err)
	}

	messages, err := uc.MessageRepo.FindMessagesByRecipientID(profile.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch inbox messages: %w", err)
	}

	unreadCount, err := uc.MessageRepo.GetUnreadMessageCount(profile.ID)
	if err != nil {
		// Log error but continue as unread count is not critical
		unreadCount = 0
	}

	return messages, unreadCount, nil
}

// GetMessage retrieves a single message by ID.
func (uc *UserUseCase) GetMessage(messageID, userID uuid.UUID) (*domain.Message, error) {
	profile, err := uc.ProfileRepo.FindProfileByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found for authenticated user: %w", err)
	}

	message, err := uc.MessageRepo.FindMessageByIDAndRecipientID(messageID, profile.ID)
	if err != nil {
		return nil, fmt.Errorf("message not found or unauthorized: %w", err)
	}

	// Mark message as read
	if !message.IsRead {
		message.IsRead = true
		if err := uc.MessageRepo.UpdateMessage(message); err != nil {
			// Log error but continue
		}
	}

	return message, nil
}

// CreateMessage creates and sends a new message.
func (uc *UserUseCase) CreateMessage(senderUserID *uuid.UUID, recipientID uuid.UUID, name, email, subject, body string) error {
	var senderProfile *domain.Profile
	if senderUserID != nil {
		var err error
		senderProfile, err = uc.ProfileRepo.FindProfileByUserID(*senderUserID)
		if err != nil {
			// Log error but continue as message can be sent anonymously
			senderProfile = nil
		}
	}

	recipientProfile, err := uc.ProfileRepo.FindProfileByID(recipientID)
	if err != nil {
		return fmt.Errorf("recipient not found: %w", err)
	}

	message := domain.Message{
		RecipientID: recipientProfile.ID,
		Subject:     subject,
		Body:        body,
	}

	if senderProfile != nil {
		message.SenderID = senderProfile.ID
		message.Name = senderProfile.Name
		message.Email = senderProfile.Email
	} else {
		// If sender is not authenticated, use provided name and email from form
		message.Name = name
		message.Email = email
	}

	if err := uc.MessageRepo.CreateMessage(&message); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// GetAllProfiles retrieves all profiles with optional search and pagination.
func (uc *UserUseCase) GetAllProfiles(searchQuery string, page, limit int) ([]domain.Profile, int64, error) {
	return uc.ProfileRepo.FindAllProfiles(searchQuery, page, limit)
}

// GetProfileByID retrieves a single user profile by ID.
func (uc *UserUseCase) GetProfileByID(id uuid.UUID) (*domain.Profile, error) {
	return uc.ProfileRepo.FindProfileByID(id)
}
