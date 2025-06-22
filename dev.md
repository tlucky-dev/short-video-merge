# Development Plan: Short Video Merger

This document outlines the development steps for the Short Video Merger project, based on the requirements in `README.md` and the architecture in `design.md`.

## Phase 1: Backend - Go/Wasm Video Merging Logic

**Goal**: Create a Wasm module in Go that can merge video files. The initial focus will be on concatenating compatible MP4 files as a proof-of-concept, as general-purpose video merging is complex. If `ffmpeg.wasm` is chosen, this phase will involve setting up Go to interact with it.

1.  **Setup Go Wasm Environment**:
    *   Ensure Go is installed.
    *   Familiarize with `syscall/js` for JS-Go interop.
    *   Set up a simple "hello world" Go Wasm program and run it in a browser.
2.  **Research & Select Video Merging Strategy**:
    *   **Option A (Proof of Concept - Simple Concatenation)**:
        *   Research libraries or methods for concatenating MP4 files in pure Go (if possible and Wasm-compatible). This might involve parsing MP4 container formats and appending streams. This is highly complex for arbitrary MP4s but might work for identically encoded segments.
        *   If pure Go is too complex, investigate using `ffmpeg.wasm`.
    *   **Option B (Using `ffmpeg.wasm`)**:
        *   Learn how to use `ffmpeg.wasm` from JavaScript.
        *   Design how the Go Wasm module will orchestrate `ffmpeg.wasm` (e.g., Go receives file data, tells JS to run ffmpeg commands, gets result back). Or, JS handles ffmpeg directly, and Go Wasm is only for other logic if any. Given the requirements, Go should ideally drive the merging.
    *   **Decision**: For initial development, we will attempt to use `ffmpeg.wasm` due to its robustness in handling various video formats. The Go Wasm module will be responsible for receiving file information from the JavaScript frontend, preparing commands for `ffmpeg.wasm`, invoking `ffmpeg.wasm` (which itself runs in the browser's JS environment), and handling the output.
3.  **Implement Basic Go Wasm Video Merging Function**:
    *   Define a Go function exportable to Wasm, e.g., `mergeVideosGo(fileNames []string, fileContents [][]byte) ([]byte, error)`.
        *   `fileNames`: To help `ffmpeg.wasm` identify input files (it works with virtual file systems).
        *   `fileContents`: Byte arrays of the video files.
    *   Inside this function:
        *   Use `syscall/js` to interact with `ffmpeg.wasm` (which will be loaded in the main JS context).
        *   Create virtual files for `ffmpeg.wasm` from `fileContents`.
        *   Construct and execute the `ffmpeg` command for concatenation (e.g., using the `concat` demuxer or filter).
        *   Retrieve the output file content from `ffmpeg.wasm`.
        *   Return the merged video data as `[]byte`.
4.  **Implement Progress Reporting**:
    *   Modify `mergeVideosGo` to accept a JavaScript callback function for progress.
    *   `ffmpeg.wasm` provides progress events. Capture these in the JS glue code (or Go if possible) and use the callback to update the frontend.
    *   `mergeVideosGo(fileNames []string, fileContents [][]byte, progressCb js.Value) ([]byte, error)`.
5.  **Compile and Test**:
    *   Compile the Go code to `main.wasm`.
    *   Create a minimal HTML/JS page to load `main.wasm` and `ffmpeg.wasm.js`.
    *   Test the `mergeVideosGo` function with sample video files.
    *   Debug issues related to data transfer, `ffmpeg.wasm` interaction, and Wasm execution.

## Phase 2: Frontend - User Interface & Interaction

**Goal**: Develop the frontend HTML, CSS, and JavaScript to interact with the Wasm module.

1.  **HTML Structure (`index.html`)**:
    *   Create the basic page layout:
        *   Title.
        *   Button for directory selection (styling a hidden `input type="file" webkitdirectory`).
        *   Area to display selected directory path.
        *   "Merge Videos" button (initially hidden/disabled).
        *   Progress bar/area (initially hidden).
        *   Area for download link/status messages.
    *   Include `wasm_exec.js` (from Go installation) and a script tag for custom JS (`app.js`).
    *   Include `ffmpeg.wasm.js`.
2.  **CSS Styling (`style.css`)**:
    *   Basic styling for a presentable user interface.
3.  **JavaScript Logic (`app.js`)**:
    *   **Wasm Loading**:
        *   Load `main.wasm` (Go Wasm) using `wasm_exec.js`.
        *   Load `ffmpeg.wasm` and instantiate it.
    *   **Directory Selection**:
        *   Implement event listener for the directory selection button.
        *   When a directory is selected:
            *   Store the file list (`FileList` object).
            *   Display the directory path.
            *   Enable the "Merge Videos" button.
    *   **Merge Process Initiation**:
        *   Event listener for the "Merge Videos" button.
        *   When clicked:
            *   Disable the merge button to prevent multiple clicks.
            *   Show progress bar.
            *   Read file contents: Iterate through the `FileList`, filter for video files (e.g., by extension: `.mp4`, `.webm`, `.ogv`, `.mov`).
            *   Sort files by name.
            *   Read each selected video file as an `ArrayBuffer` or `Uint8Array`.
            *   Collect file names and their contents.
    *   **Calling Wasm `mergeVideosGo` function**:
        *   Prepare data (file names, file contents) in a format suitable for `syscall/js`.
        *   Create a JS progress callback function that updates the frontend progress bar.
        *   Call the exported `mergeVideosGo` function from the Go Wasm module.
    *   **Handling Wasm Response**:
        *   On success:
            *   Receive the merged video data (byte array).
            *   Create a Blob from the data with the correct MIME type (e.g., `video/mp4`).
            *   Create a download link for the Blob and make it visible/clickable. Suggested filename: `merged_video.mp4`.
            *   Attempt to save to the originally selected directory (if using File System Access API and permissions allow, otherwise rely on download).
            *   Update UI to show completion.
        *   On error:
            *   Display an error message to the user.
    *   **Progress Display**:
        *   The JS progress callback (passed to Wasm) will update the progress bar element (`<progress value=".." max="100">`).

## Phase 3: Integration, Testing, and Refinement

**Goal**: Ensure all parts work together smoothly, test thoroughly, and refine.

1.  **End-to-End Testing**:
    *   Test with various numbers of videos.
    *   Test with different (but compatible for `ffmpeg.wasm` concatenation) video files.
    *   Test edge cases:
        *   No directory selected.
        *   Directory with no video files.
        *   Directory with one video file.
        *   Very small/large video files (monitor browser memory).
2.  **Error Handling Refinement**:
    *   Ensure all potential errors (file reading, Wasm execution, video merging) are caught and reported gracefully.
3.  **UI/UX Improvements**:
    *   Make the UI more intuitive.
    *   Improve progress reporting (e.g., current file being processed, overall percentage).
4.  **Code Cleanup and Optimization**:
    *   Refactor JavaScript and Go code for clarity and efficiency.
    *   Minimize data copying between JS and Go Wasm where possible.
5.  **Cross-browser Testing (Basic)**:
    *   Test in modern versions of Chrome, Firefox, and Safari (especially for `webkitdirectory` and Wasm support). `ffmpeg.wasm` itself has browser compatibility notes.

## Timeline Estimation (High-Level)

*   **Phase 1 (Backend - Go/Wasm)**: 3-5 days (Includes research and `ffmpeg.wasm` integration, which can be tricky).
*   **Phase 2 (Frontend - UI/JS)**: 2-3 days.
*   **Phase 3 (Integration & Testing)**: 2-3 days.

**Total Estimated Time**: 7-11 days. This is a rough estimate and can vary based on the complexity encountered, especially with `ffmpeg.wasm` and video processing aspects.

## Key Technologies & Files

*   `index.html`: Main HTML file.
*   `style.css`: CSS for styling.
*   `app.js`: Frontend JavaScript application logic.
*   `go/main.go`: Go source code for the Wasm module.
*   `go/main.wasm`: Compiled Wasm module (output).
*   `wasm_exec.js`: Go-provided JavaScript file to run Go Wasm.
*   `ffmpeg.wasm.js` / `ffmpeg-core.js` / etc.: Files related to `ffmpeg.wasm`.
*   `README.md`: Project overview.
*   `design.md`: Design document.
*   `dev.md`: This development plan.

## Potential Roadblocks & Mitigation

*   **Complexity of `ffmpeg.wasm` Integration**: This is the highest risk. Mitigation: Start with the simplest `ffmpeg.wasm` examples, allocate sufficient time for this, and consult its documentation thoroughly.
*   **Performance with Large Files/Many Videos**: Client-side processing is limited. Mitigation: Advise users on limitations. For a V2, explore chunking or server-side solutions if client-side is insufficient.
*   **Browser Compatibility**: Wasm and advanced file APIs vary. Mitigation: Target modern browsers. Provide fallbacks (e.g., simple download link instead of direct save to directory).
*   **Debugging Wasm**: Can be challenging. Mitigation: Use browser developer tools, add extensive logging in both Go and JS.

This plan provides a structured approach to development. Flexibility will be needed as challenges arise.
