package tests

import (
	"fmt"
	"os/exec"
	"testing"

	"golite.dev/mvp/internal/optimizer"
	"golite.dev/mvp/internal/profiler"
	"golite.dev/mvp/internal/selfevolve"
)

// MockExecutor for self-evolution tests.
type MockEvolveExecutor struct{}

func (m *MockEvolveExecutor) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	// This mock is simplified; a real test might need more complex behavior.
	return nil, nil
}

// MockProfiler for testing the GA logic without running the slow profiler.
type MockProfiler struct{}

func (mp *MockProfiler) Run(sourceFile string, optConfig optimizer.Config) (*profiler.Metrics, error) {
	// Create predictable metrics based on the optimizer config.
	// Let's say ConstantFolding is very good, and DCE is slightly good.
	var runTime float64 = 100.0
	var memUsage int64 = 10000

	if optConfig.IsEnabled(optimizer.ConstantFolding) {
		runTime -= 50.0
		memUsage -= 2000
	}
	if optConfig.IsEnabled(optimizer.DeadCodeElimination) {
		runTime -= 20.0
		memUsage -= 1000
	}

	return &profiler.Metrics{
		RunTimeMs:        runTime,
		MemoryUsageBytes: memUsage,
		BinarySizeBytes:  50000,
	}, nil
}

// We need to create a dummy Profiler runner that uses our mock.
// The actual runner calls the real profiler. We'll create a test-specific runner.
type TestRunner struct {
	mockProfiler *MockProfiler
	ga           *selfevolve.GeneticAlgorithm
}

func (tr *TestRunner) Run(generations int) *selfevolve.Individual {
	pop := tr.ga.CreateInitialPopulation()
	var best *selfevolve.Individual

	for i := 0; i < generations; i++ {
		for _, ind := range pop {
			metrics, _ := tr.mockProfiler.Run("dummy.golite", optimizer.Config{EnabledPasses: ind.Chromosome})
			// Simplified fitness function for testing.
			ind.Fitness = 1.0 / (metrics.RunTimeMs + float64(metrics.MemoryUsageBytes))
		}
		pop.SortByFitness()
		best = pop[0]
		pop = tr.ga.Evolve(pop)
	}
	return best
}

func TestGeneticAlgorithm(t *testing.T) {
	// This test ensures that over a few generations, the GA trends
	// towards the optimal solution given our mock profiler's fitness landscape.
	// The optimal solution is enabling both ConstantFolding and DeadCodeElimination.

	runner := &TestRunner{
		mockProfiler: &MockProfiler{},
		ga:           selfevolve.NewGeneticAlgorithm(),
	}

	// Run for enough generations to likely find the best result.
	bestIndividual := runner.Run(20)

	if bestIndividual == nil {
		t.Fatal("GA did not produce a best individual.")
	}

	fmt.Printf("Test GA found best individual: %v, Fitness: %f\n",
		bestIndividual.PassNames(), bestIndividual.Fitness)

	expectedBestChromosome := optimizer.ConstantFolding | optimizer.DeadCodeElimination
	if bestIndividual.Chromosome != expectedBestChromosome {
		t.Errorf("expected GA to converge on chromosome %v, but got %v",
			expectedBestChromosome, bestIndividual.Chromosome)
	}
}

// LINES: 98
