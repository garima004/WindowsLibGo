package main

import (
    "fmt"
    "strconv"
    "syscall"
    "unsafe"

    "github.com/lxn/win"
)

var (
    user32                   = syscall.NewLazyDLL("user32.dll")
    procMoveWindow           = user32.NewProc("MoveWindow")
    procEnumWindows          = user32.NewProc("EnumWindows")
    procGetWindowTextW       = user32.NewProc("GetWindowTextW")
    procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
    procIsWindowVisible      = user32.NewProc("IsWindowVisible")
)

// WindowEnumProc is a callback function to enumerate windows
type WindowEnumProc func(hwnd win.HWND, lParam uintptr) uintptr

// enumWindows calls the EnumWindows API to enumerate through all top-level windows
func enumWindows(callback WindowEnumProc, lParam uintptr) {
    procEnumWindows.Call(
        syscall.NewCallback(callback),
        lParam,
    )
}

// getWindowTitle gets the title of a window (HWND)
func getWindowTitle(hwnd win.HWND) string {
    length, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
    if length == 0 {
        return ""
    }
    buffer := make([]uint16, length+1)
    procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
    return syscall.UTF16ToString(buffer)
}

// isWindowVisible checks if a window is visible
func isWindowVisible(hwnd win.HWND) bool {
    result, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
    return result != 0
}

// moveWindow moves and resizes the specified window
func moveWindow(hwnd win.HWND, x, y, width, height int32, repaint bool) bool {
    repaintInt := 0
    if repaint {
        repaintInt = 1
    }
    success, _, _ := procMoveWindow.Call(
        uintptr(hwnd),
        uintptr(x),
        uintptr(y),
        uintptr(width),
        uintptr(height),
        uintptr(repaintInt),
    )
    return success != 0
}

func main() {
    var windows []win.HWND
    var windowTitles []string

    // Enumerate all visible windows
    enumWindows(func(hwnd win.HWND, lParam uintptr) uintptr {
        if isWindowVisible(hwnd) {
            title := getWindowTitle(hwnd)
            if title != "" { // Only add windows with non-empty titles
                windows = append(windows, hwnd)
                windowTitles = append(windowTitles, title)
            }
        }
        return 1 // Continue enumeration
    }, 0)

    // Check if any windows were found
    if len(windows) == 0 {
        fmt.Println("No visible windows found.")
        return
    }

    // Display the list of windows
    fmt.Println("Select a window to resize:")
    for i, title := range windowTitles {
        fmt.Printf("%d: %s (HWND: %v)\n", i+1, title, windows[i])
    }

    // Prompt user to select a window by HWND
    var hwndInput string
    fmt.Print("Enter the HWND of the window you want to resize: ")
    _, err := fmt.Scanf("%s", &hwndInput)
    if err != nil {
        fmt.Println("Invalid input.")
        return
    }

    // Convert input to HWND
    hwnd, err := strconv.ParseUint(hwndInput, 10, 32)
    if err != nil {
        fmt.Println("Invalid HWND.")
        return
    }

    // Fixed dimensions and position
    const newWidth = 100
    const newHeight = 300
    const newX = 100
    const newY = 100

    // Resize the selected window
    if moveWindow(win.HWND(hwnd), int32(newX), int32(newY), int32(newWidth), int32(newHeight), true) {
        fmt.Println("Window resized successfully!")
    } else {
        fmt.Println("Failed to resize the window.")
    }
}
