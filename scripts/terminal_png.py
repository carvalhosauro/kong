#!/usr/bin/env python3
"""Render terminal output as a PNG for README screenshots."""

from __future__ import annotations

import argparse
import textwrap
from pathlib import Path

try:
    from PIL import Image, ImageDraw, ImageFont
except ImportError as exc:
    raise SystemExit(
        "Pillow required: pip install Pillow\n"
        "Or: dnf install python3-pillow / apt install python3-pillow"
    ) from exc


def load_font(size: int, mono: bool = True) -> ImageFont.FreeTypeFont | ImageFont.ImageFont:
    candidates = []
    if mono:
        candidates += [
            "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
            "/usr/share/fonts/dejavu/DejaVuSansMono.ttf",
            "/usr/share/fonts/google-droid/DroidSansMono.ttf",
            "/usr/share/fonts/liberation/LiberationMono-Regular.ttf",
        ]
    candidates += [
        "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
        "/usr/share/fonts/dejavu/DejaVuSans.ttf",
    ]
    for path in candidates:
        if Path(path).exists():
            return ImageFont.truetype(path, size)
    return ImageFont.load_default()


def wrap_lines(text: str, width: int) -> list[str]:
    lines: list[str] = []
    for raw in text.splitlines():
        if len(raw) <= width:
            lines.append(raw)
        else:
            lines.extend(
                textwrap.wrap(
                    raw,
                    width=width,
                    break_long_words=False,
                    replace_whitespace=False,
                )
                or [raw]
            )
    return lines


def render(
    body: str,
    title: str,
    prompt: str,
    command: str | None,
    width_px: int,
    font_size: int,
    min_height: int,
) -> Image.Image:
    font = load_font(font_size)
    title_font = load_font(font_size, mono=False)

    char_w = font.getlength("M")
    cols = max(40, int((width_px - 48) / char_w))
    lines = wrap_lines(body.rstrip("\n"), cols)

    if command:
        header = f"{prompt} {command}"
        lines = [header, ""] + lines

    line_h = font_size + 6
    pad = 20
    title_h = 36
    height = max(min_height, title_h + pad * 2 + line_h * len(lines))

    img = Image.new("RGB", (width_px, height), "#0d1117")
    draw = ImageDraw.Draw(img)

    # title bar
    draw.rectangle((0, 0, width_px, title_h), fill="#161b22")
    for i, color in enumerate(["#ff5f57", "#febc2e", "#28c840"]):
        draw.ellipse((14 + i * 22, 12, 26 + i * 22, 24), fill=color)
    draw.text((width_px // 2 - draw.textlength(title, font=title_font) // 2, 10), title, fill="#c9d1d9", font=title_font)

    y = title_h + pad
    colors = {
        "warning:": "#d29922",
        "ignored_by_cron": "#d29922",
        "cron_d": "#79c0ff",
        "user_crontab": "#a5d6ff",
        "etc_crontab": "#7ee787",
        "run_parts": "#d2a8ff",
        "anacron": "#ffa657",
    }

    for line in lines:
        color = "#e6edf3"
        for key, c in colors.items():
            if key in line:
                color = c
                break
        if line.startswith(prompt) or line.startswith("$"):
            color = "#7ee787"
        draw.text((pad, y), line, fill=color, font=font)
        y += line_h

    return img


def main() -> None:
    p = argparse.ArgumentParser()
    p.add_argument("--input", required=True)
    p.add_argument("--output", required=True)
    p.add_argument("--title", default="terminal")
    p.add_argument("--prompt", default="$")
    p.add_argument("--command", default=None)
    p.add_argument("--width", type=int, default=1280)
    p.add_argument("--font-size", type=int, default=14)
    p.add_argument("--height", type=int, default=0)
    args = p.parse_args()

    body = Path(args.input).read_text(encoding="utf-8")
    img = render(
        body=body,
        title=args.title,
        prompt=args.prompt,
        command=args.command,
        width_px=args.width,
        font_size=args.font_size,
        min_height=args.height,
    )
    Path(args.output).parent.mkdir(parents=True, exist_ok=True)
    img.save(args.output, optimize=True)
    print(args.output)


if __name__ == "__main__":
    main()
