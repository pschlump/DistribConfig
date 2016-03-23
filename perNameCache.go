package DistConfig

import (
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/pschlump/Go-FTL/server/bufferhtml"
	"github.com/pschlump/Go-FTL/server/lib"
	"github.com/pschlump/godebug" //
)

//	"github.com/pschlump/Go-FTL/server/bufferhtml"
//	"github.com/pschlump/Go-FTL/server/lib"
//	"github.com/pschlump/godebug" //

// "github.com/pschlump/Go-FTL/server/bufferhtml" //

/*
Use Case:
	Suppose that you have 100 people that all use GitHub and each config has an "on-change" hook.
	The files get verified by the user locally then pushed up - then - the hook runs.
	It runs the "cli" to check the files and see which ones have changed - and loads the data into Redis.
	Then each client when it next checks the data sees that the hash has changed on a set of data - and reloads that set.

Notes:
	1. Should be part of ../../cfg -- or its own config stuff --
	2. If you chagne any sub-config then this must be reflectd in a chagne to the hash of the "config" top level
	3. If the top level hash has chagned then walk all sub-leves and update
	4. Must implement CLI to set all of this up
		0. Place CLI in ../../../tools/config directory
		1. Must check date-time stamps for each file and load any that need to be loaded
		2. "--defualt--" comes from the ./cfg--default-- directory
		3. "working_test_for_aes_srp" is an example of a "server name" - server names must be [a-zA-Z][a-zA-Z_0-9]* - validate this.
		4. Load any subdirectoris that match with server names "./cfg-[NameOfServer]"
	+5. Need to deal with "sandbox"
		+1. A "setup of sandbox creates time-stamped (will expire data) as a copy of a named server
		+2. All mods to data will result in updating all time stamps on data for Delta-T into future
	6. Change Notification on data

	*use a "registerÇonvFx" process that adds ":itemName" and Conversion Function for stuff -- Add in "file name"
*/

// ----------------------------------------------------------------------------------------------------------------------------------------------

// This set of type/function allows for the registration of new data types to be added to the set.
// You can create a type, then in an func init()  - call the RegisterConvItem(":name","file.json",ConvFunction) and
// instantiate the data to be shared/updated.

type ConvTableType struct {
	NameSetName string              // the :name like :bits or :security
	FileName    string              // the initiazation file that this kind of data will come from, "plugins.json", or "extProcessTable.json" for example
	ConvFx      PerConvFunction     // convert from Data -> Dataparsed -- for this type of data
	InitFx      PerConvInitFunction // get initial data - empty
}

var registeredNames []ConvTableType

func RegisterConvItem(n, fn string, fx PerConvFunction) {
	registeredNames = append(registeredNames, ConvTableType{NameSetName: n, FileName: fn, ConvFx: fx})
}

// ----------------------------------------------------------------------------------------------------------------------------------------------

// Change Notification Functions - Per Named User, Per Item
// Events are Pre-Delete, Post-Delete, Change-Of-Data + name changed, 1st-Time-Init, Shutdown-Now

// ----------------------------------------------------------------------------------------------------------------------------------------------

type PerConvFunction func(data string) (parsedData interface{}, err error)       // convert from Data -> Dataparsed
type PerConvInitFunction func() (data string, parsedData interface{}, err error) // get inital default data

// ----------------------------------------------------------------------------------------------------------------------------------------------

type PerNameSomeData struct {
	NameSetName   string          // :name of self
	HashOfData    string          // sha256 hash of raw Data
	ModTimeOfData time.Time       // When was file modified that this came from --- xyzzy - change type
	FullKey       string          // Concatenated key		== KeyPrefixInRedis + KeySandbox + KeyPostfix
	KeyPostfix    string          // Chunk to add to key
	KeySandbox    string          // Chunk to add to key - quite often empty string
	Data          string          // Cached data raw
	DataParsed    interface{}     // Data converted to its final data type
	ConvFx        PerConvFunction // convert from Data -> Dataparsed
	ItemNo        int
}

type PerNameType struct {
	KeyPrefixInRedis string                      // Normally "srp:U:"		-- srp:U::{server}:Hash - is the item to check to see if change occured
	HashOfData       string                      // sha256 of hashes of data in set
	NamedSet         map[string]*PerNameSomeData // What is this data called, ":security", ":bits", ":plugins", ":ExtProcessTable", ":CommandLocaitonMap" etc.
}

type PerNameCacheType struct {
	lockIt       *sync.Mutex             // used when changing data
	byName       map[string]*PerNameType // indexed by server name - config like security is per-named server
	errors       []string                // Any parse errors
	validSandbox map[string]bool         // Created Sandboxes -- May also need expiration date/time for them
}

