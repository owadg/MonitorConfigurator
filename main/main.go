package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

	"github.com/owadg/MonitorConfigurator/displaywidgets"

)

func main() {
	//get info from OS

	//construct GUI
	myApp := app.New()
	w := myApp.NewWindow("Box Window Test")
	defer w.ShowAndRun()
	
	setLayout(&w)
}

// our 3 UI elements will be in a Center layout, which displays in a column
func setLayout (myWindow * fyne.Window) {

	
	// elements
	elements := make([]fyne.CanvasObject, 0)
	
	// display layout
	displaywidgets.NewMonitorLabel(1, 1920, 1080, 0, 0)


	// adding it to the window
	content := container.New(layout.NewCenterLayout(), elements...)
	(*myWindow).Canvas().SetContent(content)
}