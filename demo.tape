# VHS Tape for Docklog Demo
# Run with: vhs demo.tape

Output demo.gif

# Setup Terminal
Set FontSize 16
Set Width 1200
Set Height 700
Set Padding 40
Set WindowBar Fixed
Set Theme "Catppuccin Macchiato"
Set TypingSpeed 75ms

# --- Scene 1: Basic Start ---
Type "docklog start --container 'broken-*' --tail 0"
Sleep 500ms
Enter
Sleep 4s
Ctrl+C
Sleep 1s

# --- Scene 2: Smell Error (The Hero Feature) ---
# This shows how we filter noise
Type "docklog smell-error --container 'broken-*'"
Sleep 500ms
Enter
Sleep 8s
Ctrl+C
Sleep 1s

# --- Scene 3: Redaction ---
Type "docklog start --redact --container 'natural-*'"
Sleep 500ms
Enter
Sleep 5s
Ctrl+C

# Finish
Sleep 1s
Type "clear"
Enter
Sleep 1s
Type "# Docklog: Zero Noise Docker Logging"
Sleep 2s
Enter