func NewPerNameCacheType() (pnc *PerNameCacheType) {
	pnc = &PerNameCacheType{
		lockIt: &sync.Mutex{},
		byName: make(map[string]*PerNameType),
	}
	h := "" // build later
	pnc.byName["--default--"] = &PerNameType{
		KeyPrefixInRedis: "srp:U:",
		HashOfData:       h,
		NamedSet:         make(map[string]*PerNameSomeData), // What is this data called, ":security", ":bits", ":plugins", ":ExtProcessTable", ":CommandLocaitonMap" etc.
	}

	// pnc.byName["--default--"].NamedSet[":bits"] = &PerNameSomeData{}               // xyzzy - getDataFor ( "--default--", ":bits" )
	// pnc.byName["--default--"].NamedSet[":security"] = &PerNameSomeData{}           //
	// pnc.byName["--default--"].NamedSet[":plugins"] = &PerNameSomeData{}            //
	// pnc.byName["--default--"].NamedSet[":ExtProcessTable"] = &PerNameSomeData{}    //
	// pnc.byName["--default--"].NamedSet[":CommandLocationMap"] = &PerNameSomeData{} //
	// Add :trace, ":debug", ":logConfig"

	sh := ""
	sc := ""
	for ii, vv := range registeredNames {
		d, p, err := vv.InitFx()
		if err == nil {
			h := lib.Sha256(d)
			sh += sc + h // this may incorporate some sort of subtle error relating to link order - may need to sort data by :name ??
			sc = "::"
			ts := time.Now()                         // Assume timestamp is right now for any data that we do not read in.
			xfn := "./cfg--default--/" + vv.FileName // XyzzyPath555111 - may want to add "PATH" to this.
			if exists, fi := lib.ExistsGetFileInfo(xfn); exists {
				sb, err := ioutil.ReadFile(xfn)
				if err != nil {
				}
				d = string(sb)
				p, err = vv.ConvFx(d)
				if err != nil {
					bufferhtml.G_Log.Error(fmt.Sprintf("Initialization: Error: %s, Name: %s, %s\n", err, vv.NameSetName, godebug.LF()))
					goto next
				}
				ts = fi.ModTime() // get time stamp from ./cfg-default/[vv.FileName] -- if file exists
			}
			pnc.byName["--default--"].NamedSet[vv.NameSetName] = &PerNameSomeData{
				NameSetName:   vv.NameSetName, //
				HashOfData:    h,              // sha256 hash of raw Data
				ModTimeOfData: ts,             // When was file modified that this came from --- xyzzy - change type
				FullKey:       "",             //string          // Concatenated key		== KeyPrefixInRedis + KeySandbox + KeyPostfix	-- Generated on the fly
				KeyPostfix:    vv.NameSetName, //string          // Chunk to add to key
				KeySandbox:    "",             //string          // Chunk to add to key - quite often empty string	-- when creating sandbox this can be set
				Data:          d,              //string          // Cached data raw
				DataParsed:    p,              //interface{}     // Data converted to its final data type
				ConvFx:        vv.ConvFx,
				ItemNo:        ii,
			}

		next:
		} else {
			bufferhtml.G_Log.Error(fmt.Sprintf("Initialization: Error: %s, Name: %s, %s\n", err, vv.NameSetName, godebug.LF()))
		}
	}

	h = Sha256(sh)
	pnc.byName["--default--"].HashOfData = h

	return
}

// Lookup --default-- server and clone it. -- then look for files and pull them in.
func (pnc *PerNameCacheType) AddNewName(name string) {
	// Send Pre - 1st init on item message
	// Lock
	// Read in item
	// UnLock
	// Send  event
}

func (pnc *PerNameCacheType) DeleteName(name string) {
	// Send Deleteted Item Pre Notification
	// Lock
	delete(pnc.byName, name) // Delete item
	// UnLock
	// Send Deleteted Item Post Notification
}

// Hm... still working on this - params may be wrong -- what about sandbox - what are we updating.
// Why not just chagne file - and -- reload entire name -- maybee that is what this is.
func (pnc *PerNameCacheType) UpdateName(name string) {
}

// Creates a sandbox for the named server - copy all config to sandbox - Set expiration dates
func (pnc *PerNameCacheType) CreateSandbox(name, sb string) {
}

// Sample Call:  pnc.GetConfigFor ( "working_test_for_aes_srp", "", ":bits" )
func (pnc *PerNameCacheType) GetConfigFor(name string, sb, item string) (it interface{}) {
	// if have it in pnc, by server name
	// if have item by item name -> NamedSet
	// check to see if need to pull new version from Redis - check Hash
	return
}
