package authmodule

import "aegis/middleware"

func NewTokenVerifier(service *Service) middleware.TokenVerifier {
	return service
}
