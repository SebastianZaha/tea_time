Crossplatform minimal systray / menubar timer.


To bind to universal keyboard shortcut on OSX:
    Automator -> New Quick Action -> Add "Run Shell Script" -> 
    .sh with absolute path to executable (dependent on go install / go get path)
    System Settings -> Keyboard Shortcuts -> Services -> General -> bind key.

Windows (To avoid opening a console at application startup)
     go install -ldflags -H=windowsgui