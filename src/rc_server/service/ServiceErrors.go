package service

type ServiceError struct {
	Detail string
}

type InternalError ServiceError
type DoesNotExist ServiceError
type AlreadyExists ServiceError

type SensitivityLevel bool

const (
	SENSITIVE     = true
	NOT_SENSITIVE = false
)

type GenericError struct {
	ServiceError
	Code        int
	Sensitivity SensitivityLevel
}

func (g *GenericError) IsSensitive() bool {
	return g.Sensitivity == SENSITIVE
}

func (err *InternalError) Error() string {
	return err.Detail
}

func (err *DoesNotExist) Error() string {
	return err.Detail + " does not exist"
}

func (err *AlreadyExists) Error() string {
	return err.Detail + " already exists"
}

func (err *GenericError) Error() string {
	return err.Detail
}

func NewInternalError(detail string) *InternalError {
	return &InternalError{Detail: detail}
}

func NewDoesNotExistError(detail string) *DoesNotExist {
	return &DoesNotExist{Detail: detail}
}

func NewAlreadyExistsError(detail string) *AlreadyExists {
	return &AlreadyExists{Detail: detail}
}

func NewGenericError(detail string, code int, sensitivity SensitivityLevel) *GenericError {
	return &GenericError{ServiceError: ServiceError{Detail: detail}, Code: code, Sensitivity: sensitivity}
}
