# Design Document: Short Video Merger

## 1. Overview

This document outlines the design for a front-end application that allows users to merge multiple short video files from a selected local directory into a single video file. The core video processing logic will be handled by a WebAssembly (Wasm) module compiled from Go.

## 2. Architecture

The application will consist of two main components:

*   **Frontend**: Built with HTML, CSS, and JavaScript. Responsible for user interaction, file selection, communication with the Wasm module, and displaying progress.
*   **Backend (Wasm Module)**: Written in Go and compiled to Wasm. Responsible for the actual video merging process.

### 2.1 Frontend

#### 2.1.1 UI Components

*   **Directory Selection Button**: An `<input type="file" webkitdirectory directory multiple>` element will be used to allow users to select a directory. A separate button will trigger the click on this hidden input for better styling.
*   **Selected Directory Display**: A `<span>` or `<p>` element to show the path of the selected directory.
*   **Merge Button**: A `<button>` element that becomes visible/enabled after a directory is selected. Clicking this button initiates the video merging process.
*   **Progress Display**: A `<progress>` element or a custom div to show the real-time progress of the video merging operation. This could be a percentage or a textual status update.
*   **Download/Save Area**: After merging, a link or button to allow the user to save the merged video file. The file will be saved in the originally selected directory.

#### 2.1.2 JavaScript Logic

*   **File Handling**:
    *   Listen for changes on the directory selection input.
    *   Read the selected files from the directory.
    *   Sort files based on their names to ensure correct merging order.
*   **Wasm Interaction**:
    *   Load the Wasm module (`main.wasm`).
    *   Prepare video data to be passed to Wasm functions. This might involve reading file contents as byte arrays.
    *   Call exported Go functions in the Wasm module to perform the merge.
    *   Handle data returned from Wasm (the merged video).
    *   Implement callback functions that Go/Wasm can call to update progress on the frontend.
*   **Progress Update**:
    *   Update the progress display based on information received from the Wasm module.
*   **File Saving**:
    *   Once the merged video is received from Wasm, trigger a download or use the File System Access API (if available and appropriate) to save the file into the originally selected directory. The filename could be `merged_video.mp4` (or similar).

### 2.2 Backend (Go/Wasm Module)

#### 2.2.1 Go Video Merging Logic

*   **Core Library**: Utilize a Go library suitable for video manipulation that can be compiled to Wasm. `ffmpeg` is powerful, but directly using its C libraries via Cgo and compiling to Wasm can be complex. A pure Go solution or a Go library that wraps a Wasm-compatible `ffmpeg` build would be ideal. If a simple concatenation of specific formats (e.g., all MP4s with same codecs) is sufficient, a more straightforward approach might be possible. For this initial design, we'll assume a library or approach that can handle common video formats.
    *   *Initial consideration*: If a direct Go library for complex video merging proves difficult with Wasm, the first iteration might focus on a specific, easily concatenable format, or use `ffmpeg` via `exec.Command` if the Go code were running server-side (which is not the case here). For client-side Wasm, the video processing capabilities need to be carefully chosen.
    *   *Alternative for FFmpeg*: One common approach is to use a version of FFmpeg compiled to JavaScript/WebAssembly (e.g., ffmpeg.wasm). The Go Wasm module could then potentially orchestrate calls to this ffmpeg.wasm, or JavaScript could manage it directly, with Go Wasm handling file list and control logic. For this design, we'll aim for Go to manage the merging logic as much as possible.
*   **Exported Functions**:
    *   `mergeVideos(videoData [][]byte, progressCallback js.Func) ([]byte, error)`:
        *   `videoData`: A slice of byte slices, where each inner slice is the content of a video file.
        *   `progressCallback`: A JavaScript function passed from the frontend that Go can call to report progress (e.g., `progressCallback.Invoke(percentage)`).
        *   Returns the merged video as a byte slice and an error if something went wrong.
    *   `init()`: Potentially an initialization function if the Wasm module needs setup.

#### 2.2.2 Wasm Interface (`syscall/js`)

*   Use the `syscall/js` package to define functions callable from JavaScript and to call JavaScript functions (like the progress callback).
*   Manage memory carefully when passing data between JavaScript and Go/Wasm (e.g., `js.CopyBytesToGo`, `js.CopyBytesToJS`).

