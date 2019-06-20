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
func SetConfigs(configs map[string]DataSource) {
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

// ExecuteWithArgs runs the main application loop with provided arguments.
func ExecuteWithArgs(args ...string) (string, error) {
	output, err := cmd.ExecuteWithArgs(args...)
	if err != nil {
		return output, errors.WithStack(err)
	}
	return output, nil
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

// Clear restarts the application with all the command flags, configurations,
// and migration registory reset. This is primarily useful for testing.
func Clear() []Migration {
	dtoMigrations := cmd.Clear()

	rbytes, err := json.Marshal(dtoMigrations)
	if err != nil {
		panic(err)
	}

	migrations := make([]Migration, 0)
	if err := json.Unmarshal(rbytes, &migrations); err != nil {
		panic(err)
	}

	return migrations
}

// Close terminates the database connection.
func Close() error {
	return cmd.Close()
}
