package displaywidgets

import (
	"fmt"
	"math"

	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// this struct will represent the individual monitors in the layout view
// it will display the monitors number, relative resolution size,
// and relative layout with positions. Their stored position will be
// in the digital resolution basis. They will also be draggable, and at the
// end of being dragged, will convert to digital resolution coordinates
type MonitorLabelWidget struct {
	widget.Label					 // will be used to visually display what number monitor is selected
}

type MonitorInformation struct {

}

func NewMonitorLabel(numMonitor int, length int, height int, virtualX int, virtualY int) *MonitorLabelWidget {
	monitorLabel := &MonitorLabelWidget{*widget.NewLabel(fmt.Sprint(numMonitor))}
	monitorLabel.Resize(fyne.NewSize(float32(length), float32(height)))
	monitorLabel.Move(fyne.NewPos(float32(virtualX), float32(virtualY)))
	return monitorLabel
}


func (ml MonitorLabelWidget) updatedPosition (newPosition fyne.Position){
	ml.Move(newPosition)
}

// draggable interface
func (ml MonitorLabelWidget) Dragged(de *fyne.DragEvent) {
	//update the position in the backend
	//TODO - add animation to move thing?
}

// draggable interface
func (ml MonitorLabelWidget) DragEnd() {
	//we done, do nothing
}

// hoverable interface
func (ml MonitorLabelWidget) MouseIn(*desktop.MouseEvent) {
	//let's highlight it when you mouse over it
}

// hoverable interface
func (ml MonitorLabelWidget) MouseMoved(*desktop.MouseEvent) {
}

// hoverable interface
func (ml MonitorLabelWidget) MouseOut(*desktop.MouseEvent) {
	//lets stop highlighting it.
}


//note, when MonitorLabelWidgets are passed to this, they should have their positions, and resolutions with the size filled out.
type MonitorLayout struct {
}

// I am simply going to manually set a size for this. 
func (ml *MonitorLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(400, 400)
}


// @param objects - these objects all need to have a set position and set size. These will correlate to the MonitorLabelWidget's size and relative positioning.
//					from there, we can calculate from those numbers. This layout should not be used for anything other than laying out MonitorLabelWidgets
func (ml *MonitorLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	var smallestXCoord float32
	var smallestYCoord float32
	var largestXCoord float32
	var largestYCoord float32	

	//this snippet gets the smallest x,y coord, as well as rightmost and bottommost virtual coordinates
	for _, o := range objects {
		if o.Position().X < smallestXCoord {
			smallestXCoord = o.Position().X
		}
		if o.Position().Y < smallestYCoord {
			smallestYCoord = o.Position().Y
		}
		if o.Position().X + o.Size().Width > largestXCoord {
			largestXCoord = o.Position().X + o.Size().Width
		}
		if o.Position().Y + o.Size().Height < smallestYCoord {
			largestYCoord = o.Position().Y + o.Size().Height
		}
	}

	// calculate scaling factor - We scale by the larger coordinate span
	ySpan := largestYCoord - smallestYCoord
	xSpan := largestXCoord - smallestXCoord

	span := float32(math.Abs(float64(ySpan - xSpan))) 

	scaling := float32(0.75)
	scalingFactor := containerSize.Width * scaling / span

	// calculates the offset - we want things to display in the center of the widget rather than the top left corner
	xOffset := (containerSize.Width - xSpan * scalingFactor) / 2
	yOffset := (containerSize.Height - ySpan * scalingFactor) / 2

	// we can now apply the scaling factor to the size and positions and add the offset to correctly layout the widgets
	for _, o := range objects {
		currentPosition := o.Position()
		currentSize := o.Size()

		updatedPosition := fyne.NewPos(currentPosition.X * scalingFactor + xOffset, currentPosition.Y * scalingFactor + yOffset)
		updatedSize := fyne.NewSize(currentSize.Width * scalingFactor, currentSize.Height * scalingFactor)

		o.Move(updatedPosition)
		o.Resize(updatedSize)
	}
}
