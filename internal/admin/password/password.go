package password

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/bcrypt"
)

var (
	service              string = "nextgate"
	adminPasswordService string = "admin-password"
)

var OldPasswordNotCorrect = errors.New("admin: old password is not correct")

// CheckAdminPassword reports whether an admin password is stored in the system
// keyring. A missing entry is not treated as an error.
func CheckAdminPassword() (bool, error) {
	_, err := keyring.Get(service, adminPasswordService)

	if errors.Is(err, keyring.ErrNotFound) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("admin: keyring get error: %w", err)
	}

	return true, nil
}

// GetAdminPassword returns the stored admin password hash from the system
// keyring. It returns an error when no admin password is set or the keyring
// cannot be accessed.
func GetAdminPassword() (string, error) {
	adminPassword, err := keyring.Get(service, adminPasswordService)

	if err != nil {
		return "", fmt.Errorf("admin: keyring get error: %w", err)
	}

	return adminPassword, nil
}

func SetAdminPassword(newAdminPassword string) error {

	if len([]byte(newAdminPassword)) > 72 {
		return errors.New("admin: password cannot exceed 72 bytes")
	}

	adminPassword, err := keyring.Get(service, adminPasswordService)

	if errors.Is(err, keyring.ErrNotFound) {
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(newAdminPassword),
			12,
		)

		if err != nil {
			return fmt.Errorf("admin: bcrypt hash error: %w", err)
		}

		err = keyring.Set(
			service,
			adminPasswordService,
			string(hashedPassword),
		)

		if err != nil {
			return fmt.Errorf("admin: keyring set error: %w", err)
		}

		return nil
	}

	if err != nil {
		return fmt.Errorf("admin: keyring get error: %w", err)
	}

	var oldPassword string

	fmt.Print("Enter old password: ")
	fmt.Scan(&oldPassword)

	err = bcrypt.CompareHashAndPassword(
		[]byte(adminPassword),
		[]byte(oldPassword),
	)

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return OldPasswordNotCorrect
		}

		return fmt.Errorf("admin: bcrypt compare error: %w", err)
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword(
		[]byte(newAdminPassword),
		12,
	)

	if err != nil {
		return fmt.Errorf("admin: bcrypt hash error: %w", err)
	}

	err = keyring.Set(service, adminPasswordService, string(hashedNewPassword))

	if err != nil {
		return fmt.Errorf("admin: keyring set error: %w", err)
	}

	return nil
}
