package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name      string    `gorm:"size:255;not null"`
	Email     string    `gorm:"size:255;not null;unique"`
	Username  string    `gorm:"size:255;not null;unique"`
	Password  string    `gorm:"size:255;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Profile   Profile   `gorm:"foreignKey:UserID"`
	Projects  []Project `gorm:"foreignKey:OwnerID"`
	Messages  []Message `gorm:"foreignKey:RecipientID"`
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	return
}

func (user *User) HashPassword(password string) {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	user.Password = string(bytes)
}

func (user *User) CheckPasswordHash(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

type Profile struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;unique"`
	Name           string    `gorm:"size:255"`
	Email          string    `gorm:"size:255"`
	Username       string    `gorm:"size:255"`
	Location       string    `gorm:"size:255"`
	ShortIntro     string    `gorm:"size:255"`
	Bio            string
	ProfileImage   string  `gorm:"size:255;default:'user-default.png'"`
	SocialGithub   string  `gorm:"size:255"`
	SocialLinkedin string  `gorm:"size:255"`
	SocialWebsite  string  `gorm:"size:255"`
	Skills         []Skill `gorm:"foreignKey:OwnerID"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (profile *Profile) BeforeCreate(tx *gorm.DB) (err error) {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	return
}

type Skill struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null"`
	Name        string    `gorm:"size:255;not null"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (skill *Skill) BeforeCreate(tx *gorm.DB) (err error) {
	if skill.ID == uuid.Nil {
		skill.ID = uuid.New()
	}
	return
}

type Message struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Sender      User      `gorm:"foreignKey:SenderID"`
	SenderID    uuid.UUID `gorm:"type:uuid"`
	Recipient   User      `gorm:"foreignKey:RecipientID"`
	RecipientID uuid.UUID `gorm:"type:uuid"`
	Name        string    `gorm:"size:255;not null"`
	Email       string    `gorm:"size:255;not null"`
	Subject     string    `gorm:"size:255;not null"`
	Body        string    `gorm:"not null"`
	IsRead      bool      `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (message *Message) BeforeCreate(tx *gorm.DB) (err error) {
	if message.ID == uuid.Nil {
		message.ID = uuid.New()
	}
	return
}

type Project struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Owner         User      `gorm:"foreignKey:OwnerID"`
	OwnerID       uuid.UUID `gorm:"type:uuid"`
	Title         string    `gorm:"size:255;not null"`
	Description   string    `gorm:"not null"`
	FeaturedImage string    `gorm:"size:255;default:'default.jpg'"`
	DemoLink      string    `gorm:"size:255"`
	SourceLink    string    `gorm:"size:255"`
	Tags          []Tag     `gorm:"many2many:project_tags;"`
	Reviews       []Review  `gorm:"foreignKey:ProjectID"`
	VoteTotal     int       `gorm:"default:0"`
	VoteRatio     int       `gorm:"default:0"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (project *Project) BeforeCreate(tx *gorm.DB) (err error) {
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}
	return
}

type Tag struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name      string    `gorm:"size:255;not null"`
	Projects  []Project `gorm:"many2many:project_tags;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (tag *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	if tag.ID == uuid.Nil {
		tag.ID = uuid.New()
	}
	return
}

type Review struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Project   Project   `gorm:"foreignKey:ProjectID"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null"`
	Owner     User      `gorm:"foreignKey:OwnerID"`
	OwnerID   uuid.UUID `gorm:"type:uuid;not null"`
	Body      string    `gorm:"not null"`
	Value     string    `gorm:"size:255;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (review *Review) BeforeCreate(tx *gorm.DB) (err error) {
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	return
}
