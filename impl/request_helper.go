package impl

import (
	"github.com/gemfire/cloudcache-management-cf-plugin/domain"
)

//go:generate counterfeiter . RequestHelper

// RequestHelper interface provides a way to get request related items
type RequestHelper interface {
	ExecuteCommand(endpointURL string, httpAction string, commandData *domain.CommandData) (urlResponse string, err error)
	GetTargetAndClusterCommand(args []string) (target string, userCommand domain.UserCommand)
	GetEndPoints(commandData *domain.CommandData) error
}
