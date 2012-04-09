package main

import (
	// "log" 
	// "os" 
	// "runtime/pprof" 

	"code.google.com/p/jamslam-x-go-binding/xgb"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/BurntSushi/wingo/logger"
)

// global variables!
var (
	X       *xgbutil.XUtil
	WM      *state
	ROOT    *window
	CONF    *conf
	THEME   *theme
	PROMPTS prompts
)

func main() {
	var err error

	// f, err := os.Create("zzz.prof") 
	// if err != nil { 
	// log.Fatal(err) 
	// } 
	// pprof.StartCPUProfile(f) 
	// defer pprof.StopCPUProfile() 

	X, err = xgbutil.Dial("")
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Println("Error connecting to X, quitting...")
		return
	}
	defer X.Conn().Close()

	// Allow key and mouse bindings to do their thang
	keybind.Initialize(X)
	mousebind.Initialize(X)

	// Create a root window abstraction and load its geometry
	ROOT = newWindow(X.RootWin())
	_, err = ROOT.geometry()
	if err != nil {
		logger.Error.Println("Could not get ROOT window geometry because: %v",
			err)
		logger.Error.Println("Cannot continue. Quitting...")
		return
	}

	// Load configuration
	err = loadConfig()
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Println("No configuration found. Quitting...")
		return
	}

	// Load theme
	err = loadTheme()
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Println("No theme configuration found. Quitting...")
		return
	}

	// Initialize prompts
	promptsInitialize()

	// Create WM state
	WM = newState()
	WM.headsLoad()

	// Set supported atoms
	ewmh.SupportedSet(X, []string{"_NET_WM_ICON"})

	// Attach all global key bindings
	attachAllKeys()

	// Attach all root mouse bindings
	rootMouseConfig()

	// Setup some cursors we use
	setupCursors()

	// Listen to Root. It is all-important.
	ROOT.listen(xgb.EventMaskPropertyChange |
		xgb.EventMaskStructureNotify |
		xgb.EventMaskSubstructureNotify |
		xgb.EventMaskSubstructureRedirect)

	// Update state when the root window changes size
	xevent.ConfigureNotifyFun(rootGeometryChange).Connect(X, ROOT.id)

	// Oblige map request events
	xevent.MapRequestFun(clientMapRequest).Connect(X, ROOT.id)

	// Oblige configure requests from windows we don't manage.
	xevent.ConfigureRequestFun(configureRequest).Connect(X, ROOT.id)

	// Listen to Root client message events.
	// We satisfy EWMH with these AND it also provides a mechanism
	// to issue commands using wingo-cmd.
	xevent.ClientMessageFun(commandHandler).Connect(X, ROOT.id)

	xevent.Main(X)

	// println("Writing memory profile...") 
	// f, err = os.Create("zzz.mprof") 
	// if err != nil { 
	// log.Fatal(err) 
	// } 
	// pprof.WriteHeapProfile(f) 
	// f.Close() 
}
