package types

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	DateOfBirth  time.Time `json:"date_of_birth"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TokenPair struct {
	AccessToken        string    `json:"-"`
	AccessTokenExpiry  time.Time `json:"-"`
	RefreshToken       string    `json:"-"`
	RefreshTokenExpiry time.Time `json:"-"`
}

type AuthUser struct {
	User      User      `json:"user"`
	TokenPair TokenPair `json:"token_pair"`
}

type RegisterUser struct {
	Username    string    `json:"username" validate:"required,min=3,max=50"`
	FirstName   string    `json:"first_name" validate:"required,min=3,max=100"`
	LastName    string    `json:"last_name" validate:"required,min=3,max=100"`
	DateOfBirth time.Time `json:"date_of_birth" validate:"required,date_of_birth"`
	Email       string    `json:"email" validate:"required,email"`
	Password    string    `json:"password" validate:"required,min=8,max=50"`
}

type UpdateUser struct {
	Username    *string    `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	FirstName   *string    `json:"first_name,omitempty" validate:"omitempty,min=3,max=100"`
	LastName    *string    `json:"last_name,omitempty" validate:"omitempty,min=3,max=100"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty" validate:"omitempty,date_of_birth"`
}

type UpdatePassword struct {
	OldPassword string `json:"old_password" validate:"required,min=8,max=50"`
	Password    string `json:"password" validate:"required,min=8,max=50"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Project struct {
	ID          string    `json:"id"`
	CreatorID   string    `json:"creator_id"`
	AdminIDs    []string  `json:"admin_ids"`
	MembersIDs  []string  `json:"members_ids"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Columns *[]KanbanColumn `json:"columns,omitempty"`
}

type ProjectAdmin struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}

type ProjectMember struct {
	ProjectID string `json:"project_id"`
	UserID    string `json:"user_id"`
}

type CreateProject struct {
	AdminEmails []string `json:"admin_emails,omitempty" validate:"omitempty,dive,email"`
	MemberEmails []string `json:"member_emails,omitempty" validate:"omitempty,dive,email"`
	AdminIDs []string `json:"-"`
	MemberIDs []string `json:"-"`
	Name        string   `json:"name" validate:"required,min=3,max=50"`
	Description string   `json:"description" validate:"max=500"`
}

type UpdateProject struct {
	AdminIDs    *[]string `json:"admin_ids,omitempty" validate:"omitempty,dive,uuid"`
	MembersIDs  *[]string `json:"members_ids,omitempty" validate:"omitempty,dive,uuid"`
	Name        *string   `json:"name,omitempty" validate:"omitempty,min=3,max=50"`
	Description *string   `json:"description,omitempty" validate:"omitempty,max=500"`
}

type KanbanColumn struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Order     int    `json:"order"`

	Label *KanbanColumnLabel `json:"label,omitempty"`
	Rows  *[]KanbanRow       `json:"rows,omitempty"`
}

type CreateKanbanColumn struct {
	ProjectID string `json:"project_id" validate:"required,uuid"`
	Name      string `json:"name" validate:"required,min=3,max=50"`
	Order     int    `json:"order" validate:"required,min=0,max=20"`

	LabelID *string `json:"label_id,omitempty" validate:"omitempty,uuid"`
}

type UpdateKanbanColumn struct {
	ProjectID   string  `json:"project_id" validate:"required,uuid"`
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=50"`
	Order       *int    `json:"order,omitempty" validate:"omitempty,min=0,max=20"`
	LabelID     *string `json:"label_id,omitempty" validate:"omitempty,uuid"`
	DeleteLabel *bool   `json:"delete_label,omitempty" validate:"omitempty,oneof=true false"`
}

type SpecialTag string

const (
	ToDoTag       SpecialTag = "TODO"
	InProgressTag SpecialTag = "IN_PROGRESS"
	TestingTag    SpecialTag = "TESTING"
	CompletedTag  SpecialTag = "COMPLETED"
)

type KanbanColumnLabel struct {
	ID         string      `json:"id"`
	ProjectID  string      `json:"project_id"`
	SpecialTag *SpecialTag `json:"special_tag"`
	Name       string      `json:"name"`
	Color      int         `json:"color"`
}

type CreateKanbanColumnLabel struct {
	ProjectID string `json:"project_id" validate:"required,uuid"`
	Name      string `json:"name" validate:"required,min=3,max=50"`
	Color     int    `json:"color" validate:"required,min=0,max=16777215"`
}

type UpdateKanbanColumnLabel struct {
	ProjectID string  `json:"project_id" validate:"required,uuid"`
	Name      *string `json:"name,omitempty" validate:"omitempty,min=3,max=50"`
	Color     *int    `json:"color,omitempty" validate:"omitempty,min=0,max=16777215"`
}

type Priority string

const (
	LowPriority    Priority = "LOW"
	MediumPriority Priority = "MEDIUM"
	HighPriority   Priority = "HIGH"
)

