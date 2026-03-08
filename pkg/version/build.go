package version

import (
	"fmt"
	"runtime"
)

var (
	Version string
	GitHash string
	Build   string
)

func Print() {
	fmt.Printf("Version: %s \n", Version)
	fmt.Printf("Build Time: %s \n", Build)
	fmt.Printf("Git Hash: %s \n", GitHash)
	fmt.Printf("Runtime: %s \n", runtime.Version())
}

func Println() {
	//fmt.Printf("%s  %s  %s \n", Version, Build, GitHash)
	println(Version, Build, GitHash)
}
