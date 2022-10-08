package resize

import "interview-fm-backend/internal/utils"

type Service struct {
}

func NewResizerService() *Service {
	return &Service{}
}

func (s *Service) ResizeImage(data []byte, width uint, height uint) ([]byte, error) {
	return utils.ResizeImage(data, width, height)
}
