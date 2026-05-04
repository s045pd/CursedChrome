package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// All models use UUID primary keys with column name "id" and Sequelize-style
// "createdAt"/"updatedAt" timestamp columns so existing Postgres data
// (cursedchrome-db volume) keeps working without migration.

type BaseUUID struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;column:id"`
	CreatedAt time.Time `gorm:"column:createdAt;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updatedAt;autoUpdateTime"`
}

// BeforeCreate auto-assigns a UUID v4 if none was set.
func (b *BaseUUID) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// User mirrors Sequelize Users table (database.js:Users).
type User struct {
	BaseUUID
	Username                 string `gorm:"type:text;uniqueIndex;column:username"`
	Password                 string `gorm:"type:text;column:password"`
	PasswordShouldBeChanged  bool   `gorm:"not null;default:false;column:password_should_be_changed"`
}

func (User) TableName() string { return "users" }

// Bot mirrors Sequelize Bots table.
type Bot struct {
	BaseUUID
	BrowserID       string     `gorm:"type:text;not null;index;column:browser_id"`
	Name            string     `gorm:"type:text;not null;default:'Untitled Proxy';column:name"`
	ProxyUsername   string     `gorm:"type:text;uniqueIndex;column:proxy_username"`
	ProxyPassword   string     `gorm:"type:text;not null;index;column:proxy_password"`
	IsOnline        bool       `gorm:"not null;default:true;column:is_online"`
	LastOnline      time.Time  `gorm:"not null;column:last_online"`
	CurrentTab      JSONMap    `gorm:"type:jsonb;column:current_tab"`
	CurrentTabImage string     `gorm:"type:text;default:'';column:current_tab_image"`
	Tabs            JSONArray  `gorm:"type:jsonb;column:tabs"`
	History         JSONArray  `gorm:"type:jsonb;column:history"`
	Cookies         JSONArray  `gorm:"type:jsonb;column:cookies"`
	State           string     `gorm:"type:text;default:'';column:state"`
	Bookmarks       JSONArray  `gorm:"type:jsonb;column:bookmarks"`
	Downloads       JSONArray  `gorm:"type:jsonb;column:downloads"`
	SwitchConfig    JSONMap    `gorm:"type:jsonb;column:switch_config"`
	DataConfig      JSONMap    `gorm:"type:jsonb;column:data_config"`
	UserAgent       string     `gorm:"type:text;column:user_agent"`
	Recording       JSONArray  `gorm:"type:jsonb;column:recording"`
	Screenshots     JSONArray  `gorm:"type:jsonb;column:screenshots"`
	Activity        JSONArray  `gorm:"type:jsonb;column:activity"`
	LastActiveAt    *time.Time `gorm:"column:last_active_at"`
}

func (Bot) TableName() string { return "bots" }

// BotRecording mirrors Sequelize BotRecording table.
type BotRecording struct {
	BaseUUID
	Recording string     `gorm:"type:text;not null;column:recording"`
	Bot       uuid.UUID  `gorm:"type:uuid;index;column:bot"`
	Text      string     `gorm:"type:text;column:text"`
	Timestamp *time.Time `gorm:"column:timestamp"`
	SessionID string     `gorm:"type:text;column:session_id"`
}

func (BotRecording) TableName() string { return "bot_recordings" }

// BotScreenshot mirrors Sequelize BotScreenshots table.
type BotScreenshot struct {
	BaseUUID
	BotID      uuid.UUID `gorm:"type:uuid;index;column:bot_id"`
	URL        string    `gorm:"type:text;column:url"`
	Title      string    `gorm:"type:text;column:title"`
	ImageData  string    `gorm:"type:text;not null;column:image_data"`
	SessionID  string    `gorm:"type:text;column:session_id"`
	Difference *float64  `gorm:"column:difference"`
	Timestamp  time.Time `gorm:"not null;index;column:timestamp"`
}

func (BotScreenshot) TableName() string { return "bot_screenshots" }

// BotKeyboardLog mirrors Sequelize BotKeyboardLogs table.
type BotKeyboardLog struct {
	BaseUUID
	BotID     uuid.UUID `gorm:"type:uuid;index;column:bot_id"`
	URL       string    `gorm:"type:text;column:url"`
	Title     string    `gorm:"type:text;column:title"`
	Keys      string    `gorm:"type:text;not null;column:keys"`
	Timestamp time.Time `gorm:"not null;index;column:timestamp"`
}

func (BotKeyboardLog) TableName() string { return "bot_keyboard_logs" }

// Setting mirrors Sequelize Settings table.
type Setting struct {
	BaseUUID
	Key   string `gorm:"type:text;uniqueIndex;column:key"`
	Value string `gorm:"type:text;column:value"`
}

func (Setting) TableName() string { return "settings" }

// All returns the slice of model pointers used for AutoMigrate.
func All() []any {
	return []any{
		&User{},
		&Bot{},
		&BotRecording{},
		&BotScreenshot{},
		&BotKeyboardLog{},
		&Setting{},
	}
}
