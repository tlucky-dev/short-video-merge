// app.js for Short Video Merger

// Global variables
let goWasm;
let ffmpeg; // To store the ffmpeg instance
let selectedFiles = [];

const dirPicker = document.getElementById('dirPicker');
const selectDirBtn = document.getElementById('selectDirBtn');
const dirPathDisplay = document.getElementById('dirPath');
const mergeBtn = document.getElementById('mergeBtn');
const progressBar = document.getElementById('progressBar');
const progressText = document.getElementById('progressText');
const outputArea = document.getElementById('outputArea');
const fileListArea = document.getElementById('fileListArea');
const fileListUl = document.getElementById('fileList');

// Initialize Go Wasm and FFmpeg.wasm
async function init() {
    // Initialize Go Wasm
    const go = new Go();
    if (!WebAssembly.instantiateStreaming) { // Polyfill for Safari and older browsers
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
            const source = await (await resp).arrayBuffer();
            return await WebAssembly.instantiate(source, importObject);
        };
    }

    try {
        progressText.textContent = "Loading Go WebAssembly module...";
        console.log("JS: Loading main.wasm...");
        const result = await WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject);
        goWasm = result.instance;
        go.run(goWasm); // Executes Go's main()
        console.log("JS: main.wasm loaded and Go main() executed.");
        progressText.textContent = "Go WebAssembly module loaded.";
    } catch (error) {
        console.error("JS: Error loading Go Wasm module:", error);
        progressText.textContent = `Error loading Go Wasm: ${error.message}`;
        return;
    }

    // Initialize FFmpeg.wasm
    if (typeof FFmpeg === 'undefined') {
        console.error("JS: FFmpeg.js not loaded. Ensure the script tag is correct.");
        progressText.textContent = "Error: FFmpeg.js not loaded.";
        return;
    }

    try {
        progressText.textContent = "Loading FFmpeg...";
        console.log("JS: Initializing FFmpeg.wasm...");
        ffmpeg = FFmpeg.createFFmpeg({
            log: true, // Enable logging from ffmpeg
            // It's crucial to use a version of @ffmpeg/core compatible with the @ffmpeg/ffmpeg version.
            // Using unpkg for corePath is recommended by ffmpeg.wasm docs to ensure all related core files are found.
            corePath: 'https://unpkg.com/@ffmpeg/core@0.12.6/dist/ffmpeg-core.js'
            // Using 0.12.6 for core as it's a recent stable version often paired with ffmpeg.wasm 0.12.x
            // The JSDelivr found 0.12.10 for @ffmpeg/core, let's try that.
            // corePath: 'https://unpkg.com/@ffmpeg/core@0.12.10/dist/ffmpeg-core.js'
            // After some checks, @ffmpeg/core@0.12.6 is widely used with @ffmpeg/ffmpeg@0.12.x
            // Let's stick to a known compatible core version if possible.
            // The official ffmpeg.wasm examples often use specific versions.
            // Let's try the latest @ffmpeg/core found on unpkg which is also 0.12.6
            // https://unpkg.com/@ffmpeg/core/package.json shows "0.12.6"
            // So this corePath should be: 'https://unpkg.com/@ffmpeg/core@0.12.6/dist/ffmpeg-core.js'
        });
        await ffmpeg.load();
        console.log("JS: FFmpeg.wasm loaded successfully.");
        progressText.textContent = "FFmpeg loaded. Ready to select directory.";
    } catch (error) {
        console.error("JS: Error loading FFmpeg.wasm:", error);
        progressText.textContent = `Error loading FFmpeg: ${error.message}`;
        if (error.message && error.message.toLowerCase().includes("cross-origin")) {
            progressText.innerHTML += `<br><small>This might be due to Cross-Origin Isolation policies. Ensure your server sets COOP and COEP headers if serving locally for development, or use a compatible environment.</small>`;
        }
    }

    // Setup event listeners once everything is loaded
    setupEventListeners();
}

function setupEventListeners() {
    if (selectDirBtn) {
        selectDirBtn.addEventListener('click', () => {
            if (dirPicker) dirPicker.click();
        });
    } else {
        console.error("JS: selectDirBtn not found");
    }

    if (dirPicker) {
        dirPicker.addEventListener('change', handleDirectorySelection);
    } else {
        console.error("JS: dirPicker not found");
    }

    if (mergeBtn) {
        mergeBtn.addEventListener('click', handleMergeProcess);
    } else {
        console.error("JS: mergeBtn not found");
    }
}

