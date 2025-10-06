package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
)

const (
	MinPassLen = 8
)

// The User Store Implements the interface of a User_store
type UserStore struct {
	db *sql.DB
}

type LinkStore struct {
	db *sql.DB
}

// Links model
type Links struct {
	Platform string `json:"platform"`
	Url      string `json:"url"`
}

// The User Model
type User struct {
	Id            string  `json:"id"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	Password      PassW   `json:"-"`
	Gender        string  `json:"gender"`
	Age           int     `json:"age"`
	Role          string  `json:"role"`
	PlatformLinks []Links `json:"links"`
	IsVerified    bool    `json:"is_verified"`
	CreatedAt     string  `json:"created_at"`
}

// User update Options
type UpdatePayload struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
	Gender    *string `json:"gender"`
}

// Function to change the user password
// Validate the uuid before calling the function
func (u *UserStore) ChangePassword(ctx context.Context, id, new_pass string) error {
	var pw PassW
	if len(new_pass) < 8 {
		log.Printf("error: want pass len: %d, got: %d", MinPassLen, len(new_pass))
		return ErrPasswordTooShort
	}
	if err := pw.Hash(new_pass); err != nil {
		// logging for debugging
		log.Printf("error hashing password: %v\n", err.Error())
		return err
	}
	query := `
		UPDATE users
		SET password = $1
		WHERE id = $2
	`
	res, err := u.db.ExecContext(ctx, query, pw.hashed_pass, id)
	if err != nil {
		log.Printf("error changing password: %v\n", err.Error())
		return err
	}
	if count, _ := res.RowsAffected(); count == 0 {
		log.Printf("invalid id: %v\n", id)
		return ErrInvalidId
	}
	return nil
}

// Function to mark a user verified
func (u *UserStore) VerifyUser(ctx context.Context, entity, id string) error {
	if err := uuid.Validate(id); err != nil {
		log.Printf("error invalid id: %v", err.Error())
		return ErrInvalidId
	}
	user_query := `
		UPDATE users
		SET is_verified = true
		WHERE id = $1
	`
	brand_query := `
		UPDATE brands
		SET is_verified = true
		WHERE id = $1
	`
	switch entity {
	case "users":
		res, err := u.db.ExecContext(ctx, user_query, id)
		if err != nil {
			log.Printf("invalid user id")
			return err
		}
		if count, _ := res.RowsAffected(); count == 0 {
			return ErrInvalidId
		}
	case "brands":
		res, err := u.db.ExecContext(ctx, brand_query, id)
		if err != nil {
			log.Printf("invalid brand id")
			return err
		}
		if count, _ := res.RowsAffected(); count == 0 {
			return ErrInvalidId
		}
	default:
		return fmt.Errorf("invalid credentials")
	}
	// successfully verified
	return nil
}

// Returns the user object pointer and and error, searching by their ID
func (u *UserStore) GetUserById(ctx context.Context, id string) (*User, error) {
	// filter by id
	// join the roles table to get the name of the role
	query := `
		SELECT u.id, u.first_name, u.last_name, u.email, u.password, u.gender,
		u.age, r.name, u.is_verified, u.created_at
		FROM users u
		JOIN roles r ON r.id = u.role
		WHERE u.id = $1
	`
	var user User
	// Querying the user by id and scanning the values into the object
	err := u.db.QueryRowContext(ctx, query, id).Scan(
		&user.Id,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password.hashed_pass,
		&user.Gender,
		&user.Age,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
	)
	if err != nil {
		// for debugging: Comment out later
		log.Printf("DB Query Error for Id(%s): %v\n", id, err.Error())
		return nil, err
	}

	link_store := &LinkStore{db: u.db}
	user.PlatformLinks = link_store.GetLinks(ctx, user.Id)

	return &user, nil
}

// Function to return user pointer and error filtering by email
func (u *UserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	// filter by email and join the roles table to get the role name
	query := `
		SELECT u.id, u.first_name, u.last_name, u.email,
		u.password, u.gender, u.age, r.name, u.is_verified, u.created_at
		FROM users u
		JOIN roles r ON r.id = u.role
		WHERE u.email = $1
	`
	// Get the user and scan it into the object
	var user User
	err := u.db.QueryRowContext(ctx, query, email).Scan(
		&user.Id,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password.hashed_pass,
		&user.Gender,
		&user.Age,
		&user.Role,
		&user.IsVerified,
		&user.CreatedAt,
	)

	if err != nil {
		// For Debugging: Comment out later
		log.Printf("DB Query Error for MailID(%s): %v\n", email, err.Error())
		return nil, err
	}
	link_store := &LinkStore{db: u.db}
	user.PlatformLinks = link_store.GetLinks(ctx, user.Id)
	return &user, nil
}

// Function to create a new user
func (u *UserStore) CreateUser(ctx context.Context, user *User) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error Beginning transaction: %v\n", err.Error())
		return err
	}
	defer tx.Rollback()
	query := `
		INSERT INTO users (id, first_name, last_name, email, password, gender, age, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = tx.ExecContext(ctx, query,
		user.Id,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password.hashed_pass,
		user.Gender,
		user.Age,
		user.Role,
	)
	if err != nil {
		// For Debugging
		log.Printf("Error creating a user: %v\n", err.Error())
		return err
	}
	// Open an account for user
	// Ask them to open an account only when they want to join a campaign
	// Add their Platform Links
	if user.PlatformLinks != nil {
		var link_store *LinkStore = &LinkStore{db: u.db}
		err = link_store.AddLinks(ctx, user.Id, user.PlatformLinks)
		if err != nil {
			log.Printf("error adding platform links: %v", err.Error())
			return err
		}
	}
	tx.Commit()
	return nil
}

// Function to Delete a user
func (u *UserStore) DeleteUser(ctx context.Context, id string) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error Beginning transaction: %v\n", err.Error())
		return err
	}
	defer tx.Rollback()
	_, err = u.GetUserById(ctx, id)
	if err != nil {
		return err
	}
	query := `
		DELETE FROM users
		WHERE id = $1
	`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Clean up the links related to the user
	query = `
		DELETE FROM platform_links
		WHERE userid = $1
	`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Delete the users bank account
	query = `
		DELETE FROM accounts
		WHERE holder_id = $1
	`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Delete all the related submissions
	query = `
		DELETE FROM submissions
		WHERE creator_id = $1
	`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Cleaned up the User related fields
	tx.Commit()
	return nil
}

