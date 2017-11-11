package main

import (
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"

	"image"
	"image/color"

	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Point struct {
	x         float64
	y         float64
	distances []float64
}

type Individual struct {
	way   []int
	score float64
}
type IntSlice []int
type ByScore []Individual

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].score < a[j].score }

/* calculateScore calculates the score for a given way*/
func (i *Individual) calculateScore(cities *[]Point) {
	i.score = 0
	for idx, k := range i.way {
		var j int
		if idx+1 < len(i.way) {
			j = i.way[idx+1]
		} else {
			j = i.way[0]
		}
		i.score += (*cities)[k].distances[j]
	}
}

func (i *Individual) createRandomWay(cities *[]Point) {
	rand.Seed(time.Now().UTC().UnixNano())
	i.way = rand.Perm(len(*cities))
	i.calculateScore(cities)
}

func (i *Individual) mutate() {
	rand.Seed(time.Now().UTC().UnixNano())
	c1, c2 := rand.Intn(len(i.way)-1), rand.Intn(len(i.way)-1)
	i.way[c1], i.way[c2] = i.way[c2], i.way[c1]
}

func (slice IntSlice) isIn(value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func (i Individual) crossover(partner *Individual) []Individual {
	pivot1 := rand.Intn(len(i.way) - 1)
	rand.Seed(time.Now().UTC().UnixNano())
	pivot2 := rand.Intn(len(i.way)-pivot1) + pivot1
	i1 := Individual{way: nil, score: 0}
	i2 := Individual{way: nil, score: 0}

	// for the first child
	for _, v := range i.way {
		if !IntSlice(partner.way[pivot1 : pivot2+1]).isIn(v) {
			i1.way = append(i1.way, v)
		}
	}

	// dragons ahead!
	x := make([]int, 0)
	x = append(x, i1.way[:pivot1]...)
	x = append(x, partner.way[pivot1:pivot2+1]...)
	x = append(x, i1.way[pivot1:]...)
	i1.way = x

	// and for the second one
	for _, v := range partner.way {
		if !IntSlice(i.way[pivot1 : pivot2+1]).isIn(v) {
			i2.way = append(i2.way, v)
		}
	}
	// dragons ahead pt2!
	x = make([]int, 0)
	x = append(x, i2.way[:pivot1]...)
	x = append(x, i.way[pivot1:pivot2+1]...)
	x = append(x, i2.way[pivot1:]...)
	i2.way = x

	return []Individual{i1, i2}
}

type Env struct {
	crossoverChance     float64
	mutationChance      float64
	newIndividualFactor float64
	chooseBestChange    float64
	breakAfter          int

	maxGenerations    int
	populationSize    int
	currentGeneration int
	populationScore   float64

	cities     []Point
	population []Individual
}

func (e *Env) initialize() {
	e.calcDistances()
	e.population = make([]Individual, e.populationSize)
	e.createRandomPopulation()
}

func (e *Env) createRandomPopulation() {
	for i := 0; i < e.populationSize; i++ {
		e.population[i] = Individual{way: nil, score: 0}
		e.population[i].createRandomWay(&(e.cities))
	}
	e.calcScore()
}

func (e *Env) calcDistances() {
	for i, v := range e.cities {
		e.cities[i].distances = make([]float64, len(e.cities))
		for k, v2 := range e.cities {
			dx, dy := v.x-v2.x, v.y-v2.y
			e.cities[i].distances[k] = math.Sqrt(dx*dx + dy*dy)
		}
	}
}

/* calcScore calculates the score for all the population
and sorts the population by score
*/
func (e *Env) calcScore() {
	e.populationScore = 0
	for i := range e.population {
		e.population[i].calculateScore(&e.cities)
		e.populationScore += e.population[i].score
	}
	sort.Sort(ByScore(e.population))
}

func (e *Env) doCrossover() {
	rand.Seed(time.Now().UTC().UnixNano())
	if rand.Float64() < e.crossoverChance {
		crossoverCount := int(e.newIndividualFactor * float64(e.populationSize))
		children := make([]Individual, 0)
		//fmt.Printf("Adding %d new children\n", crossover_count * 2)

		// generates children
		for i := 0; i < crossoverCount; i++ {
			var p1 int
			var p2 int
			min := int(e.populationSize / 3)
			// choose parents for crossover
			if rand.Float64() < e.chooseBestChange {
				p1, p2 = rand.Intn(min), rand.Intn(min)
			} else {
				p1, p2 = rand.Intn(e.populationSize-min)+min, rand.Intn(e.populationSize-min)+min
			}
			children = append(children, e.population[p1].crossover(&e.population[p2])...)
		}
		for i := range children {
			children[i].calculateScore(&e.cities)
		}
		// selects some children for the population
		for _, v := range children {
			addMeScore := 0.5
			randomScore := rand.Float64()
			for j := range e.population {
				minusJPosition := len(e.population) - 1 - j
				minusJScore := e.population[minusJPosition].score
				addMeScore += (minusJScore / e.populationScore)
				if addMeScore < 1-randomScore {
					e.population[minusJPosition] = v
					break
				}
			}
			e.calcScore()
		}
		fmt.Printf("Best score: %f\n", e.population[0].score)
	}
}

// mutates a random element in a random route
func (e *Env) doMutation() {
	rand.Seed(time.Now().UTC().UnixNano())
	if rand.Float64() < e.mutationChance {

		idx := rand.Intn(e.populationSize / 3)
		p1, p2 := rand.Intn(len(e.population[idx].way)), rand.Intn(len(e.population[idx].way))
		e.population[idx].way[p1], e.population[idx].way[p2] = e.population[idx].way[p2], e.population[idx].way[p1]
		e.calcScore()
		//fmt.Printf("Mutation done\n")
	}
}

func (e *Env) run() {
	currentScore := e.populationScore
	noChanges := 0
	for i := 0; i < e.maxGenerations; i++ {
		e.currentGeneration++
		fmt.Printf("Current generation: %d\n", e.currentGeneration)
		e.doCrossover()
		e.doMutation()

		if currentScore == e.populationScore {
			noChanges++
		} else {
			noChanges = 0
		}

		if noChanges == e.breakAfter {
			fmt.Printf("Stuck in local maximum for %d generations\n", e.breakAfter)
			break
		}
		currentScore = e.populationScore
	}
	drawWay(e.population[0].way, e.cities)
}

func loadPoints() []Point {
	file, err := os.Open("./data/testb.txt")
	if err != nil {
		fmt.Fprint(os.Stderr, "Cannot open points file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	points := make([]Point, 0)
	for scanner.Scan() {
		pData := strings.Split(scanner.Text(), " ")
		x, _ := strconv.ParseFloat(pData[1], 64)
		y, _ := strconv.ParseFloat(pData[2], 64)
		points = append(points, Point{
			x,
			y,
			nil,
		})
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprint(os.Stderr, "Points file reading error!")
	}
	return points
}

func drawWay(way []int, points []Point) {

	biggestX, biggestY := points[0].x, points[0].y
	for _, p := range points[1:] {
		if biggestX < p.x {
			biggestX = p.x
		}
		if biggestY < p.y {
			biggestY = p.y
		}
	}
	padding := 40.0
	resizeBy := 10.0

	dest := image.NewRGBA(image.Rect(0, 0, int(biggestX*resizeBy+padding), int(biggestY*resizeBy+2*padding)))
	gc := draw2dimg.NewGraphicContext(dest)

	gc.SetFillColor(color.RGBA{0x44, 0x44, 0x44, 0})
	gc.SetStrokeColor(color.RGBA{0x44, 0x44, 0x44, 0xff})
	gc.SetLineWidth(1)

	draw2dkit.Circle(gc, padding+points[way[0]].x*resizeBy, padding+points[way[0]].y*resizeBy, 3)
	gc.MoveTo(padding+points[way[0]].x*resizeBy, padding+points[way[0]].y*resizeBy)
	for _, v := range way[1:] {
		gc.LineTo(padding+points[v].x*resizeBy, padding+points[v].y*resizeBy)
		gc.MoveTo(padding+points[v].x*resizeBy, padding+points[v].y*resizeBy)
		gc.Close()
		draw2dkit.Circle(gc, padding+points[v].x*resizeBy, padding+points[v].y*resizeBy, 3)
	}
	gc.LineTo(padding+points[way[0]].x*resizeBy, padding+points[way[0]].y*resizeBy)
	gc.FillStroke()
	draw2dimg.SaveToPngFile("way.png", dest)
}

func main() {
	start := time.Now()
	points := loadPoints()

	env := Env{
		maxGenerations:      9000,
		crossoverChance:     0.95,
		mutationChance:      0.4,
		chooseBestChange:    0.95,
		currentGeneration:   0,
		newIndividualFactor: 0.2,
		populationSize:      100,
		breakAfter:          100,
		population:          nil,
		cities:              points,
	}

	env.initialize()
	env.run()
	elapsed := time.Since(start)
	fmt.Printf("Tiempo de ejecución: %s", elapsed)
	fmt.Printf("%f\n", env.population[0].score)
	fmt.Printf("%d\n", env.maxGenerations)
	fmt.Printf("%f\n", env.crossoverChance)
	fmt.Printf("%f\n", env.mutationChance)
	fmt.Printf("%f\n", env.chooseBestChange)
	fmt.Printf("%f\n", env.newIndividualFactor)

}
