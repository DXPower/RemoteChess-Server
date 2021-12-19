package service

import "fmt"

type SensitivityLevel bool

const (
	SENSITIVE     = true
	NOT_SENSITIVE = false
)

type ServiceError struct {
	Detail          string
	HttpCodeHint    int
	SensitivityHint SensitivityLevel
	format          string
}

func (s *ServiceError) IsSensitive() bool {
	return s.SensitivityHint == SENSITIVE
}

func (s *ServiceError) Error() string {
	return fmt.Sprintf(s.format, s.Detail)
}

func NewInternalError(detail string) *ServiceError {
	return &ServiceError{detail, 500, SENSITIVE, "%s"}
}

func NewDoesNotExistError(detail string) *ServiceError {
	return &ServiceError{detail, 404, NOT_SENSITIVE, "%s does not exist"}
}

func NewAlreadyExistsError(detail string) *ServiceError {
	return &ServiceError{detail, 409, NOT_SENSITIVE, "%s already exists"}
}

func NewInvalidInputError(detail string) *ServiceError {
	return &ServiceError{detail, 400, NOT_SENSITIVE, "%s is malformed"}
}

func NewGenericError(detail string, httpCodeHint int, sensitivityHint SensitivityLevel) *ServiceError {
	return &ServiceError{detail, httpCodeHint, sensitivityHint, "%s"}
}
