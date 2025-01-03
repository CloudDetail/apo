// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"gorm.io/gorm"
	"strconv"
)

const (
	adminPassword = "APO2024@admin"

	userFieldSql = "user_id, username, role, phone, email, corporation"
)

type User struct {
	UserID      int64  `gorm:"column:user_id;primary_key" json:"userId,omitempty"`
	Username    string `gorm:"column:username;uniqueIdx" json:"username,omitempty"`
	Password    string `gorm:"column:password" json:"-"`
	Role        string `gorm:"column:role" json:"role,omitempty"`
	Phone       string `gorm:"column:phone" json:"phone,omitempty"`
	Email       string `gorm:"column:email" json:"email,omitempty"`
	Corporation string `gorm:"column:corporation" json:"corporation,omitempty"`

	RoleList    []Role    `gorm:"-" json:"roleList,omitempty"`
	FeatureList []Feature `gorm:"-" json:"featureList,omitempty"`
}

func (t *User) TableName() string {
	return "user"
}

func (repo *daoRepo) createAdmin() error {
	var admin User
	err := repo.db.Where("username = ?", model.ROLE_ADMIN).First(&admin).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			admin = User{
				UserID:   util.Generator.GenerateID(),
				Username: model.ROLE_ADMIN,
				Password: Encrypt(adminPassword),
			}

			if err = repo.db.Create(&admin).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	var role Role
	if err = repo.db.Where("role_name = ?", model.ROLE_ADMIN).First(&role).Error; err != nil {
		return err
	}
	userRole := &UserRole{
		UserID: admin.UserID,
		RoleID: role.RoleID,
	}
	return repo.db.Save(userRole).Error
}

func (repo *daoRepo) createAnonymousUser() error {
	conf := config.Get().User.AnonymousUser
	var anonymousUser User
	err := repo.db.Where("username = ?", conf.Username).First(&anonymousUser).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			anonymousUser = User{
				UserID:   util.Generator.GenerateID(),
				Username: conf.Username,
				// random password
				Password: Encrypt(strconv.FormatInt(util.Generator.GenerateID(), 10)),
			}

			if err = repo.db.Create(&anonymousUser).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	var role Role
	if err = repo.db.Where("role_name = ?", conf.Role).First(&role).Error; err != nil {
		return err
	}

	var existingUserRole UserRole
	err = repo.db.Where("user_id = ?", anonymousUser.UserID).First(&existingUserRole).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userRole := &UserRole{
				UserID: anonymousUser.UserID,
				RoleID: role.RoleID,
			}
			err = repo.db.Create(userRole).Error
		} else {
			return err
		}
	} else {
		err = repo.db.Model(&existingUserRole).Update("role_id", role.RoleID).Error
	}

	return err
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

func (repo *daoRepo) CreateUser(ctx context.Context, user *User) error {
	db := repo.GetContextDB(ctx)
	var count int64
	db.Model(&User{}).Where("username = ?", user.Username).Count(&count)
	if count > 0 {
		return model.NewErrWithMessage(errors.New("user already exists"), code.UserAlreadyExists)
	}
	user.Password = Encrypt(user.Password)
	return db.Create(user).Error
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

func (repo *daoRepo) GetAnonymousUser() (User, error) {
	conf := config.Get().User.AnonymousUser
	var user User
	err := repo.db.Select(userFieldSql).Where("username = ?", conf.Username).Find(&user).Error
	return user, err
}

func (repo *daoRepo) GetUserList(req *request.GetUserListRequest) ([]User, int64, error) {
	var users []User
	var count int64
	query := repo.db.Select(userFieldSql)
	if len(req.Username) > 0 {
		query = query.Where("username like ?", fmt.Sprintf("%%%s%%", req.Username))
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

func (repo *daoRepo) RemoveUser(ctx context.Context, userID int64) error {
	return repo.GetContextDB(ctx).Model(&User{}).Where("user_id = ?", userID).Delete(nil).Error
}

func (repo *daoRepo) UserExists(userID int64) (bool, error) {
	var count int64
	if err := repo.db.Model(&User{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
