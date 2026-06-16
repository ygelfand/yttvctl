# yttvctl demo recording
# Install vhs: brew install vhs
# Run: vhs assets/demo.vhs

Output assets/demo.gif

Set FontSize 16
Set Width 1680
Set Height 1050
Set Padding 10
Set WindowBar Rings
Set FontFamily "MesloLGS NF"
Set Shell zsh

# Launch the TUI
Type "./bin/yttvctl tui"
Enter
Sleep 2s

# Scroll the channel grid
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 300ms

# Open the airing detail overlay, then close
Enter
Sleep 800ms
Escape
Sleep 300ms

# Open the device picker, scroll, and pick a Chromecast
Type "d"
Sleep 800ms
Down
Sleep 300ms
Enter
Sleep 1s

# Cast the highlighted channel, then stop
Type "c"
Sleep 2.5s
Type "s"
Sleep 800ms

# Ensure no overlay is open before quitting
Escape
Sleep 200ms
Type "q"
Sleep 500ms
