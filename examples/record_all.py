#!/usr/bin/env python3
"""Record all BasementUI examples as asciicast files and convert to GIF."""

import json
import os
import subprocess
import sys
import time
import pexpect

PROJ_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
GO_DIR = os.path.join(PROJ_DIR, "go")
EXAMPLES_DIR = os.path.join(PROJ_DIR, "examples")
BIN_DIR = os.path.join(EXAMPLES_DIR, "bin")
DEFAULT_COLS = 60
DEFAULT_ROWS = 15
LARGE_COLS= 80
LARGE_ROWS = 40

# ANSI escape sequences for arrow keys
UP = "\x1b[A"
DOWN = "\x1b[B"
RIGHT = "\x1b[C"
LEFT = "\x1b[D"
ESC = "\x1b"
ENTER = "\r"
BACKSPACE = "\x7f"


def build_example(name, command):
    """Pre-build example binary for instant startup."""
    os.makedirs(BIN_DIR, exist_ok=True)
    bin_path = os.path.join(BIN_DIR, name)

    # Extract the source path and build tags from the go run command
    parts = command.split()
    tags = ""
    src = parts[-1]  # Last arg is the source path
    if "-tags" in command:
        idx = parts.index("-tags")
        tags = f"-tags {parts[idx+1]}"

    build_cmd = f"cd {GO_DIR} && go build {tags} -o {bin_path} {src}"
    print(f"  Building {name}...")
    ret = os.system(build_cmd + " 2>&1")
    if ret != 0:
        print(f"  ERROR: Build failed for {name}")
        return None
    return bin_path


def record_cast(name, bin_path, actions, cols=DEFAULT_COLS, rows=DEFAULT_ROWS):
    """
    Record an asciicast v2 file by running the pre-built binary.

    actions: list of tuples (delay_seconds, keys_to_send_or_None)
    """
    cast_path = os.path.join(EXAMPLES_DIR, f"{name}.cast")

    print(f"  Recording {name}...")

    # Start the pre-built binary in a PTY
    child = pexpect.spawn(
        bin_path,
        encoding=None,  # binary mode
        dimensions=(rows, cols),
        timeout=30,
    )

    # Collect output events: (timestamp, "o", data)
    events = []
    start_time = time.time()

    def capture_output(wait=0.1):
        """Read any available output from the child."""
        try:
            while True:
                data = child.read_nonblocking(size=16384, timeout=wait)
                if data:
                    t = time.time() - start_time
                    events.append((t, "o", data.decode("utf-8", errors="replace")))
                    wait = 0.005  # Capture more frames (was 0.01)
        except (pexpect.TIMEOUT, pexpect.EOF):
            pass

    # Wait for initial render
    time.sleep(0.1)
    capture_output(wait=0.05)

    # Execute the scripted actions
    for action in actions:
        if isinstance(action, tuple):
            delay, keys = action
            if delay > 0:
                time.sleep(delay)
                capture_output()
            if keys is not None:
                if isinstance(keys, bytes):
                    child.send(keys)
                else:
                    child.send(keys.encode("utf-8"))
                time.sleep(0.05)
                capture_output()

    # Wait for process to exit
    try:
        child.expect(pexpect.EOF, timeout=3)
        final_data = child.before
        if final_data:
            t = time.time() - start_time
            events.append((t, "o", final_data.decode("utf-8", errors="replace")))
    except (pexpect.TIMEOUT, pexpect.EOF):
        child.terminate(force=True)

    # Write asciicast v2 file
    header = {
        "version": 2,
        "width": cols,
        "height": rows,
        "timestamp": int(start_time),
        "env": {"SHELL": "/bin/bash", "TERM": "xterm-256color"},
    }

    with open(cast_path, "w") as f:
        f.write(json.dumps(header) + "\n")
        for t, etype, data in events:
            f.write(json.dumps([round(t, 6), etype, data]) + "\n")

    print(f"    Saved {cast_path}")
    return cast_path


def cast_to_gif(name):
    """Convert asciicast to GIF using agg."""
    cast_path = os.path.join(EXAMPLES_DIR, f"{name}.cast")
    gif_path = os.path.join(EXAMPLES_DIR, f"{name}.gif")

    print(f"  Converting {name} to GIF...")

    agg_bin = os.path.expanduser("~/.cargo/bin/agg")
    ret = os.system(
        f'{agg_bin} --font-size 14 --theme dracula '
        f'"{cast_path}" "{gif_path}" 2>&1'
    )

    if ret == 0 and os.path.exists(gif_path):
        size_kb = os.path.getsize(gif_path) / 1024
        print(f"    Created {gif_path} ({size_kb:.0f} KB)")
    else:
        print(f"    ERROR: Failed to create GIF for {name}")


# ── Example definitions ─────────────────────────────────────────────

