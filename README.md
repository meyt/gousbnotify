# gousbnotify

[![Build Status](https://github.com/meyt/gousbnotify/actions/workflows/build.yml/badge.svg)](https://github.com/meyt/gousbnotify/actions)
[![GitHub Downloads (all assets, latest release)](https://img.shields.io/github/downloads/meyt/gousbnotify/latest/total?color=green)](https://github.com/meyt/gousbnotify/releases)
![AI Assisted Badge](https://img.shields.io/badge/AI-ASSISTED-purple)

USB connect/disconnect notifications for Linux.

- Real-time device detection.
- Show pop-up notification.
- Play custom notification sound.
- Portable, just single binary (default sounds embedded).
- Built-in systemd service installer (no root access).

![image](https://github.com/user-attachments/assets/c2cf19c1-9b85-4426-8843-09e89b785dd8)

## Install

```bash
mkdir ~/.local/bin
cd ~/.local/bin
wget -O gousbnotify https://github.com/meyt/gousbnotify/releases/latest/download/gousbnotify-linux-amd64
chmod +x gousbnotify
```

Now you can run it

```bash
gousbnotify
```

## Parameters

All parameters can be used with or without `-install`

- `-install`: Install and activate background service. it will also starts automatically on system startup.
- `-uninstall`: Stop and uninstall the background service.
- `-nosound`: Disables the sound notification
- `-nonotif`: Disables the pop-up notification.
- `-connect-sound`: Sound file path for "Connect".
- `-disconnect-sound`: Sound file path for "Disconnect".

For example:

```bash
gousbnotify -install \
    -connect-sound="/mnt/drivec/Windows/Media/Windows Hardware Insert.wav"
    -disconnect-sound="/mnt/drivec/Windows/Media/Windows Hardware Remove.wav"
```

## Credits

- Based on [`udev`](https://en.wikipedia.org/wiki/Udev),
  [`libnotify`](https://specifications.freedesktop.org/notification-spec/latest/) and
  [`alsa`](https://en.wikipedia.org/wiki/Advanced_Linux_Sound_Architecture)
- Default Sounds: https://cgit.freedesktop.org/sound-theme-freedesktop/
- Assisted AI agents: DeepSeek, Grok, Perplexity, Gemini
