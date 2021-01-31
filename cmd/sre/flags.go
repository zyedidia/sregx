package main

var opts struct {
	File    string `short:"f" long:"file" description:"Read input data from file (default: read from stdin)"`
	Version bool   `short:"v" long:"version" description:"Show version information"`
	Help    bool   `short:"h" long:"help" description:"Show this help message"`
}
