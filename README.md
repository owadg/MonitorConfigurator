# MonitorConfigurator
Tool for changing monitor layouts

Some notes: 
I would like to separate the front and the back end. The back end will need to be OS specific, but the front end should be OS agnostic. 
Backend: Go or C++
Frontend: 
Okay, this is a silly thing to differentiate, because the "backend" will be a single function call, which will then change on runtime based on the os. the front end

I decided to make life difficult and use Go. 
I will be using the Fyne library- (https://github.com/fyne-io/fyne) to build a GUI
At least for windows syscalls, I will be using a method like this https://medium.com/@justen.walker/breaking-all-the-rules-using-go-to-call-windows-api-2cbfd8c79724
TODO: Find out how to best change configuration on linux DEs, at least KDE.
TODO: Spin up a Fyne app
TODO: Get the system's display configuration
TODO: graphically display the system's display config (monitors, resolution, refresh rate, relative position)
TODO: graphically edit configs
TODO: Save configs
TODO: Load configs
TODO: Actually change the config on windows
TODOL Change the config on KDE


Requirements:
Frontend: 
          Pull existing system configuration into canvas
          Create a builder for desktop displays configurations.
          Be able to save, load, and edit configurations
          Have it be able to pass, for each monitor the resolution, refresh rate, and position based on selected configuration to the backend
Backend
          Given the resolution, refresh rate, position, os, be able to change the system display configuration.
