package displaywidgets

import (
	"fyne.io/fyne/v2"
)

// the goal of this struct is to provide a useful way to get the selected information 
type displayInformation struct{
	primary 					bool
	selectedResolutionWidth		int
	selectedResolutionHeight 	int
	selectedOrientation 		int // can be values of 0, 1, 2, 3
	selectedRefreshRate 		int
	selectedPosition 			int
}

