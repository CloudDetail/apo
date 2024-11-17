package user

func (s *service) RemoveUser(username string) error {
	return s.dbRepo.RemoveUser(username)
}
