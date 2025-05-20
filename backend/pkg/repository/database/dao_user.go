// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"gorm.io/gorm"
)

const (
	adminPasswd = "APO2024@admin"

	userFieldSql = "user_id, username, phone, email, corporation"
)

const AnonymousUsername = "anonymous"

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

func (repo *daoRepo) createAdmin(ctx core.Context) error {
	var admin User
	err := repo.GetContextDB(ctx).Where("username = ?", model.ROLE_ADMIN).First(&admin).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			admin = User{
				UserID:   util.Generator.GenerateID(),
				Username: model.ROLE_ADMIN,
				Password: Encrypt(adminPasswd),
			}

			if err = repo.GetContextDB(ctx).Create(&admin).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	var role Role
	if err = repo.GetContextDB(ctx).Where("role_name = ?", model.ROLE_ADMIN).First(&role).Error; err != nil {
		return err
	}
	userRole := &UserRole{
		UserID: admin.UserID,
		RoleID: role.RoleID,
	}
	return repo.GetContextDB(ctx).Save(userRole).Error
}

func (repo *daoRepo) createAnonymousUser(ctx core.Context) error {
	conf := config.Get().User.AnonymousUser
	var anonymousUser User
	err := repo.GetContextDB(ctx).Where("username = ?", AnonymousUsername).First(&anonymousUser).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			anonymousUser = User{
				UserID:   util.Generator.GenerateID(),
				Username: AnonymousUsername,
				// random password
				Password: Encrypt(strconv.FormatInt(util.Generator.GenerateID(), 10)),
			}

			if err = repo.GetContextDB(ctx).Create(&anonymousUser).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	var role Role
	if !isValidRoleName(conf.Role) {
		return errors.New("invalid role")
	}

	if err = repo.GetContextDB(ctx).Where("role_name = ?", conf.Role).First(&role).Error; err != nil {
		return err
	}

	if role.RoleID <= 0 {
		return errors.New("role does not exist")
	}

	if anonymousUser.UserID <= 0 {
		return errors.New("anonymous user does not exist")
	}

	var existingUserRole UserRole
	err = repo.GetContextDB(ctx).Where("user_id = ?", anonymousUser.UserID).First(&existingUserRole).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userRole := &UserRole{
				UserID: anonymousUser.UserID,
				RoleID: role.RoleID,
			}
			err = repo.GetContextDB(ctx).Create(userRole).Error
		} else {
			return err
		}
	} else {
		err = repo.GetContextDB(ctx).Model(&existingUserRole).Update("role_id", role.RoleID).Error
	}

	return err
}

func Encrypt(raw string) string {
	h := md5.New()
	h.Write([]byte(raw))
	return hex.EncodeToString(h.Sum(nil))
}

func (repo *daoRepo) Login(ctx core.Context, username, password string) (*User, error) {
	var user User
	err := repo.GetContextDB(ctx).Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, core.Error(code.UserNotExistsError, "user does not exists")
	} else if err != nil {
		return nil, err
	}
	enPassword := Encrypt(password)
	if enPassword != user.Password {
		return nil, core.Error(code.UserPasswdIncorrectError, "password incorrect")
	}

	return &user, nil
}

func (repo *daoRepo) CreateUser(ctx core.Context, user *User) error {
	db := repo.GetContextDB(ctx)
	var count int64
	err := db.Model(&User{}).Where("username = ?", user.Username).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return core.Error(code.UserAlreadyExists, "user already exists")
	}
	user.Password = Encrypt(user.Password)
	return db.Create(user).Error
}

func (repo *daoRepo) UpdateUserPhone(ctx core.Context, userID int64, phone string) error {
	var user User
	err := repo.GetContextDB(ctx).Where("user_id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.Error(code.UserNotExistsError, "user does not exists")
	} else if err != nil {
		return err
	}
	user.Phone = phone
	return repo.GetContextDB(ctx).Updates(&user).Error
}

