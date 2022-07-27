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
