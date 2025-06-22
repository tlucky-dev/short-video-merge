#  短视频合并

## 背景

开发一个前端项目，用于短视频的合并。

## 要求

1. 前端界面元素如下
  1.1 一个选择按钮，点击后，可以选择本地目录，并在界面上展示出本地目录路径
  1.2 选择目录之后，出现合并按钮
  1.3 点击合并按钮后，根据目录中文件名的顺序，将多个视频进行合并，合并成一个视频
  1.4 合并时，界面实时显示合并的进度
  1.5 最终合并好之后，将合并之后的文件存储到选中的目录中
2. 技术相关
   2.1 后端使用golang开发，编译成Wasm
   2.2 前端加载wasm，用于视频的合成处理

## How to Run

1.  **Prerequisites**:
    *   Go (Golang) installed (for compiling Go to Wasm). Tested with Go 1.18+.
    *   A modern web browser that supports WebAssembly and the `webkitdirectory` attribute for directory selection (e.g., Chrome, Firefox, Edge).
    *   `wasm_exec.js` file: This file is part of your Go installation. You need to copy it to the root directory of this project.
        *   On Unix-like systems, you can usually find it at `$(go env GOROOT)/misc/wasm/wasm_exec.js`.
        *   Copy it using: `cp $(go env GOROOT)/misc/wasm/wasm_exec.js .` (from the project root).

2.  **Build the Wasm Module**:
    *   Navigate to the project root directory in your terminal.
    *   Run the command: `GOOS=js GOARCH=wasm go build -o main.wasm go/main.go`
    *   This will create/update the `main.wasm` file in the project root.

3.  **Serve the Application**:
    *   You need to serve the `index.html` file through a local HTTP server. Opening it directly as a file (`file://...`) will likely not work due to browser security restrictions (CORS, Wasm loading).
    *   A simple way to do this is using Python's built-in HTTP server:
        *   If you have Python 3: `python -m http.server`
        *   If you have Python 2: `python -m SimpleHTTPServer`
    *   Alternatively, you can use `npx serve` (requires Node.js): `npx serve .`
    *   The server will typically start on `http://localhost:8000` or a similar address. Open this address in your web browser.

4.  **Usage**:
    *   Click the "Select Directory" button and choose a directory containing video files.
    *   The application will list the detected video files (sorted by name).
    *   Click the "Merge Videos" button.
    *   Wait for the merging process to complete. Progress will be displayed.
    *   Once done, a download link for the `merged_video.mp4` will appear.

## Limitations & Notes

*   **Video Compatibility**: The merging process uses `ffmpeg.wasm` with the `-c copy` flag. This means it attempts to concatenate video streams without re-encoding. For this to work reliably, the input video files should have **compatible codecs, resolutions, frame rates, and other stream parameters**. If videos are incompatible, the merge might fail or the output video might be corrupted.
*   **Performance**: Client-side video processing is resource-intensive. Merging large files or many files can be slow and consume significant browser memory and CPU. The browser might become unresponsive during very heavy operations.
*   **Browser Support**: Relies on modern browser features like WebAssembly, `FileList`, `FileReader`, `URL.createObjectURL`, and `input[type=file][webkitdirectory]`. `ffmpeg.wasm` itself has specific browser requirements (e.g., SharedArrayBuffer support, which might require specific server headers like COOP/COEP for full performance/threading, though the current implementation doesn't explicitly enable ffmpeg threading).
*   **Error Handling**: Basic error handling is in place. If FFmpeg encounters an issue with the files or the merging process, an error message should be displayed. More detailed FFmpeg logs are available in the browser's developer console.
*   **File Order**: Videos are merged in alphanumeric order of their filenames from the selected directory.
*   **Output Format**: The output is always `merged_video.mp4`.
*   **`wasm_exec.js`**: Ensure this file is present. The application will not run without it.
