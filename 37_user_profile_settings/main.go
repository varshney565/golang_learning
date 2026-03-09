// Package main demonstrates user profile settings management in Go.
// Topics: struct design, functional options, validation, JSON persistence,
//         immutable updates, change tracking.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// -----------------------------------------------------------------------
// SECTION 1: Profile Data Model
// -----------------------------------------------------------------------
// A UserProfile holds display preferences and account settings.
// JSON tags control serialization; omitempty skips zero values.

type Theme string

const (
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
	ThemeSystem Theme = "system"
)

type NotificationSettings struct {
	Email    bool `json:"email"`
	Push     bool `json:"push"`
	SMS      bool `json:"sms"`
	Newsletter bool `json:"newsletter,omitempty"`
}

type UserProfile struct {
	ID           string               `json:"id"`
	Username     string               `json:"username"`
	DisplayName  string               `json:"display_name"`
	Email        string               `json:"email"`
	Bio          string               `json:"bio,omitempty"`
	Theme        Theme                `json:"theme"`
	Timezone     string               `json:"timezone"`
	Language     string               `json:"language"`
	Notifications NotificationSettings `json:"notifications"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

// -----------------------------------------------------------------------
// SECTION 2: Validation
// -----------------------------------------------------------------------
// Validate returns an error if the profile contains invalid data.
// errors.Join (Go 1.20+) collects multiple validation errors at once.

var ErrInvalidProfile = errors.New("invalid profile")

func (p *UserProfile) Validate() error {
	var errs []error

	if strings.TrimSpace(p.Username) == "" {
		errs = append(errs, fmt.Errorf("username: must not be empty"))
	}
	if len(p.Username) > 32 {
		errs = append(errs, fmt.Errorf("username: must be 32 characters or fewer"))
	}
	if !strings.Contains(p.Email, "@") {
		errs = append(errs, fmt.Errorf("email: %q is not a valid address", p.Email))
	}
	switch p.Theme {
	case ThemeLight, ThemeDark, ThemeSystem:
		// valid
	default:
		errs = append(errs, fmt.Errorf("theme: %q is not one of light|dark|system", p.Theme))
	}
	if len(p.Bio) > 160 {
		errs = append(errs, fmt.Errorf("bio: must be 160 characters or fewer (got %d)", len(p.Bio)))
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %w", ErrInvalidProfile, errors.Join(errs...))
	}
	return nil
}

// -----------------------------------------------------------------------
// SECTION 3: Functional Options Pattern for Updates
// -----------------------------------------------------------------------
// ProfileOption is a function that mutates a UserProfile.
// Callers compose only the changes they need; unchanged fields stay intact.

type ProfileOption func(*UserProfile)

func WithDisplayName(name string) ProfileOption {
	return func(p *UserProfile) { p.DisplayName = name }
}

func WithBio(bio string) ProfileOption {
	return func(p *UserProfile) { p.Bio = bio }
}

func WithTheme(t Theme) ProfileOption {
	return func(p *UserProfile) { p.Theme = t }
}

func WithTimezone(tz string) ProfileOption {
	return func(p *UserProfile) { p.Timezone = tz }
}

func WithLanguage(lang string) ProfileOption {
	return func(p *UserProfile) { p.Language = lang }
}

func WithNotifications(n NotificationSettings) ProfileOption {
	return func(p *UserProfile) { p.Notifications = n }
}

// UpdateProfile applies options to a copy of the profile, validates the result,
// and returns the updated profile.  The original is never modified.
func UpdateProfile(current UserProfile, opts ...ProfileOption) (UserProfile, error) {
	updated := current // value copy — original is unchanged
	for _, opt := range opts {
		opt(&updated)
	}
	updated.UpdatedAt = time.Now().UTC()

	if err := updated.Validate(); err != nil {
		return current, fmt.Errorf("update rejected: %w", err)
	}
	return updated, nil
}

// -----------------------------------------------------------------------
// SECTION 4: Change Tracking
// -----------------------------------------------------------------------
// Diff compares two profiles and returns human-readable change lines.
// Useful for audit logs or "you changed X" notifications.

func Diff(before, after UserProfile) []string {
	var changes []string
	check := func(field, a, b string) {
		if a != b {
			changes = append(changes, fmt.Sprintf("  %s: %q → %q", field, a, b))
		}
	}
	check("display_name", before.DisplayName, after.DisplayName)
	check("bio", before.Bio, after.Bio)
	check("theme", string(before.Theme), string(after.Theme))
	check("timezone", before.Timezone, after.Timezone)
	check("language", before.Language, after.Language)
	return changes
}

// -----------------------------------------------------------------------
// SECTION 5: JSON Round-Trip (Persistence)
// -----------------------------------------------------------------------
// Serialize to JSON for storage; deserialize to restore.

func serializeDemo(p UserProfile) {
	fmt.Println("\nJSON persistence:")

	data, err := json.MarshalIndent(p, "  ", "  ")
	if err != nil {
		fmt.Printf("  marshal error: %v\n", err)
		return
	}
	fmt.Printf("  serialized:\n  %s\n", data)

	var restored UserProfile
	if err := json.Unmarshal(data, &restored); err != nil {
		fmt.Printf("  unmarshal error: %v\n", err)
		return
	}
	fmt.Printf("  restored username: %s, theme: %s\n", restored.Username, restored.Theme)
}

// -----------------------------------------------------------------------
// SECTION 6: Putting It All Together
// -----------------------------------------------------------------------

func main() {
	// Create the initial profile
	profile := UserProfile{
		ID:          "usr_001",
		Username:    "alice",
		DisplayName: "Alice",
		Email:       "alice@example.com",
		Theme:       ThemeLight,
		Timezone:    "UTC",
		Language:    "en",
		Notifications: NotificationSettings{
			Email: true,
			Push:  false,
		},
		UpdatedAt: time.Now().UTC(),
	}

	fmt.Println("=== User Profile Settings ===")
	fmt.Printf("\nInitial profile: %s (%s) theme=%s lang=%s\n",
		profile.DisplayName, profile.Email, profile.Theme, profile.Language)

	// --- Valid update ---
	fmt.Println("\n--- Apply valid settings update ---")
	updated, err := UpdateProfile(profile,
		WithDisplayName("Alice Smith"),
		WithBio("Go developer. Loves concurrency."),
		WithTheme(ThemeDark),
		WithLanguage("fr"),
		WithNotifications(NotificationSettings{Email: true, Push: true, SMS: false}),
	)
	if err != nil {
		fmt.Printf("  error: %v\n", err)
	} else {
		fmt.Printf("  updated: %s, theme=%s, lang=%s, bio=%q\n",
			updated.DisplayName, updated.Theme, updated.Language, updated.Bio)

		changes := Diff(profile, updated)
		fmt.Println("  changes:")
		for _, c := range changes {
			fmt.Println(c)
		}
	}

	// --- Validation failure: bad email + bio too long ---
	fmt.Println("\n--- Apply invalid update (should be rejected) ---")
	_, err = UpdateProfile(profile,
		WithDisplayName("Bob"),
		// email not changed, username stays "alice"
		WithBio(strings.Repeat("x", 200)), // too long
		WithTheme("neon"),                 // invalid theme
	)
	if err != nil {
		fmt.Printf("  rejected (expected): %v\n", err)
		fmt.Printf("  is ErrInvalidProfile: %v\n", errors.Is(err, ErrInvalidProfile))
	}

	// --- JSON persistence round-trip ---
	serializeDemo(updated)
}