## 3. Workflow

1.  **User selects a directory**:
    *   Frontend displays the selected directory path.
    *   Merge button becomes active.
2.  **User clicks "Merge"**:
    *   JavaScript reads all files from the selected directory.
    *   Files are filtered (e.g., to include only video files like `.mp4`, `.mov`, `.webm`) and sorted by name.
    *   The content of each video file is read into a byte array.
    *   JavaScript calls the `mergeVideos` function in the loaded Wasm module, passing the video data and a progress callback function.
3.  **Video Merging (Wasm)**:
    *   The Go Wasm module receives the video data.
    *   It processes the videos sequentially (or using a more advanced merging strategy).
    *   Periodically, it calls the JavaScript `progressCallback` function with the current progress percentage.
    *   Once merging is complete, it returns the merged video data (as a byte slice) to JavaScript.
4.  **Displaying/Saving Result**:
    *   JavaScript receives the merged video data.
    *   The progress bar is updated to 100%.
    *   JavaScript creates a Blob from the merged video data and generates a download link, or attempts to save it directly to the user's selected directory if permissions and APIs allow. The filename will be something like `merged_output.mp4`.

## 4. Technology Stack

*   **Frontend**: HTML5, CSS3, JavaScript (ES6+)
*   **Backend Logic**: Go (compiled to WebAssembly)
*   **Video Processing (Go)**: To be determined. Potential options:
    *   A pure Go video manipulation library (if one exists that is suitable and Wasm-compatible).
    *   Using `ffmpeg.wasm` directly from JavaScript, orchestrated by the Go Wasm module or by JavaScript itself.
    *   *Initial simplification*: For a first pass, if general video merging is too complex, one might start with a very specific scenario, like concatenating raw H.264 streams if the input videos are guaranteed to be in that format and compatible. However, the requirement implies general video files.

## 5. Data Flow

```
User Action (Select Directory) -> JS (Read Files) -> User Action (Click Merge) -> JS (Prepare Data)
                                                                                     |
                                                                                     v
                                                                   JS (Call Wasm: mergeVideos)
                                                                                     |
                                                                                     v
                                                                        Go/Wasm (Process Videos) --(Progress)--> JS (Update UI)
                                                                                     |
                                                                                     v
                                                                   JS (Receive Merged Video)
                                                                                     |
                                                                                     v
                                                                   JS (Save/Download File)
```

## 6. Error Handling

*   **Frontend**:
    *   Invalid directory selection.
    *   No video files found in the directory.
    *   Errors during Wasm module loading or execution.
    *   Errors during file saving.
*   **Backend (Go/Wasm)**:
    *   Errors during video decoding/processing/encoding.
    *   Errors due to incompatible video formats or codecs.
    *   These errors should be propagated back to the JavaScript caller.

## 7. Future Considerations / Potential Challenges

*   **Performance**: Merging videos client-side can be resource-intensive. Performance will depend on video size, number of videos, and the efficiency of the Wasm video processing.
*   **Large File Handling**: Browsers have memory limitations. Passing very large video files to/from Wasm needs careful memory management. Streaming or chunking might be necessary for very large files, but this significantly increases complexity.
*   **Video Codec/Format Compatibility**: Ensuring compatibility across various video formats and codecs is a major challenge for video processing. Using a robust library like FFmpeg (via its Wasm port) is often the most practical solution for broad compatibility.
*   **Wasm Compilation of Go Video Libraries**: Many Go multimedia libraries might rely on Cgo and system dependencies (like FFmpeg itself), which can be very challenging to compile to Wasm. This is the biggest technical risk. If `ffmpeg.wasm` is used, Go's role might shift to primarily orchestration and file handling rather than direct video stream manipulation.
*   **User Experience for Long Processes**: For large merges, providing clear, continuous feedback and possibly an option to cancel will be important.
*   **File System Access API**: For a smoother "save to selected directory" experience, the File System Access API could be explored, but it has browser compatibility and permission considerations. A standard download link is a safe fallback.

This design document provides a high-level plan. Specific implementation details, especially around the Go video merging library, will require further research and potential adjustments during development.
