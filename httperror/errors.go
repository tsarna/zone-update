package httperror

type HttpError interface {
  error
  HttpStatus() int
}

type errorWithStatus struct {
  err        error
  httpStatus int
}

func (err errorWithStatus) Error() string {
  return err.err.Error()
}

func (err errorWithStatus) HttpStatus() int {
  return err.httpStatus
}

func Error(httpStatus int, err error) HttpError {
  return errorWithStatus { httpStatus: httpStatus, err: err }
}
