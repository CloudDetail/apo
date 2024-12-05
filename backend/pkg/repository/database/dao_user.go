package database

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	RoleAdmin     = "admin"
	adminUsername = "admin"
	adminPassword = "APO2024@admin"

	userFieldSql = "user_id, username, role, phone, email, corporation"
)

type User struct {
	UserID      int64  `gorm:"user_id;primary_key" json:"userId,omitempty"`
	Username    string `gorm:"username;unique" json:"username,omitempty"`
	Password    string `gorm:"password" json:"-"`
	Role        string `gorm:"role" json:"role,omitempty"`
	Phone       string `gorm:"phone" json:"phone,omitempty"`
	Email       string `gorm:"email" json:"email,omitempty"`
	Corporation string `gorm:"corporation" json:"corporation,omitempty"`
}

func (t *User) TableName() string {
	return "user"
}

func createAdmin(db *gorm.DB) error {
	admin := &User{
		Username: adminUsername,
		Password: Encrypt(adminPassword),
		Role:     RoleAdmin,
	}
	var count int64
	db.Model(&User{}).Where("username = ?", adminUsername).Count(&count)
	if count > 0 {
		return nil
	}
	return db.Create(&admin).Error
}

func createAnonymousUser(db *gorm.DB) error {
	conf := config.Get().User.AnonymousUser
	anonymousUser := &User{
		Username: conf.Username,
		Role:     conf.Role,
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"role": anonymousUser.Role}),
	}).Create(&anonymousUser).Error
}

func Encrypt(raw string) string {
	h := md5.New()
	h.Write([]byte(raw))
	return hex.EncodeToString(h.Sum(nil))
}

func (repo *daoRepo) Login(username, password string) (*User, error) {
	var user User
	err := repo.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, model.NewErrWithMessage(errors.New("user does not exists"), code.UserNotExistsError)
	} else if err != nil {
		return nil, err
	}
	enPassword := Encrypt(password)
	if enPassword != user.Password {
		return nil, model.NewErrWithMessage(errors.New("password incorrect"), code.UserPasswordIncorrectError)
	}

	return &user, nil
}

func (repo *daoRepo) CreateUser(user *User) error {
	var count int64
	repo.db.Model(&User{}).Where("username = ?", user.Username).Count(&count)
	if count > 0 {
		return model.NewErrWithMessage(errors.New("user already exists"), code.UserAlreadyExists)
	}
	user.Password = Encrypt(user.Password)
	return repo.db.Create(user).Error
}

func (repo *daoRepo) UpdateUserPhone(userID int64, phone string) error {
	var user User
	err := repo.db.Where("user_id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewErrWithMessage(errors.New("user does not exists"), code.UserNotExistsError)
	} else if err != nil {
		return err
	}
	user.Phone = phone
	return repo.db.Updates(&user).Error
}

func (repo *daoRepo) UpdateUserEmail(userID int64, email string) error {
	var user User
	err := repo.db.Where("user_id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewErrWithMessage(errors.New("user does not exists"), code.UserNotExistsError)
	} else if err != nil {
		return err
	}
	user.Email = email
	return repo.db.Updates(&user).Error
}

func (repo *daoRepo) UpdateUserPassword(userID int64, oldPassword, newPassword string) error {
	var user User
	err := repo.db.Where("user_id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewErrWithMessage(errors.New("user does not exists"), code.UserNotExistsError)
	} else if err != nil {
		return err
	}
	if user.Password != Encrypt(oldPassword) {
		return model.NewErrWithMessage(errors.New("password incorrect"), code.UserPasswordIncorrectError)
	}
	user.Password = Encrypt(newPassword)
	return repo.db.Updates(&user).Error
}

func (repo *daoRepo) RestPassword(userID int64, newPassword string) error {
	var user User
	err := repo.db.Where("user_id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewErrWithMessage(errors.New("user does not exists"), code.UserNotExistsError)
	} else if err != nil {
		return err
	}

	user.Password = Encrypt(newPassword)
	return repo.db.Updates(&user).Error
}

func (repo *daoRepo) UpdateUserInfo(req *request.UpdateUserInfoRequest) error {
	var user User
	err := repo.db.Where("user_id = ?", req.UserID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewErrWithMessage(errors.New("user does not exist "), code.UserNotExistsError)
	} else if err != nil {
		return err
	}

	user.Corporation = req.Corporation
	user.Phone = req.Phone
	user.Email = req.Email
	return repo.db.Updates(&user).Error
}

func (repo *daoRepo) GetUserInfo(userID int64) (User, error) {
	var user User
	err := repo.db.Select(userFieldSql).Where("user_id = ?", userID).First(&user).Error
	return user, err
}

func (repo *daoRepo) GetUserList(req *request.GetUserListRequest) ([]User, int64, error) {
	var users []User
	var count int64
	query := repo.db.Select(userFieldSql)
	if len(req.Username) > 0 {
		name := "%" + req.Username + "%"
		query = query.Where("username like ?", name)
	}
	if len(req.Role) > 0 {
		query = query.Where("role = ?", req.Role)
	}
	if len(req.Corporation) > 0 {
		corporation := "%" + req.Corporation + "%"
		query = query.Where("corporation like ?", corporation)
	}
	query = query.Where("username != ?", "anonymous")
	err := query.Model(&User{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	query = query.Limit(req.PageSize).Offset((req.CurrentPage - 1) * req.PageSize)
	err = query.Find(&users).Error
	return users, count, err
}

func (repo *daoRepo) RemoveUser(userID int64, operatorID int64) error {
	var operator User
	if err := repo.db.Select("role").Where("user_id = ?", operatorID).First(&operator).Error; err != nil {
		return err
	}

	if operator.Role != RoleAdmin {
		return model.NewErrWithMessage(errors.New("no permission"), code.UserNoPermissionError)
	}

	return repo.db.Model(&User{}).Where("user_id = ?", userID).Delete(nil).Error
}
