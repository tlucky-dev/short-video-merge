@echo off
REM build.bat - Script to package the video merger tool for Windows

SET PYTHON_CMD=python
SET PIP_CMD=pip
SET VENV_DIR=venv_windows

REM Check for Python
%PYTHON_CMD% --version > nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    echo Python could not be found. Please install Python 3 and ensure it's in your PATH.
    goto :eof
)

REM Check for Pip
%PIP_CMD% --version > nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    echo Pip could not be found. Please ensure pip is installed and in your PATH.
    goto :eof
)

echo Creating virtual environment in %VENV_DIR%...
%PYTHON_CMD% -m venv %VENV_DIR%
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to create virtual environment.
    goto :eof
)

echo Activating virtual environment...
CALL %VENV_DIR%\Scripts\activate.bat
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to activate virtual environment.
    rmdir /s /q %VENV_DIR%
    goto :eof
)

echo Installing dependencies from requirements.txt...
%PIP_CMD% install --upgrade pip
%PIP_CMD% install -r requirements.txt
%PIP_CMD% install pyinstaller
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to install dependencies.
    call deactivate
    rmdir /s /q %VENV_DIR%
    goto :eof
)

echo Packaging the application with PyInstaller...
REM PyInstaller options:
REM --onefile: Create a single executable file
REM --name: Name of the executable
REM --clean: Clean PyInstaller cache and remove temporary files before building
REM --noconsole: (Optional) Prevents console window for GUI, but needed for CLI.
REM video_merger.py: Your main Python script
pyinstaller --onefile --name video_merger_windows video_merger.py --clean
IF %ERRORLEVEL% NEQ 0 (
    echo PyInstaller packaging failed.
    call deactivate
    rmdir /s /q %VENV_DIR%
    goto :eof
)

echo Deactivating virtual environment...
call deactivate

REM Optional: Clean up build artifacts except the final executable in dist/
REM echo Cleaning up build artifacts...
REM rmdir /s /q build
REM del *.spec
REM xcopy dist\* . /Y (Move executable to current directory - optional)
REM rmdir /s /q dist (If executable moved)

echo.
echo Build successful!
echo The executable 'video_merger_windows.exe' can be found in the 'dist' directory.
echo You might want to move it to a directory in your PATH.
echo.
echo To run: dist\video_merger_windows.exe --dir C:\path\to\videos --output merged_video.mp4

:eof
exit /b 0
