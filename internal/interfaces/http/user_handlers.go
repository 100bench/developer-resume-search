package http

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"devsearch-go/internal/domain"
	"devsearch-go/internal/infrastructure/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetProfiles handles fetching all user profiles
func (h *Handler) GetProfiles(c *gin.Context) {
	var profiles []domain.Profile

	searchQuery := c.Query("search_query")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := 3

	profiles, _, err := h.UserUseCase.GetAllProfiles(searchQuery, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profiles"})
		return
	}

	c.JSON(http.StatusOK, profiles)
}

// GetUserProfile handles fetching a single user profile by ID
func (h *Handler) GetUserProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID"})
		return
	}

	profile, err := h.UserUseCase.GetProfileByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// RegisterUser handles user registration
func (h *Handler) RegisterUser(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	password2 := c.PostForm("password2") // For password confirmation

	if password != password2 {
		utils.SetFlashMessage(c, utils.FlashError, "Passwords do not match")
		c.Redirect(http.StatusFound, "/register")
		return
	}

	_, _, err := h.UserUseCase.RegisterUser(username, email, password)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, err.Error())
		c.Redirect(http.StatusFound, "/register")
		return
	}

	// Set user ID in session upon successful registration (requires finding the user again)
	user, err := h.UserUseCase.LoginUser(username, password)
	if err != nil {
		log.Printf("Failed to log in user automatically after registration: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to log in user automatically")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	session := sessions.Default(c)
	session.Set("userID", user.ID.String())
	if err := session.Save(); err != nil {
		log.Printf("Failed to save session: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to log in user automatically")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	utils.SetFlashMessage(c, utils.FlashSuccess, "User account was created!")
	c.Redirect(http.StatusFound, "/profiles") // Redirect to profiles page after registration
}

// LoginUser handles user login
func (h *Handler) LoginUser(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	user, err := h.UserUseCase.LoginUser(username, password)
	if err != nil {
		utils.SetFlashMessage(c, utils.FlashError, err.Error())
		c.Redirect(http.StatusFound, "/login")
		return
	}

	session := sessions.Default(c)
	session.Set("userID", user.ID.String())
	if err := session.Save(); err != nil {
		log.Printf("Failed to save session: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to log in")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	utils.SetFlashMessage(c, utils.FlashInfo, "User was logged in!")
	c.Redirect(http.StatusFound, "/profiles")
}

// LogoutUser handles user logout
func (h *Handler) LogoutUser(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1}) // Expire the cookie
	if err := session.Save(); err != nil {
		log.Printf("Failed to save session on logout: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to log out")
		c.Redirect(http.StatusFound, "/")
		return
	}

	utils.SetFlashMessage(c, utils.FlashInfo, "User was logged out!")
	c.Redirect(http.StatusFound, "/login") // Redirect to login page after logout
}

