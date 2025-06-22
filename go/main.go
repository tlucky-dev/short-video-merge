package main

import (
	"fmt"
	"strings"
	"syscall/js"
)

// helloWorld is a simple function that will be callable from JavaScript.
// It returns a greeting string. Kept for basic Wasm functionality testing.
func helloWorld(this js.Value, args []js.Value) interface{} {
	fmt.Println("Go: helloWorld function called from JavaScript")
	if len(args) > 0 && args[0].Type() == js.TypeString {
		name := args[0].String()
		return fmt.Sprintf("Hello, %s! This is Go speaking (from helloWorld).", name)
	}
	return "Hello from Go! You didn't provide a name (from helloWorld)."
}

// mergeVideosGo handles the video merging logic.
// It expects:
// args[0]: array of file names (js.Value wrapping a JS array of strings)
// args[1]: array of file contents (js.Value wrapping a JS array of Uint8Array)
// args[2]: progress callback function (js.Value wrapping a JS function)
// Returns a JS Promise which resolves with a Uint8Array (merged video data) or rejects with an error.
func mergeVideosGo(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		errorMsg := "Go: mergeVideosGo expects 3 arguments: fileNames (Array<string>), fileContents (Array<Uint8Array>), progressCallback (function)"
		fmt.Println(errorMsg) // Keep this error visible
		// Immediately return a rejected promise
		errorObject := js.Global().Get("Error").New(errorMsg)
		return js.Global().Get("Promise").Call("reject", errorObject)
	}

	fileNamesJS := args[0]
	fileContentsJS := args[1]
	progressCallbackJS := args[2] // Now we'll use this

	if fileNamesJS.Type() != js.TypeObject || fileNamesJS.Get("length").Type() != js.TypeNumber {
		errorMsg := "Go: fileNames argument must be an array."
		fmt.Println(errorMsg)
		errorObject := js.Global().Get("Error").New(errorMsg)
		return js.Global().Get("Promise").Call("reject", errorObject)
	}
	if fileContentsJS.Type() != js.TypeObject || fileContentsJS.Get("length").Type() != js.TypeNumber {
		errorMsg := "Go: fileContents argument must be an array."
		fmt.Println(errorMsg)
		errorObject := js.Global().Get("Error").New(errorMsg)
		return js.Global().Get("Promise").Call("reject", errorObject)
	}
	if progressCallbackJS.Type() != js.TypeFunction {
		errorMsg := "Go: progressCallback argument must be a function."
		fmt.Println(errorMsg)
		errorObject := js.Global().Get("Error").New(errorMsg)
		return js.Global().Get("Promise").Call("reject", errorObject)
	}

	numFiles := fileNamesJS.Length()
	if numFiles == 0 {
		errorMsg := "Go: No files provided for merging."
		fmt.Println(errorMsg)
		errorObject := js.Global().Get("Error").New(errorMsg)
		return js.Global().Get("Promise").Call("reject", errorObject)
	}
	if numFiles != fileContentsJS.Length() {
		errorMsg := "Go: Mismatch between number of file names and file contents."
		fmt.Println(errorMsg)
		errorObject := js.Global().Get("Error").New(errorMsg)
		return js.Global().Get("Promise").Call("reject", errorObject)
	}

	// Create a new Promise
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(js.FuncOf(func(this_ js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		// Run FFmpeg processing in a goroutine to avoid blocking and allow Promise to return
		go func() {
			// fmt.Println("Go: Starting video merging process in goroutine...") // Less verbose

			ffmpeg := js.Global().Get("ffmpeg") // Get ffmpeg instance from JS global scope
			if ffmpeg.IsUndefined() {
				errStr := "Go: ffmpeg instance not found in JavaScript global scope."
				fmt.Println(errStr) // Keep error
				reject.Invoke(js.Global().Get("Error").New(errStr))
				return
			}

			// 1. Write files to FFmpeg's virtual file system & create concat list
			var concatFileContentBuilder strings.Builder
			for i := 0; i < numFiles; i++ {
				name := fileNamesJS.Index(i).String()
				data := fileContentsJS.Index(i) // This is a Uint8Array

				// fmt.Printf("Go: Writing file '%s' to FFmpeg FS\n", name) // Less verbose
				ffmpeg.Call("FS", "writeFile", name, data)
				concatFileContentBuilder.WriteString(fmt.Sprintf("file '%s'\n", name))
			}

			concatListFileName := "mylist.txt"
			concatListContent := concatFileContentBuilder.String()
			// fmt.Printf("Go: Writing concat list file '%s' to FFmpeg FS with content:\n%s\n", concatListFileName, concatListContent) // Less verbose
			ffmpeg.Call("FS", "writeFile", concatListFileName, js.ValueOf(concatListContent))

			// 2. Construct and run FFmpeg command
			// Using concat demuxer: ffmpeg -f concat -safe 0 -i mylist.txt -c copy output.mp4
			// The `-safe 0` is needed if filenames contain special characters or are absolute paths (not the case here, but good practice for concat demuxer).
			// `-c copy` attempts to copy codecs without re-encoding. This requires inputs to have compatible streams.
			outputFileName := "output.mp4"
			ffmpegArgs := []interface{}{
				"-f", "concat",
				"-safe", "0", // Allow relative paths in concat list
				"-i", concatListFileName,
				"-c", "copy", // Copy codecs, no re-encoding
				outputFileName,
			}

			// fmt.Printf("Go: Running FFmpeg command: ffmpeg %v\n", ffmpegArgs) // Less verbose

			// The `run` command itself might be asynchronous if it uses web workers.
			// However, the `ffmpeg.run` in `ffmpeg.wasm` v0.11+ is synchronous if not using `createFFmpeg({thread: true})`
			// For v0.12, it's generally synchronous for the main thread version.
			// We will assume it's synchronous here and await its completion within the goroutine.
			// If it were internally async and returned a Promise, we'd need to handle that.
			// The `ffmpeg.run` method does not return a Promise. It blocks until complete.


			// Setup progress reporting
			var goProgressCb js.Func
			goProgressCb = js.FuncOf(func(this_ js.Value, cbArgs []js.Value) interface{} {
				if len(cbArgs) > 0 && cbArgs[0].Type() == js.TypeObject {
					progressObj := cbArgs[0]
					ratio := progressObj.Get("ratio").Float() // ratio is 0.0 to 1.0
					// fmt.Printf("Go Progress: Ratio %.2f\n", ratio) // For debugging
					progressCallbackJS.Invoke(ratio * 100) // Send percentage to JS
				}
				return nil
			})
			defer goProgressCb.Release() // Release the callback when the goroutine finishes

			ffmpeg.Call("setProgress", goProgressCb)
			fmt.Println("Go: Registered FFmpeg progress callback.")

			defer func() {
				// Clear progress callback after run, in case ffmpeg instance is reused.
				// This might not be strictly necessary if a new ffmpeg instance is created each time,
				// but good practice. Passing null or undefined usually unsets it.
				ffmpeg.Call("setProgress", js.Null())
				fmt.Println("Go: Cleared FFmpeg progress callback.")

				if r := recover(); r != nil {
					errMsg := fmt.Sprintf("Go: Panic recovered in ffmpeg.run: %v", r)
					fmt.Println(errMsg)
					reject.Invoke(js.Global().Get("Error").New(errMsg))
				}
			}()

			// Ensure ffmpeg.run is called correctly. It expects variadic string arguments.
			cmdArgsJS := make([]js.Value, len(ffmpegArgs))
			for i, arg := range ffmpegArgs {
				cmdArgsJS[i] = js.ValueOf(arg)
			}

			// The `run` method is on the ffmpeg instance.
			// `ffmpeg.Call("run", ...)`
			// The arguments need to be passed as individual arguments to Call, not as a slice.
			// So, convert interface{} slice to js.Value slice and then to individual args for Call.

			runArgs := make([]interface{}, len(ffmpegArgs))
			for i, v := range ffmpegArgs {
				runArgs[i] = v
			}

			// fmt.Println("Go: Invoking ffmpeg.run...") // Less verbose
			ffmpeg.Call("run", runArgs...) // Pass arguments as variadic
			// fmt.Println("Go: FFmpeg command finished.") // Less verbose

			// 3. Read the output file
			// fmt.Printf("Go: Reading output file '%s' from FFmpeg FS\n", outputFileName) // Less verbose
			mergedVideoData := ffmpeg.Call("FS", "readFile", outputFileName) // Returns Uint8Array

			if mergedVideoData.IsUndefined() || mergedVideoData.IsNull() {
				errStr := fmt.Sprintf("Go: Failed to read merged file '%s' from FFmpeg FS.", outputFileName)
				fmt.Println(errStr) // Keep error
				reject.Invoke(js.Global().Get("Error").New(errStr))
				return
			}

			// fmt.Println("Go: Successfully merged video. Resolving promise.") // Less verbose
			resolve.Invoke(mergedVideoData)
		}()

		return nil // Important: Handler for a Promise constructor must return undefined or null
	}))
}

func main() {
	fmt.Println("Go: WebAssembly module loaded and main function executed.") // Simplified startup message
	c := make(chan struct{}, 0) // Channel to keep the Go program alive

	// Export functions to JavaScript
	js.Global().Set("goHelloWorld", js.FuncOf(helloWorld))     // Kept for basic testing
	js.Global().Set("mergeVideosGo", js.FuncOf(mergeVideosGo)) // Main application function

	// fmt.Println("Go: 'goHelloWorld' and 'mergeVideosGo' functions exported.") // Less verbose on startup

	<-c // Block main from exiting, allowing JS to call exported functions
}
