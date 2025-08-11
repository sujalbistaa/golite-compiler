package selfevolve

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"golite.dev/mvp/internal/optimizer"
	"golite.dev/mvp/internal/profiler"
)

const checkpointFile = "evolution_checkpoint.json"

// Runner orchestrates the entire self-evolution process.
type Runner struct {
	profiler *profiler.Profiler
	ga       *GeneticAlgorithm
}

// State represents the saved state of the evolution process.
type State struct {
	Generation int        `json:"generation"`
	Population Population `json:"population"`
}

// NewRunner creates a new evolution runner.
func NewRunner(executor profiler.Executor, workDir string) *Runner {
	return &Runner{
		profiler: profiler.New(executor, workDir),
		ga:       NewGeneticAlgorithm(),
	}
}

// Run executes the genetic algorithm for a number of generations over a corpus.
func (r *Runner) Run(corpusDir string, generations int) (*Individual, error) {
	corpusFiles, err := findGoLiteFiles(corpusDir)
	if err != nil || len(corpusFiles) == 0 {
		return nil, fmt.Errorf("corpus directory is empty or could not be read: %w", err)
	}

	pop := r.loadOrInitializePopulation()
	var bestOverall *Individual

	for gen := 1; gen <= generations; gen++ {
		// Evaluate the fitness of each individual in the population.
		for _, individual := range pop {
			individual.Fitness = r.calculateFitness(individual, corpusFiles)
		}

		pop.SortByFitness()
		currentBest := pop[0]
		if bestOverall == nil || currentBest.Fitness > bestOverall.Fitness {
			bestOverall = currentBest
		}

		fmt.Printf("Generation %d/%d | Best Fitness: %.2f | Best Config: %s\n",
			gen, generations, currentBest.Fitness, currentBest.PassNames())

		// Evolve to the next generation.
		pop = r.ga.Evolve(pop)

		// Save checkpoint.
		if gen%5 == 0 {
			if err := r.saveCheckpoint(gen, pop); err != nil {
				fmt.Printf("Warning: could not save checkpoint: %v\n", err)
			}
		}
	}

	return bestOverall, nil
}

// calculateFitness runs the profiler for an individual over the entire corpus
// and returns an aggregated fitness score.
func (r *Runner) calculateFitness(ind *Individual, corpusFiles []string) float64 {
	var totalScore float64
	optConfig := optimizer.Config{EnabledPasses: ind.Chromosome}

	for _, file := range corpusFiles {
		metrics, err := r.profiler.Run(file, optConfig)
		if err != nil {
			// A failing build results in the worst possible fitness.
			fmt.Fprintf(os.Stderr, "Warning: profiling failed for %s with config %v: %v\n", file, ind.PassNames(), err)
			return 0.0
		}
		// Fitness function: lower is better for metrics, so we take the inverse.
		// We add 1 to avoid division by zero. Weights can be tuned.
		// For this example, we prioritize runtime heavily.
		wRuntime := 10.0
		wMemory := 1.0
		wSize := 0.5
		score := 1.0 / (wRuntime*metrics.RunTimeMs + wMemory*float64(metrics.MemoryUsageBytes) + wSize*float64(metrics.BinarySizeBytes) + 1.0)
		totalScore += score
	}

	return totalScore / float64(len(corpusFiles)) // Average score
}

func (r *Runner) loadOrInitializePopulation() Population {
	data, err := ioutil.ReadFile(checkpointFile)
	if err == nil {
		var state State
		if json.Unmarshal(data, &state) == nil {
			fmt.Printf("Resuming from checkpoint at generation %d.\n", state.Generation)
			return state.Population
		}
	}
	fmt.Println("No valid checkpoint found. Starting a new evolution.")
	return r.ga.CreateInitialPopulation()
}

func (r *Runner) saveCheckpoint(generation int, pop Population) error {
	state := State{
		Generation: generation,
		Population: pop,
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("Saving checkpoint to %s\n", checkpointFile)
	return ioutil.WriteFile(checkpointFile, data, 0644)
}

func findGoLiteFiles(rootDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".golite" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// LINES: 161
