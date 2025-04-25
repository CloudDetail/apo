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
	adminPasswd = "APO2024@admin"

	userFieldSql = "user_id, username, phone, email, corporation"
)

type User struct {
	UserID      int64  `gorm:"column:user_id;primary_key" json:"userId,omitempty"`
	Username    string `gorm:"column:username;uniqueIdx;type:varchar(20)" json:"username,omitempty"`
	Password    string `gorm:"column:password;type:varchar(200)" json:"-"`
	Phone       string `gorm:"column:phone;type:varchar(20)" json:"phone,omitempty"`
	Email       string `gorm:"column:email;type:varchar(50)" json:"email,omitempty"`
	Corporation string `gorm:"column:corporation;type:varchar(50)" json:"corporation,omitempty"`

	RoleList    []Role    `gorm:"many2many:user_role;joinForeignKey:UserID;joinReferences:RoleID" json:"roleList,omitempty"`
	TeamList    []Team    `gorm:"many2many:user_team;joinForeignKey:UserID;joinReferences:TeamID" json:"teamList,omitempty"`
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
				Password: Encrypt(adminPasswd),
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
		return nil, model.NewErrWithMessage(errors.New("password incorrect"), code.UserPasswdIncorrectError)
	}

	return &user, nil
}

func (repo *daoRepo) CreateUser(ctx context.Context, user *User) error {
	db := repo.GetContextDB(ctx)
	var count int64
	err := db.Model(&User{}).Where("username = ?", user.Username).Count(&count).Error
	if err != nil {
		return err
	}
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
		return model.NewErrWithMessage(errors.New("password incorrect"), code.UserPasswdIncorrectError)
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

func (repo *daoRepo) UpdateUserInfo(ctx context.Context, userID int64, phone string, email string, corporation string) error {
	var user User
	err := repo.GetContextDB(ctx).Where("user_id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewErrWithMessage(errors.New("user does not exist "), code.UserNotExistsError)
	} else if err != nil {
		return err
	}

	user.Corporation = corporation
	user.Phone = phone
	user.Email = email
	return repo.GetContextDB(ctx).Updates(&user).Error
}

func (repo *daoRepo) GetUserInfo(userID int64) (User, error) {
	var user User
	err := repo.db.
		Select(userFieldSql).
		Preload("RoleList").
		Preload("TeamList").
		Where("user_id = ?", userID).
		First(&user).Error
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

	query := repo.db.Model(&User{}).Preload("RoleList").Preload("TeamList")

	if len(req.Username) > 0 {
		query = query.Where("username LIKE ?", fmt.Sprintf("%%%s%%", req.Username))
	}

	if len(req.RoleList) > 0 {
		subQuery := repo.db.Table("user_role").
			Select("user_id").
			Where("role_id IN ?", req.RoleList)
		query = query.Where("user_id IN (?)", subQuery)
	}

	if len(req.TeamList) > 0 {
		subQuery := repo.db.Table("user_team").
			Select("user_id").
			Where("team_id IN ?", req.TeamList)
		query = query.Where("user_id IN (?)", subQuery)
	}

	if len(req.Corporation) > 0 {
		query = query.Where("corporation LIKE ?", fmt.Sprintf("%%%s%%", req.Corporation))
	}

	query = query.Where("username != ?", config.Get().User.AnonymousUser.Username)

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	query = query.Limit(req.PageSize).Offset((req.CurrentPage - 1) * req.PageSize)

	err = query.Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, count, nil
}

func (repo *daoRepo) RemoveUser(ctx context.Context, userID int64) error {
	err := repo.GetContextDB(ctx).Model(&User{}).Where("user_id = ?", userID).Delete(nil).Error
	if err != nil {
		return err
	}

	err = repo.GetContextDB(ctx).
		Model(&UserRole{}).
		Where("user_id = ?", userID).
		Delete(nil).
		Error

	if err != nil {
		return err
	}

	return repo.GetContextDB(ctx).
		Model(&UserTeam{}).
		Where("user_id = ?", userID).
		Delete(nil).
		Error
}

func (repo *daoRepo) UserExists(userIDs ...int64) (bool, error) {
	var count int64
	if err := repo.db.Model(&User{}).Where("user_id IN ?", userIDs).Count(&count).Error; err != nil {
		return false, err
	}

	return count == int64(len(userIDs)), nil
}
