package auth

type Service struct{
	repo Repository
}

func (s *Service) Login(req SigninRequest) (string, error) {
	s.repo.FindByEmail(req.Email)
}