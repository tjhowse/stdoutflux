# stdoutflux

Connects to a miniflux instance, pulls unread entries, formats them for printing on a limited-line display (32 characters by default) and outputs to stdout.

It's quite dumb. Much of it was written by copilot.

I use it like this:

    stdoutflux | pv --rate-limit 32 | receipt

Where `pv` is [PipeViewer](https://www.ivarch.com/programs/pv.shtml) and `receipt` is a bash script that sends each line to my MQTT broker to be printed by a thermal receipt printer running esphome.

It does this:

![A photo of a long ribbon of receipt paper printing from a receipt printer. Slashdot articles are printed on the paper.](images/demo.jpg)

You can set the miniflux URL, API token, line length, and whether printed articles should be marked read (default false) with command-line arguments or environment variables. CLI args take precedence.

    MINIFLUX_URL
    MINIFLUX_API_TOKEN
    STDOUTFLUX_LINE_LENGTH
    STDOUTFLUX_MARK_READ

Theoretically you could run this in a cronjob, or something.

## Installation

I'm not sure why you'd want to do this, but...

    go install github.com/tjhowse/stdoutflux

Optionally, to install PipeViewer:

    sudo apt update
    sudo apt install -y pv

## Known issues

My thermal printer doesn't use a monospaced font, so sometimes it wraps lines of its own accord, splitting words annoyingly. The default of 32 characters is the maximum characters that can fit on a line.