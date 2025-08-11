package optimizer

// Pass is a type representing a single optimization pass.
type Pass int

const (
	// ConstantFolding pass replaces constant expressions with their evaluated values.
	ConstantFolding Pass = 1 << iota
	// DeadCodeElimination pass removes code that is unreachable.
	DeadCodeElimination
)

// AllPasses is a convenience constant that enables all available optimization passes.
const AllPasses = ConstantFolding | DeadCodeElimination

// Config holds the configuration for the optimizer, specifying which passes are enabled.
type Config struct {
	EnabledPasses Pass
}

// IsEnabled checks if a specific optimization pass is enabled in the configuration.
func (c *Config) IsEnabled(pass Pass) bool {
	return c.EnabledPasses&pass != 0
}

// LINES: 24
