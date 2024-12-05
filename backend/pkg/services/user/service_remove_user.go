package user

func (s *service) RemoveUser(userID int64, operatorID int64) error {
	return s.dbRepo.RemoveUser(userID, operatorID)
}
