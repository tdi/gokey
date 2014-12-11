package main

import (
	"flag"
	"fmt"
	"github.com/tdi/gokey/keystok"
	"os"
	"os/user"
)

var verbose bool
var version string = "0.4"

func print_help() {
	fmt.Printf("gokey version: %s, keystok lib version: %s\n", version, keystok.Version)
	fmt.Println("usage: gokey [-h] [-a ACCESS_TOKEN] [-c CACHE_DIR] [-v] {ls, get} ...")
	os.Exit(0)
}

func print_list_keys(keys map[string]string) {

	if verbose {
		fmt.Printf("KEY ID                         DESCRIPTION\n")
		fmt.Printf("------------------------------ ------------------------------------------\n")
		for k, v := range keys {
			fmt.Printf("%-30s %s\n", k, v)
		}

	} else {
		for k, _ := range keys {
			fmt.Println(k)
		}
	}
}

func main() {

	usr, _ := user.Current()
	dir := usr.HomeDir

	default_cache_dir := fmt.Sprintf("%s/.keystok", dir)

	accessTokenPtr := flag.String("a", "", "access token")
	cacheDirPtr := flag.String("c", "", "cache dir location")
	verbosePtr := flag.Bool("v", false, "verbose")
	useCachePtr := flag.Bool("nc", false, "no cache, default false")
	flag.Parse()

	if len(os.Args) < 2 {
		print_help()
	}

	var access_token string = os.Getenv("KEYSTOK_ACCESS_TOKEN")
	var cache_dir string = os.Getenv("KEYSTOK_CACHE_DIR")

	verbose = *verbosePtr
	if access_token == "" {
		access_token = *accessTokenPtr
	}
	if cache_dir == "" {
		if *cacheDirPtr == "" {
			cache_dir = default_cache_dir
		} else {
			cache_dir = *cacheDirPtr
		}
	}

	if access_token == "" {
		fmt.Println("No access_token at given or KEYSTOK_ACCESS_TOKEN var set")
		print_help()
	}

	var kc keystok.KeystokClient = keystok.GetKeystokClient(access_token)
	kc.Opts.CacheDir = cache_dir
	kc.Opts.UseCache = *useCachePtr

	var command string = ""

	if os.Args[len(os.Args)-1] == "ls" {
		command = "ls"
	} else if os.Args[len(os.Args)-2] == "get" {
		command = "get"
	} else {
		print_help()
	}

	if command == "ls" {
		print_list_keys(kc.ListKeys())
	} else {
		fmt.Println(kc.GetKey(os.Args[len(os.Args)-1]))
	}
	os.Exit(0)
}
