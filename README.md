`ffmpeg-tui` is an interactive, simple terminal-based wrapper for FFmpeg built in Go. It simplifies complex video and audio editing operations without hiding what happens under the hood. 

Every action you select instantly constructs the exact, raw FFmpeg command line in real-time, allowing you to learn the flag syntax or tweak the final parameters manually before execution.

## Key Features
* **Live Interactive Command Engine:** Real-time generation and editing of FFmpeg pipeline arguments. Switch focus directly to the terminal input matrix to modify or inject manual parameters.
* **Target Analyzer (FFprobe Integration):** Asynchronously probes incoming media upon startup, displaying resolution, duration, bitrate, file size, frames per second, and codec parameters.
* **Non-Blocking Execution & Progress Streaming:** Runs FFmpeg background tasks using Go contexts while parsing `stderr` time logs into responsive TUI progress indicators.
* **Input Validation Guard:** Parses inputs (`hh:mm:ss` or raw seconds) against actual probed media duration to prevent illegal encoding operations (e.g., out-of-bound trimming or invalid split points).
* **Smart Secondary File Resolution:** Automatically resolves relative or absolute paths for external assets (subtitles, custom audio tracks) relative to the active working directory or source media path.
* **Session History Tracking:** Logs all executed operations, showing real-time success or failure statuses alongside target output file destinations.

### Supported Tools & Operations

| Tool | Description | Configurable Parameters |
| :--- | :--- | :--- |
| **Crop Video** | Aspect ratio transformation or automated crop detection | `9:16` (Shorts/Reels), `1:1` (Square), `16:9` (Widescreen), `Auto Crop Black Bars` (`cropdetect`) |
| **Trim Segment** | Precise lossy/lossless video clipping | Start & End Timestamps (`HH:MM:SS` or seconds) |
| **Split Video** | Chained 2-way split based on a target timestamp | Cutoff Timestamp (`HH:MM:SS` or seconds) |
| **Audio Normalization** | Volume equalization using EBU R128 standard | Automatic `loudnorm` filter integration (`I=-16`, `TP=-1.5`, `LRA=11`) |
| **Frame Export** | Extract a single still frame as PNG | Extraction Timestamp (`HH:MM:SS` or seconds) |
| **Burn Subtitles** | Hardburn `.srt` subtitles with custom styling | File Path, Alignment Position (`bottom`, `top`, `center`), Vertical Offset (px), Background Overlay Color, Text Color |
| **Convert Format** | Re-encode containers or convert to Animated GIF | `MP4`, `MKV`, `MOV`, `AVI`, `MP3` (Pure Audio), `Animated GIF` (with custom palette generation and Lanczos scaling) |
| **CRF Compression** | Balance image quality and file size via x264 | `CRF 23` (Visually Lossless / Balanced), `CRF 28` (High Compression / Social Sharing) |
| **Separate Video/Audio** | Multi-output extraction stream | Strips audio into clean video + extracts high-quality audio (`.mp3`) simultaneously |
| **Strip Metadata** | Privacy cleaner | Strips global metadata tags and stream details (`-map_metadata -1`) |
| **Replace Audio Track** | Audio track override | Input Audio Track Path (`.mp3`, `.wav`, `.m4a`), maps video from stream 0 and audio from stream 1 |


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
# Debian/Ubuntu
sudo apt update && sudo apt install ffmpeg

# Fedora
sudo dnf install ffmpeg

# Arch Linux
sudo pacman -S ffmpeg
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
- Esc or Ctrl + C : Abort the current encoding job or safely exit the application.

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
go build -o ffmpeg-tui  ./cmd/ffmpeg-tui 
```

3. Run the Application
Pass any local video or audio file as an argument to launch the terminal dashboard:
```Bash
./ffmpeg-tui path/to/your/video.mp4
```

## Structural Schema
```
FFmpeg-tui
├── cmd
│   └── ffmpeg-tui
│       └── main.go       # Application Entrypoint: parses arguments and initializes the Bubble Tea loop
├── internal
│   ├── ffmpeg            # Core Module: Handles low-level process wrapper interactions
│   │   ├── command.go    # Edit options data models and command builder logic
│   │   ├── ffprobe.go    # Asynchronous media probing and metadata parser
│   │   └── runner.go     # Process execution, OS channel spawning, and progress listener
│   └── tui               # Interface Module: Houses user interface layouts and event handling
│       ├── messages.go   # Internal Bubble Tea message definitions
│       ├── model.go      # Defines Elm-architecture runtime application states and global struct fields
│       ├── styles.go     # Theme colors, borders, and style variables
│       ├── update.go     # Event multiplexer responding to key strokes, ticks, and state changes
│       └── view.go       # UI layout rendering engine built with Lipgloss
├── go.mod                # Core Go module manifesto tracking external structural dependencies
├── go.sum                # Cryptographic checksums for exact project package locks
└── README.md
```
