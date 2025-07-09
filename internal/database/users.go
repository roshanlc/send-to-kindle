package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// GetUserByID retreives a user by id
func (db *DB) GetUserByID(userID int) (User, error) {
	query := `SELECT name,email,password,smtp_to,added_at FROM users WHERE id = ?;`

	result := db.Database.QueryRow(query, userID)

	if err := result.Err(); err != nil {
		return User{}, err
	}

	user := User{
		ID: userID,
	}

	var smtpTo sql.NullString

	err := result.Scan(&user.Name,
		&user.Email,
		&user.Password,
		&smtpTo,
		&user.AddedAt,
	)

	if err != nil {
		return User{}, err
	}

	if smtpTo.Valid {
		user.SmtpTo = toReceipientArray(smtpTo.String)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(userEmail string) (User, error) {
	if userEmail == "" {
		return User{}, fmt.Errorf("please provide a valid non-empty email")
	}
	query := `SELECT id,name,email,password,smtp_to,added_at FROM users WHERE email = ?;`

	result := db.Database.QueryRow(query, userEmail)

	if err := result.Err(); err != nil {
		return User{}, err
	}

	var user User
	var smtpTo sql.NullString

	err := result.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&smtpTo,
		&user.AddedAt,
	)

	if err != nil {
		return User{}, err
	}

	if smtpTo.Valid {
		user.SmtpTo = toReceipientArray(smtpTo.String)
	}

	return user, nil
}

func toReceipientArray(value string) []string {
	return strings.Split(value, ",")
}

func fromReceipientArray(value []string) string {
	return strings.Join(value, ",")
}

func validateUser(user *User) error {
	if user.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if user.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if user.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	return nil
}

// AddUser adds a user to the table
func (db *DB) AddUser(user User) (int, error) {
	err := validateUser(&user)
	if err != nil {
		return 0, err
	}
	tx, err := db.Database.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `INSERT INTO users (name,email,password,smtp_to) VALUES(?,?,?,?);`

	result, err := tx.Exec(query,
		user.Name,
		user.Email,
		user.Password,
		fromReceipientArray(user.SmtpTo),
	)

	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// DeleteUser deletes a user by id
func (db *DB) DeleteUser(userID int) error {
	tx, err := db.Database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM users WHERE id=?;`

	result, err := tx.Exec(query, userID)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	r, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// check the count of rows affected
	if r == 0 {
		return ErrNoRowDeleted
	}
	return nil
}

// UpdateUser updates a user row. Supports updating the name, email and password of the user.
// Only provide value for the property to be updated. Keep them empty if field is not be updated.
func (db *DB) UpdateUser(user User) error {
	if user.ID == 0 {
		return fmt.Errorf("userID cannot be empty")
	}
	tx, err := db.Database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		queryParts []string
		args       []any
	)

	if user.Email != "" {
		queryParts = append(queryParts, "email = ?")
		args = append(args, user.Email)
	}

	if user.Name != "" {
		queryParts = append(queryParts, "name = ?")
		args = append(args, user.Name)
	}

	if user.Password != "" {
		queryParts = append(queryParts, "password = ?")
		args = append(args, user.Password)
	}

	if len(user.SmtpTo) != 0 {
		queryParts = append(queryParts, "smtp_to = ?")
		args = append(args, fromReceipientArray(user.SmtpTo))
	}

	query := fmt.Sprintf(`UPDATE users SET %s WHERE id=?;`,
		strings.Join(queryParts, ","))

	args = append(args, user.ID)
	result, err := tx.Exec(query, args...)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	r, err := result.RowsAffected()
	if err != nil {
		return nil
	}

	// check the count of rows affected
	if r == 0 {
		return ErrNoRowDeleted
	}
	return nil
}
