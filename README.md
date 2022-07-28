# pathos - CLI for editing a PATH env variable

## Demo

![pathos DEMO](assets/demo.gif "pathos DEMO")

## Installation
    go get github.com/chip/pathos@latest

## IMPORTANT

Once you have added or removed path enties, this app will save a shell script
to `$HOME/pathos.sh`, which **MUST BE SOURCED** to take effect within your
shell.

    source ~/pathos.sh

## Navigation

Like any typical [Bubble Tea](https://github.com/charmbracelet/bubbletea) app,
you have the following keys that are available for navigation:

Key | Description
---|---
k | Move up the list of directories
j | Move down the list of directories
l | Move to the next page
h | Move to the previous page
g/home | Move to the top of the list
G/end | Move to the end of the list
q | Quit pathos **(does NOT save changes)**
? | Toggle between regular and full help

## Commands

Key | Description
---|---
N | Add a new directory to the list at the current cursor position
D | Delete a directory at the current cursor position
S | Save all changes made to the list of directories

## Color Highlighting

Color | Description
---|---
<span style="background-color:black"> &nbsp; <span style="color:yellow">Yellow</span> &nbsp; </span> | Shows current cursor position</span>
<span style="background-color:black"> &nbsp; <span style="color:red">Red</span> &nbsp; </span> | Indicates directories that **do not exist**
<span style="background-color:black"> &nbsp; <span style="color:aqua">Aqua</span> &nbsp; </span> | Indicates duplicate directories
