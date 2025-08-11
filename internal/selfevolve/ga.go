package selfevolve

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"golite.dev/mvp/internal/optimizer"
)

// Individual represents one set of compiler configurations (a chromosome).
// For now, it only contains the optimizer pass configuration.
type Individual struct {
	Chromosome optimizer.Pass
	Fitness    float64
}

// PassNames returns a slice of strings representing the enabled optimizer passes.
func (i *Individual) PassNames() []string {
	var names []string
	if i.Chromosome&optimizer.ConstantFolding != 0 {
		names = append(names, "ConstantFolding")
	}
	if i.Chromosome&optimizer.DeadCodeElimination != 0 {
		names = append(names, "DeadCodeElimination")
	}
	if len(names) == 0 {
		return []string{"None"}
	}
	return names
}

// Population is a collection of individuals.
type Population []*Individual

// SortByFitness sorts the population in descending order of fitness.
func (p Population) SortByFitness() {
	sort.Slice(p, func(i, j int) bool {
		return p[i].Fitness > p[j].Fitness
	})
}

// GeneticAlgorithm holds the parameters and logic for the evolution process.
type GeneticAlgorithm struct {
	PopulationSize  int
	ElitismCount    int
	MutationRate    float64
	AvailablePasses []optimizer.Pass
}

// NewGeneticAlgorithm creates a GA with default parameters.
func NewGeneticAlgorithm() *GeneticAlgorithm {
	return &GeneticAlgorithm{
		PopulationSize:  20,
		ElitismCount:    2,
		MutationRate:    0.1,
		AvailablePasses: []optimizer.Pass{optimizer.ConstantFolding, optimizer.DeadCodeElimination},
	}
}

// CreateInitialPopulation creates a starting population with random chromosomes.
func (ga *GeneticAlgorithm) CreateInitialPopulation() Population {
	pop := make(Population, ga.PopulationSize)
	for i := range pop {
		var chrom optimizer.Pass
		for _, pass := range ga.AvailablePasses {
			if rand.Float64() < 0.5 {
				chrom |= pass
			}
		}
		pop[i] = &Individual{Chromosome: chrom}
	}
	return pop
}

// Evolve creates a new generation from the current one.
func (ga *GeneticAlgorithm) Evolve(pop Population) Population {
	pop.SortByFitness()
	newPop := make(Population, 0, ga.PopulationSize)

	// Elitism: carry over the best individuals directly.
	for i := 0; i < ga.ElitismCount; i++ {
		newPop = append(newPop, pop[i])
	}

	// Create the rest of the new population through selection and crossover.
	for i := ga.ElitismCount; i < ga.PopulationSize; i++ {
		parent1 := ga.tournamentSelection(pop)
		parent2 := ga.tournamentSelection(pop)
		child := ga.crossover(parent1, parent2)
		newPop = append(newPop, child)
	}

	// Mutate the new population (except for the elite).
	for i := ga.ElitismCount; i < ga.PopulationSize; i++ {
		ga.mutate(newPop[i])
	}

	return newPop
}

// tournamentSelection selects a fit individual from the population.
func (ga *GeneticAlgorithm) tournamentSelection(pop Population) *Individual {
	tournamentSize := 5
	best := pop[rand.Intn(len(pop))]
	for i := 1; i < tournamentSize; i++ {
		next := pop[rand.Intn(len(pop))]
		if next.Fitness > best.Fitness {
			best = next
		}
	}
	return best
}

// crossover combines the chromosomes of two parents to create a child.
func (ga *GeneticAlgorithm) crossover(p1, p2 *Individual) *Individual {
	// Simple single-point crossover on the bitmask.
	// We can't do a simple bitwise crossover, so we'll just pick one.
	// A more complex chromosome would allow for more interesting crossovers.
	childChromosome := p1.Chromosome
	if rand.Float64() < 0.5 {
		childChromosome = p2.Chromosome
	}
	return &Individual{Chromosome: childChromosome}
}

// mutate randomly alters an individual's chromosome.
func (ga *GeneticAlgorithm) mutate(ind *Individual) {
	if rand.Float64() < ga.MutationRate {
		// Flip a random bit corresponding to an available pass.
		passToFlip := ga.AvailablePasses[rand.Intn(len(ga.AvailablePasses))]
		ind.Chromosome ^= passToFlip
	}
}

// For debugging and display
func (p Population) String() string {
	var b strings.Builder
	for i, ind := range p {
		b.WriteString(fmt.Sprintf("  %d: Fitness=%.2f, Passes=%v\n", i, ind.Fitness, ind.PassNames()))
	}
	return b.String()
}

// LINES: 153
