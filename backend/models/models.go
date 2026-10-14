package models

import (
	"time"

	"gorm.io/gorm"
)

// Package models holds the GORM entities, which are also the GraphQL object
// types: gqlgen.yml maps each GraphQL type straight onto the struct here, so
// graph/model/ only contains generated input and payload types. That means
// these structs carry three sets of contracts at once — database columns,
// JSON field names for the REST endpoints, and the shape resolvers return —
// and renaming a field or changing a tag breaks at least one of them.
//
// Conventions across every model:
//   - There are no migration files; services.migrateDatabase runs AutoMigrate,
//     so a model must be listed there to get a table.
//   - A gorm.DeletedAt field means soft delete: Delete sets deleted_at, and
//     every subsequent query silently filters those rows out. Rows are never
//     actually removed, and a uniqueIndex still sees deleted rows, so a
//     soft-deleted username cannot be reused.
//   - IDs are uint here and the ID scalar in GraphQL. gqlgen.yml binds the
//     two directly (graphql.UintID), so an `ID uint` field needs no resolver.
//     On the wire an id is a quoted decimal string, "12".

// User is an account. There is no public sign-up: users are created by an
// existing admin through the createUser mutation, or by the dev seeder.
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Username  string         `gorm:"uniqueIndex" json:"username"`
	// bcrypt hash, never the plaintext. `json:"-"` keeps it out of REST
	// responses, and gqlgen.yml marks the GraphQL password field
	// resolver:false so it is not exposed there either.
	Password []byte `json:"-"`
	// Admin gates almost every mutation. It is copied into the access token,
	// so a change here only takes effect on the user's next token refresh.
	Admin bool `json:"admin"`
	// TokenVersion is the session generation counter. Every issued JWT
	// carries the value current at the time it was minted, and
	// services.Auth rejects a token whose value no longer matches, so
	// incrementing this column revokes every outstanding token for the user
	// at once (logout, password change). It is the only server-side
	// revocation mechanism there is.
	//
	// Pre-migration rows: AutoMigrate adds the column with DEFAULT 1, but a
	// row that somehow ends up at 0 is normalised to 1 by
	// services.tokenVersionOf, as is a token minted before the "tv" claim
	// existed. So the migration invalidates nobody's existing session, and
	// the first bump moves a user to 2 and invalidates everything older.
	TokenVersion uint `gorm:"not null;default:1" json:"-"`
}

// Post is a blog entry. Author is loaded with Preload("Author") by the post
// resolvers; without that preload it is nil rather than an error.
type Post struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Title     string         `gorm:"not null" json:"title"`
	AuthorID  uint           `json:"-"`
	Author    *User          `gorm:"foreignKey:AuthorID" json:"author"`
	Content   string         `json:"content"`
}

// Message is one chat message. The table is capped at 50 rows by the
// WebSocket hub, which soft-deletes older ones on every insert.
//
// Two things to watch: Content is exposed as "text", not "content"; and
// AuthorID is a per-process connection counter from the WebSocket hub, NOT a
// User ID — chat is pseudonymous and the number is meaningless across
// restarts.
type Message struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Content  string `json:"text"`
	AuthorID uint   `json:"authorId"`
	FileURL  string `json:"fileUrl,omitempty"`
	// Private messages are visible only to admins; both the WebSocket
	// broadcast and the messages query filter on it.
	Private   bool           `gorm:"not null;default:false" json:"private"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

// Activity is a "currently doing" entry on the home page. Structurally
// identical to Favorite but kept separate so the two lists can diverge.
type Activity struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Type      string         `json:"type"`
	Name      string         `json:"name"`
	Link      *string        `json:"link"`
}

// Favorite is a "currently enjoying" entry on the home page.
type Favorite struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Type      string         `json:"type"`
	Name      string         `json:"name"`
	Link      *string        `json:"link"`
}

// Rowing is one rowing machine session, created by uploading a photo of the
// display (see handlers.CreateRowing).
//
// Date is the photo's EXIF capture time and acts as the deduplication key —
// it is not the row's creation time. Time is total seconds; Distance is
// metres; TimePer500m is seconds and derived from the other two; Calories is
// an estimate computed from distance, not read off the machine.
type Rowing struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Date        time.Time      `json:"date"`
	Time        uint64         `json:"time"`
	Distance    uint64         `json:"distance"`
	TimePer500m float64        `json:"timePer500m"`
	Calories    float64        `json:"calories"`
}

// Bookmark is a link on the bookmarks page, grouped by Category.
type Bookmark struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Category  string         `gorm:"not null" json:"category"`
	Name      string         `gorm:"not null" json:"name"`
	Link      string         `gorm:"not null" json:"link"`
}

// Place is an entry on the places-to-go list. Creating or editing one needs
// only a signed-in user, not an admin — the one exception to the
// admin-writes-everything rule.
type Place struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Title     string         `gorm:"not null" json:"title"`
	Location  *string        `json:"location"`
	Notes     *string        `json:"notes"`
	ImageURL  *string        `json:"imageUrl"`
	Category  *string        `json:"category"`
	Cost      *string        `json:"cost"`
	Priority  int            `gorm:"default:0" json:"priority"`
	Done      bool           `gorm:"default:false" json:"done"`
}
