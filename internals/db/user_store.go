package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

// The User Store Implements the interface of a User_store
type UserStore struct {
	db *sql.DB
}

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
	CreatedAt     string  `json:"created_at"`
}

// User update Options
type UpdatePayload struct {
	Id        string  `json:"id"` // taking id in payload to get the target user
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
	Gender    *string `json:"gender"`
	// TODO : Update this to make it more secure
	NewPass *string `json:"new_pass"`
}

// Returns the user object pointer and and error, searching by their ID
func (u *UserStore) GetUserById(ctx context.Context, id string) (*User, error) {
	// filter by id
	// join the roles table to get the name of the role
	query := `
		SELECT u.id, u.first_name, u.last_name, u.email, u.password, u.gender,
		u.age, r.name, u.created_at
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
		&user.CreatedAt,
	)
	if err != nil {
		// for debugging: Comment out later
		log.Printf("DB Query Error for Id(%s): %v\n\n", id, err.Error())
		return nil, err
	}

	return &user, nil
}

// Function to return user pointer and error filtering by email
func (u *UserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	// filter by email and join the roles table to get the role name
	query := `
		SELECT u.id, u.first_name, u.last_name, u.email,
		u.password, u.gender, u.age, r.name, u.created_at
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
		&user.CreatedAt,
	)

	if err != nil {
		// For Debugging: Comment out later
		log.Printf("DB QUery Error for MailID(%s): %v\n", email, err.Error())
		return nil, err
	}

	return &user, nil
}

// Function to create a new user
func (u *UserStore) CreateUser(ctx context.Context, user *User) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error Beginning transaction: %v\n", err.Error())
		return err
	}
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
	// TODO: Open an account for user
	// TODO: Add their Platform Links
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
	// TODO : Clean up the links related to the user
	// query = `
	// 	DELETE FROM platform_links
	// 	WHERE id = $1
	// `
	// _, err = tx.ExecContext(ctx, query, id)
	// if err != nil {
	// 	return err
	// }
	// TODO: Delete the users bank account
	// I am not deleting the submissions
	// TODO: Replace the userid in submissions to NA
	tx.Commit()
	return nil
}

func (u *UserStore) UpdateUser(ctx context.Context, payload UpdatePayload) error {
	var p PassW // To hash the new password if given
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
	if payload.NewPass != nil {
		p.Hash(*payload.NewPass)
		setClauses = append(setClauses, fmt.Sprintf("password = $%d", i))
		args = append(args, p.hashed_pass)
		i++
	}

	if len(setClauses) == 0 {
		return errors.New("no fields to update")
	}

	queryBuilder.WriteString(strings.Join(setClauses, ", "))
	queryBuilder.WriteString(fmt.Sprintf(" WHERE id = $%d", i))
	args = append(args, payload.Id)

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
