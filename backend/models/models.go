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
	// Permissions names the gated services this non-admin user may use, one
	// key each from services.PermissionTypes (admins pass every check
	// regardless). Unlike Admin it is never copied into a token: the gate
	// reads it from the row, which services.Auth loads on every request
	// anyway, so a grant or revoke takes effect on the user's next request.
	// Stored as a JSON array in one text column; a row from before the
	// column existed reads as NULL, i.e. no permissions.
	Permissions []string `gorm:"serializer:json" json:"permissions"`
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

// Box is one panel on a canvas page. Page names the canvas it belongs to:
// "index" for the landing page, or a Page.Slug for an admin-created page
// (the home page is the Page "stp") — it is a key, not a foreign key, so the
// built-in landing canvas needs no Page row. Its position and size are in px on a
// fixed 1400px-wide design canvas, which the page scales down on narrower
// screens — so the stored numbers mean the same thing for every visitor. Kind "widget" renders a built-in Vue component named by Widget;
// kind "html" renders HTML in a sandboxed iframe on the front end, which is
// what makes storing admin-authored script here acceptable.
type Box struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	// default:stp is what the boxes that predate pages are migrated to.
	Page   string  `gorm:"not null;default:stp;index" json:"page"`
	Kind   string  `gorm:"not null" json:"kind"`
	Widget *string `json:"widget"`
	HTML   *string `json:"html"`
	Title  *string `json:"title"`
	X      int     `gorm:"not null;default:0" json:"x"`
	Y      int     `gorm:"not null;default:0" json:"y"`
	W      int     `gorm:"not null;default:300" json:"w"`
	H      int     `gorm:"not null;default:200" json:"h"`
	Z      int     `gorm:"not null;default:0" json:"z"`
	// No gorm default on Framed: GORM omits zero values for defaulted columns
	// on insert, so default:true would turn an explicit false into true.
	// CreateBox applies the true default itself.
	Framed       bool `gorm:"not null" json:"framed"`
	HideOnMobile bool `gorm:"not null;default:false" json:"hideOnMobile"`
	// AdminOnly boxes are never sent to anyone but an admin (the boxes
	// resolver filters them out), so they suit private notes.
	AdminOnly bool `gorm:"not null;default:false" json:"adminOnly"`
	// ComponentID links an html box to a Component, whose HTML it shows
	// instead of its own: the row's HTML is nil, and the resolvers fill it in
	// from the component before answering (graph/component_helpers.go). There
	// is deliberately no association field — with one, DB.Save(&box) would
	// write the component back too.
	ComponentID *uint `gorm:"index" json:"componentId"`
}

// Page is an admin-created canvas page, served by the SPA's catch-all route
// at /<Slug> (e.g. "shrines/evangelion"). Its content is the Boxes whose Page
// equals Slug. The unique index is partial so that a deleted page's slug can
// be reused — a plain uniqueIndex would still collide with the soft-deleted
// row.
type Page struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Slug      string         `gorm:"not null;index:idx_pages_slug,unique,where:deleted_at IS NULL" json:"slug"`
	Title     string         `gorm:"not null" json:"title"`
	// AdminOnly hides the whole page, boxes included, from everyone else.
	AdminOnly bool `gorm:"not null;default:false" json:"adminOnly"`
}

// Background is the page background behind one canvas (Page is the same page
// key a Box has: "index" for the landing page, or a Page's slug), so
// built-in canvases and admin-created pages are set the same way. A page with
// no row shows the theme's halftone; deleting the row is "reset to default".
// The unique index is partial for the same reason as Page's.
type Background struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Page      string         `gorm:"not null;index:idx_backgrounds_page,unique,where:deleted_at IS NULL" json:"page"`
	// Color is a CSS colour, or "" for the theme's page colour.
	Color string `gorm:"not null;default:''" json:"color"`
	// Image is a URL, or "" for none. Drawn over Color.
	Image string `gorm:"not null;default:''" json:"image"`
	// Fit is how Image fills the window: "cover", "contain" or "tile".
	Fit string `gorm:"not null;default:cover" json:"fit"`
}

// Layout is a named copy of one canvas page's boxes and background, saved so
// it can be loaded onto any page later (graph/layout_helpers.go). It is a
// snapshot, not a link: loading it creates new boxes that belong to the
// target page, so editing either page afterwards never touches the other or
// the layout. The content is JSON rather than child rows because nothing
// queries inside it — it is only ever written whole and read whole. The
// unique index is partial for the same reason as Page's.
type Layout struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Name      string         `gorm:"not null;index:idx_layouts_name,unique,where:deleted_at IS NULL" json:"name"`
	// SourcePage is the page key it was saved from, for the list only.
	SourcePage string `gorm:"not null" json:"sourcePage"`
	// Boxes is a JSON array of the boxes' content and geometry, without id
	// or page.
	Boxes string `gorm:"type:text;not null" json:"-"`
	// Background is the page's background as JSON, or nil if it had none.
	Background *string `gorm:"type:text" json:"-"`
}

// Component is a named HTML box that can be placed on any canvas page, either
// as a copy (an ordinary box) or linked (a box with ComponentID set). Linked
// boxes share only the HTML — editing it changes every one of them at once —
// while title, geometry, frame and visibility stay each box's own, since
// those are about where it sits on a page rather than what it is. Title, W,
// H, Framed and AdminOnly are the defaults a newly placed box starts with.
// Widgets are not components: a widget box is just a name, with nothing
// worth sharing. The unique index is partial for the same reason as Page's.
type Component struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Name      string         `gorm:"not null;index:idx_components_name,unique,where:deleted_at IS NULL" json:"name"`
	Title     *string        `json:"title"`
	HTML      string         `gorm:"type:text;not null" json:"html"`
	W         int            `gorm:"not null" json:"w"`
	H         int            `gorm:"not null" json:"h"`
	// No gorm default on Framed, for the same reason as Box.Framed.
	Framed    bool `gorm:"not null" json:"framed"`
	AdminOnly bool `gorm:"not null;default:false" json:"adminOnly"`
}
