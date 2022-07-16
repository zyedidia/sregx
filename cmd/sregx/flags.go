package main

var opts struct {
	Inplace bool `short:"i" long:"in-place" description:"Change the input file in-place"`
	Version bool `short:"v" long:"version" description:"Show version information"`
	Help    bool `short:"h" long:"help" description:"Show this help message"`
}
