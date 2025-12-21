package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	t.Run("should hash password successfully", func(t *testing.T) {
		password := RandomString(6)
		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)
		require.NotEmpty(t, hashedPassword)
		require.NotEqual(t, password, hashedPassword)
	})

	t.Run("should generate different hashes for same password", func(t *testing.T) {
		password := RandomString(6)
		hashedPassword1, err := HashPassword(password)
		require.NoError(t, err)

		hashedPassword2, err := HashPassword(password)
		require.NoError(t, err)

		// bcrypt generates different hashes each time due to salt
		require.NotEqual(t, hashedPassword1, hashedPassword2)
	})

	t.Run("should handle empty password", func(t *testing.T) {
		hashedPassword, err := HashPassword("")
		require.Error(t, err)
		require.Empty(t, hashedPassword)
	})
}

func TestCheckPassword(t *testing.T) {
	t.Run("should verify correct password", func(t *testing.T) {
		password := RandomString(6)
		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)

		err = CheckPassword(password, hashedPassword)
		require.NoError(t, err)
	})

	t.Run("should reject incorrect password", func(t *testing.T) {
		password := RandomString(6)
		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)

		wrongPassword := RandomString(6)
		err = CheckPassword(wrongPassword, hashedPassword)
		require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
	})

	t.Run("should reject empty password", func(t *testing.T) {
		password := RandomString(6)
		hashedPassword, err := HashPassword(password)
		require.NoError(t, err)

		err = CheckPassword("", hashedPassword)
		require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
	})
}