func (u *UserStore) UpdateUser(ctx context.Context, id string, payload UpdatePayload) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting a transaction: %v\n", err.Error())
		return err
	}
	// String buffer
	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE users SET ")

	args := []any{}          // update arguments
	setClauses := []string{} // update clauses

	i := 1 // postgres placeholders start from $1

	if payload.FirstName != nil {
		setClauses = append(setClauses, fmt.Sprintf("first_name = $%d", i))
		args = append(args, *payload.FirstName)
		i++
	}
	if payload.LastName != nil {
		setClauses = append(setClauses, fmt.Sprintf("last_name = $%d", i))
		args = append(args, *payload.LastName)
		i++
	}
	if payload.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", i))
		args = append(args, *payload.Email)
		i++
	}
	if payload.Gender != nil {
		setClauses = append(setClauses, fmt.Sprintf("gender = $%d", i))
		args = append(args, *payload.Gender)
		i++
	}

	if len(setClauses) == 0 {
		return errors.New("no fields to update")
	}

	queryBuilder.WriteString(strings.Join(setClauses, ", "))
	queryBuilder.WriteString(fmt.Sprintf(" WHERE id = $%d", i))
	args = append(args, id)

	query := queryBuilder.String()
	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		// Error executing the query
		tx.Rollback()
		return err
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		// Wrong user Id
		tx.Rollback()
		return sql.ErrNoRows
	}
	// Successfully updated the user information
	tx.Commit()
	return nil
}

// ----------  LinkStore Implementation ---------------

func (l *LinkStore) AddLinks(ctx context.Context, id string, links []Links) error {
	query := `
		INSERT INTO platform_links (userid, platform, url)
		VALUES ($1, $2, $3)
	`
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		// error starting a transaction
		log.Printf("error beginning transaction: %v", err.Error())
		return err
	}
	for _, v := range links {
		_, err := tx.ExecContext(ctx, query, id, v.Platform, v.Url)
		if err != nil {
			// rollback the inserted links
			tx.Rollback()
			log.Printf("error inserting links: %v", err.Error())
			return err
		}
	}

	// successfully inserted the links
	tx.Commit()
	return nil
}

// Deletes a link associated to a creator
func (l *LinkStore) DeleteLinks(ctx context.Context, id string, platform string) error {
	query := `
		DELETE FROM platform_links
		WHERE userid = $1 AND platform = $2
	`
	if id == "" || platform == "" {
		return fmt.Errorf("invalid id or platform")
	}
	_, err := l.db.ExecContext(ctx, query, id, platform)
	if err != nil {
		log.Printf("error deleting links: %v", err.Error())
		return err
	}
	// Successfully deleted the link
	return nil
}

func (l *LinkStore) GetLinks(ctx context.Context, id string) []Links {
	var output []Links
	query := `
		SELECT platform, url
		FROM platform_links
		WHERE userid = $1
	`
	rows, err := l.db.QueryContext(ctx, query, id)
	if err != nil {
		log.Printf("error: %v", err.Error())
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var link Links
		if err = rows.Scan(
			&link.Platform,
			&link.Url,
		); err != nil {
			log.Printf("error fetching links: %v\n", err.Error())
			return nil
		}
		output = append(output, link)
	}
	return output
}
