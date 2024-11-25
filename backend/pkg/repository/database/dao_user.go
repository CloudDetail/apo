package database

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"gorm.io/gorm"
)

const (
	RoleAdmin = "admin"
)

type User struct {
	UserID      int64  `gorm:"user_id;primary_key;autoIncrement" json:"-"`
	Username    string `gorm:"username" json:"username,omitempty"`
	Password    string `gorm:"password" json:"password,omitempty"`
	Role        string `gorm:"role" json:"role,omitempty"`
	Phone       string `gorm:"phone" json:"phone,omitempty"`
	Email       string `gorm:"email" json:"email,omitempty"`
	Corporation string `gorm:"corporation" json:"corporation,omitempty"`
}

func (t *User) TableName() string {
	return "user"
}

func Encrypt(raw string) string {
	h := md5.New()
	h.Write([]byte(raw))
	return hex.EncodeToString(h.Sum(nil))
}

func (repo *daoRepo) Login(username, password string) error {
	var user User
	err := repo.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewErrWithMessage(errors.New("user does not exists"), code.UserNotExistsError)
	} else if err != nil {
		return err
	}
	enPassword := Encrypt(password)
	if enPassword != user.Password {
		return model.NewErrWithMessage(errors.New("password incorrect"), code.UserPasswordIncorrectError)
	}

	return nil
}

func (repo *daoRepo) CreateUser(username, password string) error {
	var count int64
	repo.db.Model(&User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return model.NewErrWithMessage(errors.New("user already exists"), code.UserAlreadyExists)
	}
	user := User{
		Username: username,
		Password: Encrypt(password),
	}
	return repo.db.Create(&user).Error
}

func (repo *daoRepo) UpdateUserPhone(username string, phone string) error {
	var user User
	err := repo.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewErrWithMessage(errors.New("user does not exists"), code.UserNotExistsError)
	} else if err != nil {
		return err
	}
	user.Phone = phone
	return repo.db.Save(&user).Error
}

func (repo *daoRepo) UpdateUserEmail(username string, email string) error {
	var user User
	err := repo.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewErrWithMessage(errors.New("user does not exists"), code.UserNotExistsError)
	} else if err != nil {
		return err
	}
	user.Email = email
	return repo.db.Save(&user).Error
}

func (repo *daoRepo) UpdateUserPassword(username, newPassword string) error {
	var user User
	err := repo.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewErrWithMessage(errors.New("user does not exists"), code.UserNotExistsError)
	} else if err != nil {
		return err
	}
	user.Password = Encrypt(newPassword)
	return repo.db.Save(&user).Error
}

func (repo *daoRepo) UpdateUserInfo(req *request.UpdateUserInfoRequest) error {
	var user User
	err := repo.db.Where("username = ?", req.Username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NewErrWithMessage(errors.New("user does not exist "), code.UserNotExistsError)
	} else if err != nil {
		return err
	}

	if len(req.Corporation) > 0 {
		user.Corporation = req.Corporation
	}
	return repo.db.Save(&user).Error
}

func (repo *daoRepo) GetUserInfo(username string) (User, error) {
	var user User
	err := repo.db.Select("username, role, phone, email, corporation").Where("username = ?", username).First(&user).Error
	return user, err
}

func (repo *daoRepo) GetUserList(req *request.GetUserListRequest) ([]User, int64, error) {
	var users []User
	var count int64
	query := repo.db.Select("username, role, phone, email, corporation")
	if len(req.Username) > 0 {
		name := "%" + req.Username + "%"
		query = query.Where("username like ?", name)
	}
	if len(req.Role) > 0 {
		query = query.Where("role = ?", req.Role)
	}
	if len(req.Corporation) > 0 {
		query = query.Where("corporation = ?", req.Corporation)
	}
	err := query.Model(&User{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	query = query.Limit(req.PageSize).Offset((req.CurrentPage - 1) * req.PageSize)
	err = query.Find(&users).Error
	return users, count, err
}

func (repo *daoRepo) RemoveUser(username string, operatorName string) error {
	var operator User
	if err := repo.db.Select("role").Where("username = ?", operatorName).First(&operator).Error; err != nil {
		return err
	}

	if operator.Role != RoleAdmin {
		return model.NewErrWithMessage(errors.New("no permission"), code.UserNoPermissionError)
	}

	return repo.db.Model(&User{}).Where("username = ?", username).Delete(nil).Error
}