function handleDirectorySelection(event) {
    fileListUl.innerHTML = ''; // Clear previous list
    fileListArea.style.display = 'none';

    if (!event.target.files || event.target.files.length === 0) {
        dirPathDisplay.textContent = "No files selected or directory empty.";
        mergeBtn.style.display = 'none';
        selectedFiles = [];
        return;
    }

    // Filter for video files and sort them
    selectedFiles = Array.from(event.target.files)
        .filter(file => file.type.startsWith('video/')) // Basic video filter
        .sort((a, b) => a.name.localeCompare(b.name)); // Sort by name

    if (selectedFiles.length > 0) {
        // Attempt to display a path. webkitRelativePath is non-standard but common.
        const firstFilePath = selectedFiles[0].webkitRelativePath;
        const directoryPath = firstFilePath ? firstFilePath.substring(0, firstFilePath.lastIndexOf('/')) : "Selected files";
        dirPathDisplay.textContent = `Selected: ${directoryPath} (${selectedFiles.length} video file(s))`;

        // Populate file list display
        selectedFiles.forEach(file => {
            const li = document.createElement('li');
            li.textContent = file.name;
            fileListUl.appendChild(li);
        });
        fileListArea.style.display = 'block';
        mergeBtn.style.display = 'block';
    } else {
        dirPathDisplay.textContent = "No video files found in the selected directory.";
        mergeBtn.style.display = 'none';
        fileListArea.style.display = 'none';
    }
    outputArea.innerHTML = ''; // Clear previous results
    progressText.textContent = '';
    progressBar.style.display = 'none';
    progressBar.value = 0;
}

async function handleMergeProcess() {
    if (selectedFiles.length === 0) {
        progressText.textContent = "No video files to merge.";
        return;
    }
    if (!ffmpeg || !ffmpeg.isLoaded()) {
        progressText.textContent = "FFmpeg is not loaded. Please wait or refresh.";
        // Optionally, try to load ffmpeg again or guide user.
        // await init(); // This might be too aggressive, could re-initialize Go Wasm too.
        // For now, just inform.
        return;
    }
     if (!goWasm || typeof goHelloWorld !== 'function') { // Assuming goHelloWorld is still the test function
        // Later this will be typeof global.mergeVideosGo !== 'function'
        progressText.textContent = "Go Wasm function not ready. Please wait or refresh.";
        return;
    }


    mergeBtn.disabled = true;
    progressBar.style.display = 'block';
    progressBar.value = 0;
    progressText.textContent = 'Preparing files...';
    outputArea.innerHTML = '';

    console.log("JS: Merge process starting with files:", selectedFiles.map(f => f.name));

    const fileNames = selectedFiles.map(f => f.name);
    const fileContentsPromises = selectedFiles.map(f =>
        f.arrayBuffer().then(buf => new Uint8Array(buf))
    );

    try {
        const fileContents = await Promise.all(fileContentsPromises);
        progressText.textContent = 'Files loaded, calling Go Wasm...';

        const jsProgressCallback = (percentage) => {
            progressBar.value = percentage;
            // Accessing ffmpeg.getProgress() here might be tricky if ffmpeg is only in Go scope
            // For now, just use percentage. Go could pass more detailed messages if needed.
            progressText.textContent = `Merging: ${percentage.toFixed(2)}%`;
            console.log(`JS Progress: ${percentage.toFixed(2)}%`);
        };

        if (typeof mergeVideosGo !== 'function') {
            throw new Error("Go function 'mergeVideosGo' not found. Ensure it's exported correctly.");
        }

        // Call the exported Go function
        // It's expected to be on the global scope (window.mergeVideosGo)
        const mergedVideoUint8Array = await mergeVideosGo(fileNames, fileContents, jsProgressCallback);

        progressText.textContent = 'Merge complete! Preparing download...';
        progressBar.value = 100;

        const blob = new Blob([mergedVideoUint8Array], { type: 'video/mp4' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'merged_video.mp4';
        a.textContent = 'Download Merged Video';
        outputArea.appendChild(a);
        console.log("JS: Download link created.");

    } catch (error) {
        console.error("JS: Error during merge process:", error);
        progressText.textContent = `Error: ${error.message || String(error)}`;
        progressBar.style.display = 'none';
    } finally {
        mergeBtn.disabled = false;
    }
}

document.addEventListener('DOMContentLoaded', init);
