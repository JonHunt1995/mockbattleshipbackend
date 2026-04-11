package main

type badRequest struct {
	status int
	msg    string
}

func (mr *badRequest) Error() string {
	return mr.msg
}