// GetUserAccount handles fetching the authenticated user's account details
func (h *Handler) GetUserAccount(c *gin.Context) {
	userIDStr := sessions.Default(c).Get("userID")
	if userIDStr == nil {
		utils.SetFlashMessage(c, utils.FlashError, "User not authenticated")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.Printf("Invalid user ID in session: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to get user account")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	user, err := h.UserUseCase.GetUserAccount(userID)
	if err != nil {
		log.Printf("User not found for ID %s: %v", userID.String(), err)
		utils.SetFlashMessage(c, utils.FlashError, "User not found")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.JSON(http.StatusOK, user.Profile)
}

// UpdateUserAccount handles updating the authenticated user's account details
func (h *Handler) UpdateUserAccount(c *gin.Context) {
	userIDStr := sessions.Default(c).Get("userID")
	if userIDStr == nil {
		utils.SetFlashMessage(c, utils.FlashError, "User not authenticated")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.Printf("Invalid user ID in session: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to update account")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	profileData := map[string]string{
		"name":           c.PostForm("name"),
		"email":          c.PostForm("email"),
		"username":       c.PostForm("username"),
		"short_intro":    c.PostForm("short_intro"),
		"bio":            c.PostForm("bio"),
		"social_github":  c.PostForm("social_github"),
		"social_youtube": c.PostForm("social_youtube"),
		"social_website": c.PostForm("social_website"),
	}

	var profileImage string
	file, err := c.FormFile("profile_image")
	if err == nil && file != nil {
		src, err := file.Open()
		if err != nil {
			log.Printf("Failed to open image file: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to open image file")
			c.Redirect(http.StatusFound, "/edit-account")
			return
		}
		defer src.Close()

		if _, err := os.Stat("./media/profiles"); os.IsNotExist(err) {
			os.MkdirAll("./media/profiles", os.ModePerm)
		}

		filename := fmt.Sprintf("profiles/%s%s", uuid.New().String(), filepath.Ext(file.Filename))
		dst, err := os.Create(filepath.Join(".", "media", filename))
		if err != nil {
			log.Printf("Failed to save image file: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to save image file")
			c.Redirect(http.StatusFound, "/edit-account")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			log.Printf("Failed to copy image file: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to copy image file")
			c.Redirect(http.StatusFound, "/edit-account")
			return
		}
		profileImage = filename
	} else if err != nil && err != http.ErrMissingFile {
		log.Printf("Failed to get file: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, fmt.Sprintf("Failed to get file: %v", err))
		c.Redirect(http.StatusFound, "/edit-account")
		return
	}

	_, err = h.UserUseCase.UpdateUserAccount(userID, profileData, profileImage)
	if err != nil {
		log.Printf("Failed to update profile for user %s: %v", userID.String(), err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to update profile")
		c.Redirect(http.StatusFound, "/edit-account")
		return
	}

	utils.SetFlashMessage(c, utils.FlashSuccess, "Account was updated successfully!")
	c.Redirect(http.StatusFound, "/account")
}

// CreateSkill handles creating a new skill for a user
func (h *Handler) CreateSkill(c *gin.Context) {
	userIDStr := sessions.Default(c).Get("userID")
	if userIDStr == nil {
		utils.SetFlashMessage(c, utils.FlashError, "User not authenticated")
		c.Redirect(http.StatusFound, "/login")
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.Printf("Invalid user ID in session: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to create skill")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")

	_, err = h.UserUseCase.CreateSkill(userID, name, description)
	if err != nil {
		log.Printf("Failed to create skill for user %s: %v", userID.String(), err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to create skill")
		c.Redirect(http.StatusFound, "/create-skill")
		return
	}

	utils.SetFlashMessage(c, utils.FlashSuccess, "Skill was added successfully!")
	c.Redirect(http.StatusFound, "/account")
}

// UpdateSkill handles updating an existing skill
func (h *Handler) UpdateSkill(c *gin.Context) {
	userIDStr := sessions.Default(c).Get("userID")
	if userIDStr == nil {
		utils.SetFlashMessage(c, utils.FlashError, "User not authenticated")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.Printf("Invalid user ID in session: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to update skill")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SetFlashMessage(c, utils.FlashError, "Invalid skill ID")
		c.Redirect(http.StatusFound, "/account")
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")

	_, err = h.UserUseCase.UpdateSkill(id, userID, name, description)
	if err != nil {
		log.Printf("Failed to update skill %s for user %s: %v", idStr, userID.String(), err)
		utils.SetFlashMessage(c, utils.FlashError, err.Error())
		c.Redirect(http.StatusFound, fmt.Sprintf("/update-skill/%s", idStr))
		return
	}

	utils.SetFlashMessage(c, utils.FlashSuccess, "Skill was updated successfully!")
	c.Redirect(http.StatusFound, "/account")
}

// DeleteSkill handles deleting a skill
func (h *Handler) DeleteSkill(c *gin.Context) {
	userIDStr := sessions.Default(c).Get("userID")
	if userIDStr == nil {
		utils.SetFlashMessage(c, utils.FlashError, "User not authenticated")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.Printf("Invalid user ID in session: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to delete skill")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SetFlashMessage(c, utils.FlashError, "Invalid skill ID")
		c.Redirect(http.StatusFound, "/account")
		return
	}

	if err := h.UserUseCase.DeleteSkill(id, userID); err != nil {
		log.Printf("Failed to delete skill %s for user %s: %v", idStr, userID.String(), err)
		utils.SetFlashMessage(c, utils.FlashError, err.Error())
		c.Redirect(http.StatusFound, "/account")
		return
	}

	utils.SetFlashMessage(c, utils.FlashSuccess, "Skill was deleted successfully!")
	c.Redirect(http.StatusFound, "/account")
}

// GetInbox handles fetching all messages for the authenticated user
func (h *Handler) GetInbox(c *gin.Context) {
	userIDStr := sessions.Default(c).Get("userID")
	if userIDStr == nil {
		utils.SetFlashMessage(c, utils.FlashError, "User not authenticated")
		c.Redirect(http.StatusFound, "/login")
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.Printf("Invalid user ID in session: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to get inbox")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	messages, _, err := h.UserUseCase.GetInbox(userID)
	if err != nil {
		log.Printf("Error fetching messages for recipient %s: %v", userID.String(), err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to fetch inbox")
		c.Redirect(http.StatusFound, "/account") // Redirect to account or home
		return
	}

	c.JSON(http.StatusOK, messages)
}

// GetMessage handles fetching a single message by ID
func (h *Handler) GetMessage(c *gin.Context) {
	userIDStr := sessions.Default(c).Get("userID")
	if userIDStr == nil {
		utils.SetFlashMessage(c, utils.FlashError, "User not authenticated")
		c.Redirect(http.StatusFound, "/login")
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.Printf("Invalid user ID in session: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to get message")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		utils.SetFlashMessage(c, utils.FlashError, "Invalid message ID")
		c.Redirect(http.StatusFound, "/inbox")
		return
	}

	message, err := h.UserUseCase.GetMessage(messageID, userID)
	if err != nil {
		log.Printf("Message not found or unauthorized for user %s, message %s: %v", userID.String(), messageIDStr, err)
		utils.SetFlashMessage(c, utils.FlashError, err.Error())
		c.Redirect(http.StatusFound, "/inbox")
		return
	}

	c.JSON(http.StatusOK, message)
}

// CreateMessage handles sending a new message
func (h *Handler) CreateMessage(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")

	var senderUserID *uuid.UUID
	if userIDStr != nil {
		uID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			log.Printf("Invalid user ID in session: %v", err)
		} else {
			senderUserID = &uID
		}
	}

	recipientIDStr := c.Param("id")
	recipientID, err := uuid.Parse(recipientIDStr)
	if err != nil {
		utils.SetFlashMessage(c, utils.FlashError, "Invalid recipient ID")
		c.Redirect(http.StatusFound, "/profiles")
		return
	}

	name := c.PostForm("name")
	email := c.PostForm("email")
	subject := c.PostForm("subject")
	body := c.PostForm("body")

	if err := h.UserUseCase.CreateMessage(senderUserID, recipientID, name, email, subject, body); err != nil {
		log.Printf("Failed to send message: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to send message")
		c.Redirect(http.StatusFound, fmt.Sprintf("/create-message/%s", recipientID.String()))
		return
	}

	utils.SetFlashMessage(c, utils.FlashSuccess, "Your message was successfully sent!")
	c.Redirect(http.StatusFound, fmt.Sprintf("/profile/%s", recipientID.String()))
}

// RenderAccountPage renders the authenticated user's account page
func (h *Handler) RenderAccountPage(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

	var profile domain.Profile
	var skills []domain.Skill
	var projects []domain.Project

	if isAuthenticated {
		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			log.Printf("Invalid user ID in session: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to get user account")
			c.Redirect(http.StatusFound, "/login")
			return
		}

		userAccount, err := h.UserUseCase.GetUserAccount(userID)
		if err != nil {
			log.Printf("User not found for ID %s: %v", userID.String(), err)
			utils.SetFlashMessage(c, utils.FlashError, "User not found")
			c.Redirect(http.StatusFound, "/login")
			return
		}
		profile = userAccount.Profile
		skills = userAccount.Profile.Skills
		// userAccount.Profile.Projects undefined - Projects is on User, not Profile
		projects = userAccount.Projects // Corrected access
	}

	data := utils.GetTemplateData(c, isAuthenticated)
	data.Profile = profile
	data.Skills = skills
	data.Projects = projects
	c.HTML(http.StatusOK, "users/account.html", data)
}

// RenderEditAccountPage renders the edit account page
func (h *Handler) RenderEditAccountPage(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

	var profile domain.Profile

	if isAuthenticated {
		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			log.Printf("Invalid user ID in session: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to get user account")
			c.Redirect(http.StatusFound, "/login")
			return
		}
		retrievedProfile, err := h.UserUseCase.GetProfileByID(userID)
		if err != nil {
			log.Printf("Profile not found for user %s: %v", userID.String(), err)
			utils.SetFlashMessage(c, utils.FlashError, "Profile not found")
			c.Redirect(http.StatusFound, "/login")
			return
		}
		profile = *retrievedProfile
	}
	data := utils.GetTemplateData(c, isAuthenticated)
	data.Profile = profile
	c.HTML(http.StatusOK, "users/profile_form.html", data)
}

// RenderCreateSkillPage renders the create skill page
func (h *Handler) RenderCreateSkillPage(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

	data := utils.GetTemplateData(c, isAuthenticated)
	data.FormTitle = "Create Skill"
	c.HTML(http.StatusOK, "users/skill_form.html", data)
}

// RenderUpdateSkillPage renders the update skill page
func (h *Handler) RenderUpdateSkillPage(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

	var skill domain.Skill
	if isAuthenticated {
		_, err := uuid.Parse(userIDStr.(string)) // Removed userID assignment
		if err != nil {
			log.Printf("Invalid user ID in session: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to get skill")
			c.Redirect(http.StatusFound, "/login")
			return
		}
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			utils.SetFlashMessage(c, utils.FlashError, "Invalid skill ID")
			c.Redirect(http.StatusFound, "/account")
			return
		}
		retrievedSkill, err := h.UserUseCase.GetSkillByID(id)
		if err != nil {
			log.Printf("Skill not found for ID %s: %v", idStr, err)
			utils.SetFlashMessage(c, utils.FlashError, "Skill not found")
			c.Redirect(http.StatusFound, "/account")
			return
		}
		skill = *retrievedSkill
	}
	data := utils.GetTemplateData(c, isAuthenticated)
	data.FormTitle = "Update Skill"
	data.Skill = skill
	c.HTML(http.StatusOK, "users/skill_form.html", data)
}

// RenderDeleteSkillPage renders the delete skill page
func (h *Handler) RenderDeleteSkillPage(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

	var skill domain.Skill
	if isAuthenticated {
		_, err := uuid.Parse(userIDStr.(string)) // Removed userID assignment
		if err != nil {
			log.Printf("Invalid user ID in session: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to get skill")
			c.Redirect(http.StatusFound, "/login")
			return
		}
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			utils.SetFlashMessage(c, utils.FlashError, "Invalid skill ID")
			c.Redirect(http.StatusFound, "/account")
			return
		}
		retrievedSkill, err := h.UserUseCase.GetSkillByID(id)
		if err != nil {
			log.Printf("Skill not found for ID %s: %v", idStr, err)
			utils.SetFlashMessage(c, utils.FlashError, "Skill not found")
			c.Redirect(http.StatusFound, "/account")
			return
		}
		skill = *retrievedSkill
	}
	data := utils.GetTemplateData(c, isAuthenticated)
	data.Object = skill
	c.HTML(http.StatusOK, "delete.html", data)
}

// RenderInboxPage renders the inbox page
func (h *Handler) RenderInboxPage(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

	var messages []domain.Message
	var unreadCount int64

	if isAuthenticated {
		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			log.Printf("Invalid user ID in session: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to get inbox")
			c.Redirect(http.StatusFound, "/login")
			return
		}
		messages, _, err = h.UserUseCase.GetInbox(userID)
		if err != nil {
			log.Printf("Error fetching messages for recipient %s: %v", userID.String(), err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to fetch inbox")
			c.Redirect(http.StatusFound, "/account") // Redirect to account or home
			return
		}
	}

	data := utils.GetTemplateData(c, isAuthenticated)
	data.MessageRequests = messages
	data.UnreadCount = unreadCount
	c.HTML(http.StatusOK, "users/inbox.html", data)
}

// RenderMessagePage renders a single message
func (h *Handler) RenderMessagePage(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

	var message domain.Message
	if isAuthenticated {
		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			log.Printf("Invalid user ID in session: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to get message")
			c.Redirect(http.StatusFound, "/login")
			return
		}
		messageIDStr := c.Param("id")
		messageID, err := uuid.Parse(messageIDStr)
		if err != nil {
			utils.SetFlashMessage(c, utils.FlashError, "Invalid message ID")
			c.Redirect(http.StatusFound, "/inbox")
			return
		}
		retrievedMessage, err := h.UserUseCase.GetMessage(messageID, userID)
		if err != nil {
			log.Printf("Message not found or unauthorized for user %s, message %s: %v", userID.String(), messageIDStr, err)
			utils.SetFlashMessage(c, utils.FlashError, "Message not found or you don't have permission")
			c.Redirect(http.StatusFound, "/inbox")
			return
		}
		message = *retrievedMessage // Assign the dereferenced message
	}

	data := utils.GetTemplateData(c, isAuthenticated)
	data.Message = message
	c.HTML(http.StatusOK, "users/message.html", data)
}

// RenderCreateMessagePage renders the create message page
func (h *Handler) RenderCreateMessagePage(c *gin.Context) {
	recipientIDStr := c.Param("id")
	recipientID, err := uuid.Parse(recipientIDStr)
	if err != nil {
		log.Printf("Invalid recipient ID: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Invalid recipient ID")
		c.Redirect(http.StatusFound, "/profiles") // Redirect to profiles or error page
		return
	}

	recipient, err := h.UserUseCase.GetProfileByID(recipientID)
	if err != nil {
		log.Printf("Recipient not found for ID %s: %v", recipientID.String(), err)
		utils.SetFlashMessage(c, utils.FlashError, "Recipient not found")
		c.Redirect(http.StatusFound, "/profiles") // Redirect to profiles or error page
		return
	}

	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

	data := utils.GetTemplateData(c, isAuthenticated)
	data.Recipient = *recipient // Dereference
	c.HTML(http.StatusOK, "users/message_form.html", data)
}

// RenderProfilesPage renders the list of profiles
func (h *Handler) RenderProfilesPage(c *gin.Context) {
	// Get authenticated user ID for template rendering
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

	// Search logic
	searchQuery := c.Query("search_query")

	// Pagination logic
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := 3 // Items per page, consistent with Django project

	profiles, totalProfiles, err := h.UserUseCase.GetAllProfiles(searchQuery, page, limit)
	if err != nil {
		log.Printf("Error fetching profiles: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to load profiles")
		c.Redirect(http.StatusFound, "/")
		return
	}

	pagination := utils.Paginate(c, int(totalProfiles), limit)

	data := utils.GetTemplateData(c, isAuthenticated)
	data.Profiles = profiles
	data.SearchQuery = searchQuery
	data.Pagination = pagination
	c.HTML(http.StatusOK, "users/index.html", data)
}

// RenderUserProfilePage renders a single user profile
func (h *Handler) RenderUserProfilePage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SetFlashMessage(c, utils.FlashError, "Invalid profile ID")
		c.Redirect(http.StatusFound, "/profiles")
		return
	}

	profile, err := h.UserUseCase.GetProfileByID(id)
	if err != nil {
		log.Printf("Profile not found for ID %s: %v", idStr, err)
		utils.SetFlashMessage(c, utils.FlashError, "Profile not found")
		c.Redirect(http.StatusFound, "/profiles")
		return
	}

	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil
	var currentUserID uuid.UUID
	if isAuthenticated {
		currentUserID, _ = uuid.Parse(userIDStr.(string))
	}

	// Filter skills into top and other based on description existence
	var topSkills []domain.Skill
	var otherSkills []domain.Skill

	for _, skill := range profile.Skills {
		if skill.Description != "" {
			topSkills = append(topSkills, skill)
		} else {
			otherSkills = append(otherSkills, skill)
		}
	}

	data := utils.GetTemplateData(c, isAuthenticated)
	data.Profile = *profile // Dereference
	data.TopSkills = topSkills
	data.OtherSkills = otherSkills
	data.CurrentUserID = currentUserID
	data.IsOwner = (currentUserID == profile.UserID) // Determine if authenticated user is the owner
	c.HTML(http.StatusOK, "users/profile.html", data)
}

// RenderLoginRegisterPage renders the login/register page
func (h *Handler) RenderLoginRegisterPage(c *gin.Context) {
	pageType := c.Request.URL.Path[1:] // "login" or "register"

	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

	if isAuthenticated {
		c.Redirect(http.StatusFound, "/profiles")
		return
	}

	data := utils.GetTemplateData(c, isAuthenticated)
	data.Page = pageType
	c.HTML(http.StatusOK, "users/login_register.html", data)
}