type KanbanRow struct {
	ID               string     `json:"id"`
	ColumnID         string     `json:"column_id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Order            int        `json:"order"`
	CreatorID        string     `json:"creator_id"`
	AssignedUsersIDs []string   `json:"assigned_users_ids"`
	Priority         *Priority  `json:"priority"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DueDate          *time.Time `json:"due_date"`

	Label          *KanbanRowLabel `json:"label,omitempty"`
	History        *[]HistoryPoint `json:"history,omitempty"`
	Checklist      *Checklist      `json:"check_list,omitempty"`
	CommentSection *CommentSection `json:"comment_section,omitempty"`
}

type KanbanRowAssignedUser struct {
	RowID  string `json:"row_id"`
	UserID string `json:"user_id"`
}

type CreateKanbanRow struct {
	ColumnID         string     `json:"column_id" validate:"required,uuid"`
	Name             string     `json:"name" validate:"required,min=3,max=50"`
	Description      string     `json:"description,omitempty" validate:"omitempty,min=0,max=200"`
	Order            int        `json:"order" validate:"required,min=0,max=20"`
	AssignedUsersIDs []string   `json:"assigned_users_ids,omitempty" validate:"omitempty,dive,uuid"`
	Priority         *Priority  `json:"priority,omitempty" validate:"omitempty,min=3,max=20"`
	DueDate          *time.Time `json:"due_date,omitempty" validate:"omitempty,due_date"`

	LabelID *string `json:"label_id,omitempty" validate:"omitempty,uuid"`
}

type UpdateKanbanRow struct {
	ProjectID        string     `json:"project_id" validate:"required,uuid"`
	ColumnID         string     `json:"column_id" validate:"required,uuid"`
	Name             *string    `json:"name,omitempty" validate:"omitempty,min=3,max=50"`
	Description      *string    `json:"description,omitempty" validate:"omitempty,min=3,max=200"`
	Order            *int       `json:"order,omitempty" validate:"omitempty,min=0,max=20"`
	AssignedUsersIDs *[]string  `json:"assigned_users_ids,omitempty" validate:"omitempty,dive,uuid"`
	Priority         *Priority  `json:"priority,omitempty" validate:"omitempty,min=3,max=20"`
	DueDate          *time.Time `json:"due_date,omitempty" validate:"omitempty,due_date"`

	LabelID     *string `json:"label_id,omitempty" validate:"omitempty,uuid"`
	DeleteLabel *bool   `json:"delete_label,omitempty" validate:"omitempty,oneof=true false"`
}

type KanbanRowLabel struct {
	ProjectID string `json:"project_id"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Color     int    `json:"color"`
}

type CreateKanbanRowLabel struct {
	ProjectID string `json:"project_id" validate:"required,uuid"`
	Name      string `json:"name" validate:"required,min=3,max=50"`
	Color     int    `json:"color" validate:"required,min=0,max=16777215"`
}

type UpdateKanbanRowLabel struct {
	ProjectID string  `json:"project_id" validate:"required,uuid"`
	Name      *string `json:"name,omitempty" validate:"omitempty,min=3,max=50"`
	Color     *int    `json:"color,omitempty" validate:"omitempty,min=0,max=16777215"`
}

type HistoryPoint struct {
	RowID     string    `json:"row_id"`
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type Checklist struct {
	ID    string `json:"id"`
	RowID string `json:"row_id"`

	Points *[]Point `json:"points,omitempty"`
}

type Point struct {
	ChecklistID string     `json:"checklist_id"`
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at"`
	CompletedBy *string    `json:"completed_by"`
}

type CreatePoint struct {
	ProjectID   string `json:"project_id" validate:"required,uuid"`
	ChecklistID string `json:"checklist_id" validate:"required,uuid"`
	Name        string `json:"name" validate:"required,min=3,max=50"`
	Description string `json:"description" validate:"required,min=3,max=100"`
}

type UpdatePoint struct {
	ProjectID   string  `json:"project_id" validate:"required,uuid"`
	ChecklistID string  `json:"checklist_id" validate:"required,uuid"`
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=50"`
	Description *string `json:"description,omitempty" validate:"omitempty,min=3,max=100"`

	Completed   *bool      `json:"-"`
	CompletedBy *string    `json:"-"`
	CompletedAt *time.Time `json:"-"`
}

type CommentSection struct {
	ID         string `json:"id"`
	RowID      string `json:"row_id"`
	CanComment bool   `json:"can_comment"`

	Comments *[]Comment `json:"comments"`
}

type Comment struct {
	ID               string    `json:"id"`
	CommentSectionID string    `json:"comment_section_id"`
	UserID           string    `json:"user_id"`
	Text             string    `json:"text"`
	CreatedAt        time.Time `json:"created_at"`
}

type CreateComment struct {
	CommentSectionID string `json:"comment_section_id" validate:"required,uuid"`
	Text             string `json:"text" validate:"required,min=3,max=200"`
}

type DBCredentials struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}
