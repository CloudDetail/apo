package user

func (s *service) RemoveUser(username string, operatorName string) error {
	return s.dbRepo.RemoveUser(username, operatorName)
}
