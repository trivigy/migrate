package migrate

import (
	"encoding/json"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/trivigy/migrate/internal/cmd"
	"github.com/trivigy/migrate/internal/dto"
)

// SetConfigs allows for passing database connection configurations with custom
// environment names.
func SetConfigs(configs map[string]DatabaseConfig) {
	rbytes, err := yaml.Marshal(configs)
	if err != nil {
		panic(err)
	}

	if err := cmd.SetConfigs(rbytes); err != nil {
		panic(err)
	}
}

// Execute runs the main application loop.
func Execute() error {
	if err := cmd.Execute(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Append allows for adding migrations to the migration registry list.
func Append(migration Migration) {
	rbytes, err := json.Marshal(migration)
	if err != nil {
		panic(err)
	}

	dtoMigration := dto.Migration{}
	if err := json.Unmarshal(rbytes, &dtoMigration); err != nil {
		panic(err)
	}

	if err := cmd.Append(dtoMigration); err != nil {
		panic(err)
	}
}
