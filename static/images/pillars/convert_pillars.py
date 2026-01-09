#!/usr/bin/env python3
"""
Pillar Image Processor for AionSec Website
Resizes images to 600px and converts to WebP format.

Usage:
    python convert_pillars.py

Input:  Place 1:1 cropped PNG/JPG images in ./input/
Output: Processed WebP images appear in ./output/
"""

from pathlib import Path

try:
    from PIL import Image
except ImportError:
    print("Pillow not installed. Run: pip install Pillow")
    exit(1)

# Configuration
INPUT_DIR = Path(__file__).parent / "input"
OUTPUT_DIR = Path(__file__).parent / "output"
TARGET_SIZE = 600  # pixels (width and height for 1:1 images)
WEBP_QUALITY = 85  # 0-100, 85 is good balance of quality/size

def process_images():
    """Process all images in input directory."""

    # Supported input formats
    supported = {'.png', '.jpg', '.jpeg', '.webp', '.gif'}

    # Find all images
    images = [f for f in INPUT_DIR.iterdir() if f.suffix.lower() in supported]

    if not images:
        print(f"No images found in {INPUT_DIR}")
        print(f"Supported formats: {', '.join(supported)}")
        return

    print(f"Found {len(images)} image(s) to process\n")

    for img_path in images:
        try:
            print(f"Processing: {img_path.name}")

            # Open image
            with Image.open(img_path) as img:
                # Convert to RGB if necessary (for WebP compatibility)
                if img.mode in ('RGBA', 'P'):
                    # Preserve transparency
                    img = img.convert('RGBA')
                elif img.mode != 'RGB':
                    img = img.convert('RGB')

                # Get original dimensions
                orig_w, orig_h = img.size
                print(f"  Original: {orig_w}x{orig_h}")

                # Resize (maintain aspect ratio, fit within TARGET_SIZE)
                img.thumbnail((TARGET_SIZE, TARGET_SIZE), Image.Resampling.LANCZOS)
                new_w, new_h = img.size
                print(f"  Resized:  {new_w}x{new_h}")

                # Output path (change extension to .webp)
                output_path = OUTPUT_DIR / f"{img_path.stem}.webp"

                # Save as WebP
                img.save(output_path, 'WEBP', quality=WEBP_QUALITY)

                # Report file sizes
                orig_size = img_path.stat().st_size / 1024
                new_size = output_path.stat().st_size / 1024
                reduction = (1 - new_size / orig_size) * 100

                print(f"  Size:     {orig_size:.1f}KB -> {new_size:.1f}KB ({reduction:.0f}% smaller)")
                print(f"  Output:   {output_path.name}\n")

        except Exception as e:
            print(f"  ERROR: {e}\n")

    print("Done! WebP images are in ./output/")
    print("\nNext steps:")
    print("1. Copy images from output/ to your desired location")
    print("2. Reference them in your Svelte components")

if __name__ == "__main__":
    process_images()
