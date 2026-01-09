# Pillar Image Processing

Process images for the "Three Pillars" section on the AionSec homepage.

## Quick Start

```bash
# 1. Navigate to this folder
cd /Users/faanross/repos/aionsec-svelte/static/images/pillars

# 2. Install dependency (if needed)
pip install Pillow

# 3. Run the script
python convert_pillars.py
```

## Workflow

### Step 1: Prepare Images
- Crop your 3 pillar images to **1:1 aspect ratio** (square)
- Any resolution is fine - the script will resize
- Supported formats: PNG, JPG, JPEG, WebP, GIF

### Step 2: Add to Input Folder
Place your cropped images in:
```
static/images/pillars/input/
```

Suggested naming:
- `human-centric.png`
- `continuous.png`
- `sovereign.png`

### Step 3: Run Conversion
```bash
python convert_pillars.py
```

### Step 4: Retrieve Output
Processed images appear in:
```
static/images/pillars/output/
```

Each image will be:
- Resized to **600x600px**
- Converted to **WebP format**
- Optimized at **85% quality**

### Step 5: Move to Final Location
Move the WebP files from `output/` to wherever you want them in `static/images/`.

## Configuration

Edit `convert_pillars.py` to change:
- `TARGET_SIZE = 600` - output dimensions in pixels
- `WEBP_QUALITY = 85` - compression quality (0-100)

## Troubleshooting

**"Pillow not installed"**
```bash
pip install Pillow
```

**"No images found"**
- Check images are in the `input/` folder
- Check file extensions are supported (.png, .jpg, etc.)

## Context

These images are for the homepage pillar cards. The cards will:
- Show the image by default
- Reveal text description on hover (desktop) or tap (mobile)

The archaic-cybernetic style images (gold linework on dark background) work well with the AionSec brand.
