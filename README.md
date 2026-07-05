`ffmpeg-tui` is an interactive, simple terminal-based wrapper for FFmpeg built in Go. It simplifies complex video and audio editing operations without hiding what happens under the hood. 

Every action you select instantly constructs the exact, raw FFmpeg command line in real-time, allowing you to learn the flag syntax or tweak the final parameters manually before execution.



## Key Features
*   **Live Command Engine:** View the real-time compilation of your FFmpeg command as you toggle filters. Switch focus directly to the command panel to type, edit, or inject custom flags manually.
*   **Media Analysis (FFprobe):** Automatically inspects incoming files upon launch, displaying resolution, duration, codecs, size, and bitrate.
*   **Asynchronous Processing Chain:** Runs FFmpeg tasks in non-blocking background threads, parsing the progress stream in real-time onto a responsive TUI progress bar.
*   **Input Validation Guard:** Automatically parses and evaluates input timestamps (`hh:mm:ss` or raw seconds) against the actual media duration to prevent illegal encoding ranges (e.g., out-of-bounds trimming).
*   **Essential Video Filters & Tools:**
    *   Pre-configured cropping aspects (9:16 Shorts/Reels, 1:1 Square, 16:9 Widescreen).
    *   Precise video trimming with runtime validation.
    *   Fixed 2-way video splitting based on a selected target timestamp.
    *   Single-frame extraction to PNG.
    *   Audio volume normalization (`loudnorm`).
    *   Advanced subtitle hardburning with granular control over text layout placement (bottom, top, center), background overlays, vertical pixel offsets, and custom text colors.
    *   Container format conversion (MP4, MKV, MOV, AVI, and pure MP3 audio extraction) mapping output filenames directly from the original source file.



## The Tech Stack
This project is built using a modern, type-safe terminal stack:
*   **Language:** [Go (1.21+)](https://go.dev/) for high performance, easy cross-compilation, and fast runtime concurrency.
*   **TUI Framework:** The [Charm Ecosystem](https://charm.sh/):
    *   `bubbletea` for Elm-architecture state management.
    *   `lipgloss` for adaptive layout borders, paddings, and the signature *lazygit* color schemes.
    *   `bubbles` for specialized UI components (text input fields and native progress bars).


## Prerequisites
To run and execute jobs with this tool, you must have the **FFmpeg** binaries installed and available in your system's `PATH`.

### Linux
```bash
sudo apt update && sudo apt install ffmpeg
```
### macOS
```Bash
brew install ffmpeg
```
### Windows
```Bash
# One of these
choco install ffmpeg
scoop install ffmpeg
winget install FFmpeg
```
## How to Navigate the TUI
- Tab: Cycle focus sequentially through the different panels (Video Operations -> Parameter Configuration Forms -> Live Command Input Matrix).
- ↑ / ↓ or k / j : Move up and down inside the Video Operations list.
- Enter :
    * Inside the Core Operations Menu: Focus the configuration panel for the selected filter tool.
    * Inside Parameter Configuration Panel: Lock in changes and bounce focus down to the console window.
    * Inside the Live Command Input: Execute the visible FFmpeg pipeline.
- q or Ctrl + C : Abort the current encoding job or safely exit the application.

## Compilation
Ensure you have Go installed on your machine, then clone this repository and follow these steps to compile from source.

1. Initialize and Install Dependencies
```Bash
# Clone the repository and move inside it
cd ffmpeg-tui
# Fetch and sync all Charm ecosystem packages
go mod tidy
```

2. Build the Binary
```Bash
# Compile into a single, standalone binary named 'ffmpeg-tui'
go build -o ffmpeg-tui src/main.go
```

3. Run the Application
Pass any local video or audio file as an argument to launch the terminal dashboard:
```Bash
./ffmpeg-tui path/to/your/video.mp4
```

## Structural Schema
```
ffmpeg-tui/
├── go.mod                 # Core Go module manifesto tracking external structural dependencies
├── go.sum                 # Cryptographic checksums for exact project package locks
│
└── src/
    ├── main.go            # Application Entrypoint: parses arguments and initializes the Bubble Tea loop
    │
    ├── ffmpeg/            # Core Module: Handles low-level process wrapper interactions
    │   ├── command.go     # Declares editing schemas and builds raw FFmpeg flag string arrays
    │   ├── ffprobe.go     # Probes media tracks via FFprobe asynchronous JSON pipes
    │   └── process.go     # Spawns OS execution channels, reads stderr streams, and yields progress tracking
    │
    └── tui/               # Interface Module: Houses user interface layouts and event handling
        ├── model.go       # Defines Elm-architecture runtime application states and global struct fields
        ├── view.go        # Pure view engine building responsive text matrices using Lipgloss blocks
        ├── update.go      # Event multiplexer responding to key strokes, ticks, and state changes
        └── styles.go      # Shared styling variables containing colors, box borders, and theme tokens
```
