package linkbuilder

import "fmt"

type LinkBuilder interface {
	ConfirmationLink(token string) string
}

type linkBuilder struct {
	baseURL string
}

func NewLinkBuilder(baseURL string) LinkBuilder {
	return &linkBuilder{baseURL: baseURL}
}

func (l *linkBuilder) ConfirmationLink(token string) string {
	return fmt.Sprintf("%s/api/v1/register/confim/%s", l.baseURL, token)
}
