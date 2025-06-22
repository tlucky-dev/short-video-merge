import os
import argparse
from moviepy.editor import VideoFileClip, concatenate_videoclips
from moviepy.config import change_settings

def find_video_files(directory):
    """Finds video files in the specified directory."""
    video_extensions = ['.mp4', '.avi', '.mov', '.mkv', '.flv', '.wmv']
    video_files = []
    if not os.path.isdir(directory):
        print(f"Error: Directory '{directory}' not found.")
        return None
    for item in os.listdir(directory):
        if os.pathisfile(os.path.join(directory, item)):
            _, ext = os.path.splitext(item)
            if ext.lower() in video_extensions:
                video_files.append(os.path.join(directory, item))
    return sorted(video_files) # Sort to ensure consistent order

def merge_videos(video_files, output_filename):
    """Merges a list of video files into a single output file."""
    if not video_files:
        print("No video files found to merge.")
        return False

    print(f"Found {len(video_files)} video files to merge:")
    for vf in video_files:
        print(f"  - {vf}")

    clips = []
    try:
        for video_file in video_files:
            try:
                clip = VideoFileClip(video_file)
                clips.append(clip)
            except Exception as e:
                print(f"Warning: Could not load video file '{video_file}': {e}")
                print("This file will be skipped.")

        if not clips:
            print("No valid video clips could be loaded. Merging aborted.")
            return False

        final_clip = concatenate_videoclips(clips, method="compose")
        final_clip.write_videofile(output_filename, codec="libx264", audio_codec="aac")

        # Close all clips
        for clip in clips:
            clip.close()
        if hasattr(final_clip, 'close') and callable(final_clip.close):
            final_clip.close()

        print(f"\nSuccessfully merged videos into '{output_filename}'")
        return True
    except ImportError:
        print("MoviePy library not found. Please install it: pip install moviepy")
        return False
    except OSError as e:
        if "ffmpeg" in str(e).lower() or "imageio" in str(e).lower() :
            print(f"Error: FFmpeg might be missing or not configured correctly for MoviePy. {e}")
            print("Please ensure FFmpeg is installed and accessible in your system's PATH.")
            print("You can download it from https://ffmpeg.org/download.html")
            print("If FFmpeg is installed, you might need to configure MoviePy by editing moviepy/config_defaults.py or setting FFMPEG_BINARY environment variable.")
            # Attempt to guide user for manual config if auto-detection fails
            # This is a common issue with moviepy
            try:
                # Example of how one might try to set it if they know the path
                # change_settings({"FFMPEG_BINARY": "/path/to/your/ffmpeg"}) # Or "ffmpeg.exe" on Windows
                print("\nIf you have ffmpeg installed but not in PATH, you can try setting the FFMPEG_BINARY path in moviepy's config_defaults.py")
                print("Or, set the FFMPEG_BINARY environment variable pointing to your ffmpeg executable.")
            except Exception as config_e:
                print(f"Could not apply dynamic FFMPEG_BINARY setting: {config_e}")

        else:
            print(f"An OS error occurred: {e}")
        return False
    except Exception as e:
        print(f"An unexpected error occurred during merging: {e}")
        return False

def main():
    parser = argparse.ArgumentParser(description="Merge video files from a directory.")
    parser.add_argument("--dir", type=str, required=True, help="Directory containing video files to merge.")
    parser.add_argument("--output", type=str, required=True, help="Filename for the merged output video.")

    args = parser.parse_args()

    video_files = find_video_files(args.dir)

    if video_files is None: # Directory not found
        return

    if not video_files:
        print(f"No video files found in directory '{args.dir}'.")
        return

    merge_videos(video_files, args.output)

if __name__ == "__main__":
    main()
