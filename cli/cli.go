// DistConfig command line tool

package main

import (
	"fmt"
	"os"

	flags "github.com/jessevdk/go-flags"
)

/*

cli --load=Name
cli -l Name
	Load / Update the named item
	If Name already exists in Redis
		Each resource will be checked (Update Timestamp |Hash) for changes - if found then load it.
		Any new resources will be added
	Else
		It will be added
	If modified then a "hash" will be generated for the top level that indicates this

cli --del=Name
cli -d Name
	A "delete" will be perormed on this named resource, delete from Redis also
	The "hash" for this resource will mark it as "deleted"/updated

cli --load=Name --sandbox=SBName
cli -l Name -S SBName
	1. Perform --load as above
	2. Create a sandboxed version with SBName as a copy of resource

cli --watch Name &
cli -w Name &
	Watch the Named files for changes - if changed then perform a --load operation


cli --print Name
cli -p Name
cli --dump Name
cli -d Name
	Print out the complete config for this name (--print), or dump in JSON format
	the resources to files.

cli --get=Name,Item
cli -g Name,Item
cli --get=Name,SbName,Item
cli -g Name,SbName,Item
	Get the specified name/item

cli --activit
cli -a
	Show what would be updated if --load were to be run

cli --check=Name
cli -c Name
	Perform checks on syntax without an actual load

*/
var opts struct {
	GlobalCfgFN string `short:"G" long:"globaCfgFile"  description:"Full path to global config" default:"global-config.json"`
	Load        string `short:"l" long:"load"          description:"Load named item" default:""`
	Sandbox     string `short:"S" long:"sandbox"       description:"Create sandobx for named item" default:""`
	Del         string `short:"d" long:"del"           description:"Delete named item" default:""`
	Watch       string `short:"w" long:"watch"         description:"Watch specified directory for changes and load" default:""`
	Print       string `short:"p" long:"print"         description:"Print in human readable form the nameed item" default:""`
	Dump        string `short:"D" long:"dump"          description:"Dump in JSON the named item" default:""`
	Get         string `short:"g" long:"get"           description:"Get the named item" default:""`
	Activity    string `short:"a" long:"activity"      description:"Show what would be update across named item(s)" default:""`
	Check       string `short:"c" long:"check"         description:"Check named item for syntax correctness" default:""`
}

func main() {

	fns, err := flags.ParseArgs(&opts, os.Args)
	fns = fns[1:] // fns - could be Name List

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Fatal: %s\n", err)
		os.Exit(1)
	}

	ReadGlobalConfigFile(opts.GlobalCfgFN)

	// Connect to Redis

	// Get our library initialized

	act := 0

	if opts.Load != "" {
		act++
	}

	if opts.Sandbox != "" {
		if opts.Load == "" {
			// xyzzy err
		}
	}

	if opts.Del != "" {
		act++
	}

	if opts.Watch != "" {
		act++
	}

	if opts.Print != "" {
		act++
	}

	if opts.Dump != "" {
		act++
	}

	if opts.Get != "" {
		act++
	}

	if opts.Activity != "" {
		act++
	}

	if opts.Check != "" {
		act++
	}

	if act == 0 {
		// xyzzy err
	}

}

// read Redis config and other global config for this program - create global data for that.
func ReadGlobalConfigFile(GlobalCfgFN string) {
}
