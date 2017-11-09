package main

import (
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"

	"image"
	"image/color"

	"fmt"
	"math"

	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type Point struct {
	x         float64
	y         float64
	distances []float64
}

func Factorial(n int) int {
	if n == 0 {
		return 1
	}
	return n * Factorial(n-1)
}

func calculateScore(way *[]int, cities *[]Point) float64 {
	score := 0.0
	for idx, k := range *way {
		var j int
		if idx+1 < len(*way) {
			j = (*way)[idx+1]
		} else {
			j = (*way)[0]
		}
		score += (*cities)[k].distances[j]
	}
	return score
}

func calcDistances(cities *[]Point) {
	for i, v := range *cities {
		(*cities)[i].distances = make([]float64, len(*cities))
		for k, v2 := range *cities {
			dx, dy := v.x-v2.x, v.y-v2.y
			(*cities)[i].distances[k] = math.Sqrt(dx*dx + dy*dy)
		}
	}
}

func loadPoints() []Point {
	file, err := os.Open("./data/testa.txt")
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

func permutations(p *[]int, c chan []int) {
	n := len(*p)
	indices := make([]int, n)
	cycles := make([]int, n)
	for i := 0; i < n; i++ {
		indices[i] = i
		cycles[i] = n - i
	}

	c <- *p
	stopLoop := true
	for n > 0 {
		stopLoop = true
		for i := n - 1; i >= 0; i-- {
			cycles[i]--
			if cycles[i] == 0 {
				x := make([]int, 0)
				x = append(x, indices[i+1:]...)
				x = append(x, indices[i:i+1]...)
				backup := indices[:i]
				indices = append(backup, x...)
				cycles[i] = n - i
			} else {
				j := cycles[i]
				indices[i], indices[n-j] = indices[n-j], indices[i]
				c <- indices
				stopLoop = false
				break
			}
		}
		if stopLoop {
			break
		}
	}
	close(c)
}

func main() {
	cities := loadPoints()
	calcDistances(&cities)
	points := make([]int, 0)

	for i := range cities {
		points = append(points, i)
	}

	fmt.Printf("Nro de ciudades: %d \n", len(cities))
	fmt.Printf("Factorial: %d \n", Factorial(len(cities)))

	ch := make(chan []int)
	go permutations(&points, ch)
	bestScore := -1.0
	bestWay := make([]int, len(cities))
	score := -1.0
	idx := 0
	start := time.Now()
	for i := range ch {

		if idx%5000000 == 0 && idx != 0 {
			fmt.Printf("%d Iteracion\n", idx)
		}
		idx++

		score = calculateScore(&i, &cities)

		if bestScore < 0 || bestScore > score {
			fmt.Printf("Nueva ruta candidata encontrada: %f %v \n", score, i)
			bestScore = score
			copy(bestWay, i)
		}

	}
	elapsed := time.Since(start)
	fmt.Printf("Esta es la mejor ruta! resultado: %f %v \n", bestScore, bestWay)
	fmt.Printf("Tiempo de ejecución: %s", elapsed)

	drawWay(bestWay, cities)
}
