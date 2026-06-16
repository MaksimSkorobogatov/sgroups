package app

import (
	"fmt"
	"strings"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	app_identity "github.com/H-BF/corlib/app/identity"
)

// GetVersion returns the application version based on the build information.
func GetVersion() string {
	if version := strings.TrimSpace(app_identity.Version); version != "" {
		if app_identity.BuildTag == "" {
			return fmt.Sprintf("%s|%s", version, app_identity.BuildHash)
		}
		return version
	}

	return misc.Tern(
		app_identity.BuildTag != "", app_identity.BuildTag, app_identity.BuildHash,
	)
}