func (repo *daoRepo) UpdateUserEmail(ctx core.Context, userID int64, email string) error {
	var user User
	err := repo.GetContextDB(ctx).Where("user_id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.Error(code.UserNotExistsError, "user does not exists")
	} else if err != nil {
		return err
	}
	user.Email = email
	return repo.GetContextDB(ctx).Updates(&user).Error
}

func (repo *daoRepo) UpdateUserPassword(ctx core.Context, userID int64, oldPassword, newPassword string) error {
	var user User
	err := repo.GetContextDB(ctx).Where("user_id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.Error(code.UserNotExistsError, "user does not exists")
	} else if err != nil {
		return err
	}
	if user.Password != Encrypt(oldPassword) {
		return core.Error(code.UserPasswdIncorrectError, "password incorrect")
	}
	user.Password = Encrypt(newPassword)
	return repo.GetContextDB(ctx).Updates(&user).Error
}

func (repo *daoRepo) RestPassword(ctx core.Context, userID int64, newPassword string) error {
	var user User
	err := repo.GetContextDB(ctx).Where("user_id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.Error(code.UserNotExistsError, "user does not exists")
	} else if err != nil {
		return err
	}

	user.Password = Encrypt(newPassword)
	return repo.GetContextDB(ctx).Updates(&user).Error
}

func (repo *daoRepo) UpdateUserInfo(ctx core.Context, userID int64, phone string, email string, corporation string) error {
	var user User
	err := repo.GetContextDB(ctx).Where("user_id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.Error(code.UserNotExistsError, "user does not exist ")
	} else if err != nil {
		return err
	}

	user.Corporation = corporation
	user.Phone = phone
	user.Email = email
	return repo.GetContextDB(ctx).Updates(&user).Error
}

func (repo *daoRepo) GetUserInfo(ctx core.Context, userID int64) (User, error) {
	var user User
	err := repo.GetContextDB(ctx).
		Select(userFieldSql).
		Preload("RoleList").
		Preload("TeamList").
		Where("user_id = ?", userID).
		First(&user).Error
	return user, err
}

func (repo *daoRepo) GetAnonymousUser(ctx core.Context) (User, error) {
	var user User
	err := repo.GetContextDB(ctx).Select(userFieldSql).Where("username = ?", AnonymousUsername).Find(&user).Error
	return user, err
}

func (repo *daoRepo) GetUserList(ctx core.Context, req *request.GetUserListRequest) ([]User, int64, error) {
	var users []User
	var count int64

	query := repo.GetContextDB(ctx).Model(&User{}).Preload("RoleList").Preload("TeamList")

	if len(req.Username) > 0 {
		query = query.Where("username LIKE ?", fmt.Sprintf("%%%s%%", req.Username))
	}

	if len(req.RoleList) > 0 {
		subQuery := repo.GetContextDB(ctx).Table("user_role").
			Select("user_id").
			Where("role_id IN ?", req.RoleList)
		query = query.Where("user_id IN (?)", subQuery)
	}

	if len(req.TeamList) > 0 {
		subQuery := repo.GetContextDB(ctx).Table("user_team").
			Select("user_id").
			Where("team_id IN ?", req.TeamList)
		query = query.Where("user_id IN (?)", subQuery)
	}

	if len(req.Corporation) > 0 {
		query = query.Where("corporation LIKE ?", fmt.Sprintf("%%%s%%", req.Corporation))
	}

	query = query.Where("username != ?", AnonymousUsername)

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

func (repo *daoRepo) RemoveUser(ctx core.Context, userID int64) error {
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

func (repo *daoRepo) UserExists(ctx core.Context, userIDs ...int64) (bool, error) {
	var count int64
	if err := repo.GetContextDB(ctx).Model(&User{}).Where("user_id IN ?", userIDs).Count(&count).Error; err != nil {
		return false, err
	}

	return count == int64(len(userIDs)), nil
}
