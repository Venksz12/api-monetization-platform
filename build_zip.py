from pathlib import Path
import zipfile

ROOT = Path(__file__).resolve().parent
OUT = ROOT.parent / "api-monetization-platform.zip"

with zipfile.ZipFile(OUT, "w", zipfile.ZIP_DEFLATED) as z:
    for path in ROOT.rglob("*"):
        if path.is_file() and path.name != OUT.name:
            z.write(path, path.relative_to(ROOT.parent))

print(OUT)
