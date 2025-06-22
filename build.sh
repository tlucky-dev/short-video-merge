#!/bin/bash

# build.sh - Script to package the video merger tool for Linux

PYTHON_CMD="python3"
PIP_CMD="pip3"
VENV_DIR="venv_linux"

# Check for Python and Pip
if ! command -v $PYTHON_CMD &> /dev/null
then
    echo "$PYTHON_CMD could not be found. Please install Python 3."
    exit 1
fi

if ! command -v $PIP_CMD &> /dev/null
then
    echo "$PIP_CMD could not be found. Please install pip for Python 3."
    exit 1
fi

echo "Creating virtual environment in $VENV_DIR..."
$PYTHON_CMD -m venv $VENV_DIR

if [ $? -ne 0 ]; then
    echo "Failed to create virtual environment."
    exit 1
fi

echo "Activating virtual environment..."
source $VENV_DIR/bin/activate

if [ $? -ne 0 ]; then
    echo "Failed to activate virtual environment."
    # Attempt to clean up venv dir if activation fails
    rm -rf $VENV_DIR
    exit 1
fi

echo "Installing dependencies from requirements.txt..."
$PIP_CMD install --upgrade pip
$PIP_CMD install -r requirements.txt
$PIP_CMD install pyinstaller

if [ $? -ne 0 ]; then
    echo "Failed to install dependencies."
    deactivate
    rm -rf $VENV_DIR
    exit 1
fi

echo "Packaging the application with PyInstaller..."
# PyInstaller options:
# --onefile: Create a single executable file
# --name: Name of the executable
# --clean: Clean PyInstaller cache and remove temporary files before building
# video_merger.py: Your main Python script
pyinstaller --onefile --name video_merger_linux video_merger.py --clean

if [ $? -ne 0 ]; then
    echo "PyInstaller packaging failed."
    deactivate
    rm -rf $VENV_DIR
    exit 1
fi

echo "Deactivating virtual environment..."
deactivate

# Optional: Clean up build artifacts except the final executable in dist/
# echo "Cleaning up build artifacts..."
# rm -rf build/
# rm -f *.spec
# mv dist/* . # Move executable to current directory (optional)
# rm -rf dist/ # If executable moved

echo ""
echo "Build successful!"
echo "The executable 'video_merger_linux' can be found in the 'dist' directory."
echo "You might want to move it to a directory in your PATH, e.g., /usr/local/bin/"
echo ""
echo "To run: ./dist/video_merger_linux --dir /path/to/videos --output merged_video.mp4"

exit 0