EXAMPLES = {
    "example1": {
        "command": "go run cmd/example1_hello/main.go",
        "actions": [
            (3.0, None),    # Display for 3s
            (0, "q"),       # Quit
        ],
    },
    "example2": {
        "command": "go run cmd/example2_counter/main.go",
        "actions": [
            (4.0, None),    # Watch counter for 5s
            (0, "q"),       # Quit
        ],
    },
    "example3": {
        "command": "go run cmd/example3_computed/main.go",
        "actions": [
            (4.0, None),    # Watch for 5s
            (0, "q"),       # Quit
        ],
    },
    "example4": {
        "command": "go run cmd/example4_clock/main.go",
        "actions": [
            (4.0, None),    # Watch clock for 5s
            (0, "q"),       # Quit
        ],
    },
    "example5": {
        "command": "go run cmd/example5_progress/main.go",
        "actions": [
            (6.0, None),    # Watch progress for 6s
            (0, "q"),       # Quit
        ],
    },
    "example6": {
        "command": "go run cmd/example6_conditional/main.go",
        "actions": [
            (6.0, None),    # Watch conditional for 7s
            (0, "q"),       # Quit
        ],
    },
    "example7": {
        "command": "go run cmd/example7_input/main.go",
        "actions": [
            (1.5, None),
            (0.3, RIGHT), (0.3, RIGHT), (0.3, RIGHT),
            (0.5, DOWN), (0.3, DOWN),
            (0.5, LEFT), (0.3, LEFT),
            (0.5, UP),
            (1.0, "q"),
        ],
    },
    "example8": {
        "command": "go run cmd/example8_textinput/main.go",
        "actions": [
            (1.5, None),
        ] + [(0.08, c) for c in "Hello, BasementUI!"] + [
            (1.5, None),
            (0.15, BACKSPACE), (0.15, BACKSPACE), (0.15, BACKSPACE),
            (0.15, BACKSPACE), (0.15, BACKSPACE),
            (0.8, None),
        ] + [(0.08, c) for c in "World!"] + [
            (1.5, None),
            (0, ESC),
        ],
    },
    "example9": {
        "command": "go run cmd/example9_list/main.go",
        "actions": [
            (1.5, None),
            (0.4, DOWN), (0.4, DOWN), (0.4, DOWN),
            (0.5, UP), (0.3, UP),
            (0.5, DOWN), (0.3, DOWN), (0.3, DOWN),
            (0.5, ENTER),
        ],
    },
    "example10": {
        "command": "go run cmd/example10_layout/main.go",
        "rows": LARGE_ROWS,
        "cols": LARGE_COLS,
        "actions": [
            (2.0, None),
            (0.4, DOWN), (0.4, DOWN), (0.4, DOWN),
            (0.5, UP), (0.3, UP), (0.3, UP),
            (1.0, "q"),
        ],
    },
    "example11": {
        "command": "go run cmd/example11_markdown/main.go",
        "rows": LARGE_ROWS,
        "cols": LARGE_COLS,
        "actions": [
            (2.0, None),
        ] + [(0.15, DOWN)] * 10 + [
            (1.0, None),
        ] + [(0.15, UP)] * 10 + [
            (1.0, "q"),
        ],
    },
    "example12": {
        "command": "go run -tags chroma cmd/example12_chroma/main.go",
        "rows": LARGE_ROWS,
        "cols": LARGE_COLS,
        "actions": [
            (3.0, None),
        ] + [(0.2, DOWN)] * 5 + [
            (1.0, None),
        ] + [(0.2, UP)] * 5 + [
            (1.0, "q"),
        ],
    },
}


def main():
    os.makedirs(EXAMPLES_DIR, exist_ok=True)

    # Allow specifying which examples to record
    targets = sys.argv[1:] if len(sys.argv) > 1 else list(EXAMPLES.keys())

    print("Pre-building all example binaries...\n")

    # Build all targeted examples as binaries
    bin_paths = {}
    for name in targets:
        if name not in EXAMPLES:
            print(f"Unknown example: {name}")
            continue
        bp = build_example(name, EXAMPLES[name]["command"])
        if bp:
            bin_paths[name] = bp

    print()

    for name in targets:
        if name not in bin_paths:
            continue

        ex = EXAMPLES[name]
        rows = ex.get("rows", DEFAULT_ROWS)
        
        print(f"\n{'='*50}")
        print(f"Recording {name} (rows={rows})")
        print(f"{'='*50}")

        try:
            record_cast(name, bin_paths[name], ex["actions"], rows=rows)
            cast_to_gif(name)
        except Exception as e:
            print(f"  ERROR recording {name}: {e}")
            import traceback
            traceback.print_exc()

    # Cleanup binaries
    import shutil
    if os.path.exists(BIN_DIR):
        shutil.rmtree(BIN_DIR)

    print(f"\n{'='*50}")
    print("All done!")
    print(f"{'='*50}")


if __name__ == "__main__":
    main()