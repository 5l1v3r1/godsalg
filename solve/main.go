package main

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/unixpickle/anydiff"
	"github.com/unixpickle/anynet"
	_ "github.com/unixpickle/anyplugin"
	"github.com/unixpickle/anyvec"
	"github.com/unixpickle/anyvec/anyvec32"
	"github.com/unixpickle/gocube"
	"github.com/unixpickle/godsalg"
	"github.com/unixpickle/serializer"
)

const (
	BatchSize = 1000
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: solve <network>")
		os.Exit(1)
	}
	var net anynet.Net
	if err := serializer.LoadAny(os.Args[1], &net); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load:", err)
		os.Exit(1)
	}

	cube, err := gocube.InputStickerCube()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Bad input:", err)
		os.Exit(1)
	}
	state, err := cube.CubieCube()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Bad state:", err)
		os.Exit(1)
	}

	for i := 0; true; i++ {
		solution := sampleSolution(*state, net)
		if solution != nil {
			fmt.Println("Solution:", solution)
			break
		} else {
			fmt.Println("Attempt", i, "failed")
		}
	}
}

func sampleSolution(start gocube.CubieCube, net anynet.Net) []gocube.Move {
	solutions := make([][]gocube.Move, BatchSize)
	states := make([]*gocube.CubieCube, BatchSize)
	for i := range states {
		c := start
		states[i] = &c
	}
	for i := 0; i < 21; i++ {
		var inVec []float64
		for j, x := range states {
			if x.Solved() {
				return solutions[j]
			}
			inVec = append(inVec, godsalg.CubeVector(x)...)
		}
		inRes := anydiff.NewConst(
			anyvec32.MakeVectorData(anyvec32.MakeNumericList(inVec)),
		)
		output := net.Apply(inRes, BatchSize).Output()
		anyvec.Exp(output)
		slice := output.Data().([]float32)
		for j := 0; j < BatchSize; j++ {
			part := slice[j*godsalg.NumMoves : (j+1)*godsalg.NumMoves]
			move := selectMoveVector(part)
			solutions[j] = append(solutions[j], move)
			states[j].Move(move)
		}
	}
	return nil
}

func selectMoveVector(vec []float32) gocube.Move {
	p := rand.Float32()
	for i, x := range vec {
		p -= x
		if p < 0 {
			return gocube.Move(i)
		}
	}
	return 0
}
