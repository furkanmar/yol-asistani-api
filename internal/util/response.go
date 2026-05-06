package util

type Response struct {
	Data  any     `json:"data"`
	Error *string `json:"error"`
	Meta  *Meta   `json:"meta,omitempty"`
}

type Meta struct {
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
	Total int `json:"total,omitempty"`
}

func OK(data any) Response {
	return Response{Data: data, Error: nil}
}

func OKWithMeta(data any, meta Meta) Response {
	return Response{Data: data, Error: nil, Meta: &meta}
}

func Err(msg string) Response {
	return Response{Data: nil, Error: &msg}
}
