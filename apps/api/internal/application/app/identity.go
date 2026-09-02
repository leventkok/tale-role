package app

import (
	"github.com/leventkok/tale-role/apps/api/internal/domain/iam"
	"github.com/leventkok/tale-role/apps/api/internal/domain/license"
)

// Identity is the account store. Memory or Mongo both implement it.
type Identity interface {
	PutUser(*iam.User)
	GetUser(email string) (*iam.User, bool)
	GetUserByID(id string) (*iam.User, bool)
	PutOTP(*iam.OTP)
	GetOTP(email string) (*iam.OTP, bool)
	DeleteOTP(email string)
	PutLicense(*license.ProductLicense)
	LicensesForUser(userID string) []*license.ProductLicense
	DeleteUserByID(id string)
}
